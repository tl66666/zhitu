package services

import (
	"context"
	"errors"
	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
	"strings"
	"time"
)

// 面试相关错误
var (
	ErrInterviewNotFound  = errors.New("interview not found")
	ErrInterviewEnded     = errors.New("interview already ended")
	ErrInterviewPreparing = errors.New("interview is still preparing")
	ErrInterviewStarting  = errors.New("interview is already starting")
	ErrReportGenerating   = errors.New("interview report is being generated")
	ErrResumeRequired     = errors.New("resume is required before starting the interview")
	ErrTargetJDRequired   = errors.New("target_jd is required")
	ErrResumeLocked       = errors.New("resume is locked for this interview")
	ErrInvalidMode        = errors.New("invalid interview mode")
	ErrModeNotAllowed     = errors.New("this response mode is not allowed for the interview")
	ErrMessageNotFound    = errors.New("interview message not found")
	ErrReportNotReady     = errors.New("report not generated yet, please end the interview first")
)

// 面试场景枚举
const (
	SceneTech      = "tech"
	SceneBehavior  = "behavior"
	ScenePressure  = "pressure"
	SceneHR        = "hr"
	SceneGroup     = "group"
	SceneTeaching  = "teaching"
	SceneCorporate = "corporate"
	SceneDefense   = "defense"
	SceneClient    = "client"
	ScenePublic    = "public"
	SceneMedical   = "medical"
	SceneMedia     = "media"
	SceneRemote    = "remote"
	SceneSystem    = "system"
	SceneAviation  = "aviation"
)

var validInterviewScenes = map[string]struct{}{
	SceneTech: {}, SceneBehavior: {}, ScenePressure: {}, SceneHR: {},
	SceneGroup: {}, SceneTeaching: {}, SceneCorporate: {}, SceneDefense: {},
	SceneClient: {}, ScenePublic: {}, SceneMedical: {}, SceneMedia: {},
	SceneRemote: {}, SceneSystem: {}, SceneAviation: {},
}

// 面试状态枚举
const (
	StatusOngoing      = "ongoing"
	StatusPreparing    = "preparing"
	StatusStarting     = "starting"
	StatusReviewing    = "reviewing"
	StatusCompleted    = "completed"
	StatusReportFailed = "report_failed"
	StatusCancelled    = "cancelled"
)

// 面试模式枚举
const (
	ModeText   = "text"
	ModeVoice  = "voice"
	ModeHybrid = "hybrid"
)

var validInterviewModes = map[string]struct{}{ModeText: {}, ModeVoice: {}, ModeHybrid: {}}

const (
	maxResumePromptRunes   = 12000
	maxFollowupPromptRunes = 2000
)

// InterviewService 面试业务逻辑
type InterviewService struct {
	db      *gorm.DB
	llm     *LLMService
	profile *ProfileService
	storage *config.StorageConfig
}

// NewInterviewService 构造 InterviewService
func NewInterviewService(db *gorm.DB, llm *LLMService, profile *ProfileService, storage *config.StorageConfig) *InterviewService {
	return &InterviewService{db: db, llm: llm, profile: profile, storage: storage}
}

// CreateInterviewInput 创建面试入参
type CreateInterviewInput struct {
	Scene          string `json:"scene" binding:"required"`
	TargetCompany  string `json:"target_company"`
	TargetPosition string `json:"target_position" binding:"required"`
	TargetJD       string `json:"target_jd" binding:"required"`
	ResumeID       uint   `json:"resume_id" binding:"required"`
	VersionID      uint   `json:"version_id"`
	Difficulty     string `json:"difficulty"`
	TotalQuestions int    `json:"total_questions"`
	Mode           string `json:"mode"`
	ExaminerStyle  string `json:"examiner_style"`
	TrainingFocus  string `json:"training_focus"`
}

// AttachResumeInput 面试发送简历入参
type AttachResumeInput struct {
	ResumeID  uint `json:"resume_id" binding:"required"`
	VersionID uint `json:"version_id"` // 可选，不传则用简历当前版本
}

// Create 创建面试准备会话。首题由 Start 在简历和 JD 已固化后生成。
func (s *InterviewService) Create(ctx context.Context, userID uint, in *CreateInterviewInput) (*models.Interview, error) {
	if _, ok := validInterviewScenes[in.Scene]; !ok {
		return nil, errors.New("invalid interview scene")
	}
	if in.Difficulty == "" {
		in.Difficulty = "mid"
	}
	if in.TotalQuestions == 0 {
		in.TotalQuestions = 8
	}
	if in.TotalQuestions < 5 || in.TotalQuestions > 15 {
		return nil, errors.New("total_questions must be between 5 and 15")
	}
	if in.Mode == "" {
		in.Mode = ModeHybrid
	}
	if _, ok := validInterviewModes[in.Mode]; !ok {
		return nil, ErrInvalidMode
	}
	if strings.TrimSpace(in.TargetJD) == "" {
		return nil, ErrTargetJDRequired
	}
	if in.ResumeID == 0 {
		return nil, ErrResumeRequired
	}
	resume, version, err := s.loadResumeSnapshot(userID, in.ResumeID, in.VersionID)
	if err != nil {
		return nil, err
	}

	interview := &models.Interview{
		UserID:            userID,
		Scene:             in.Scene,
		TargetCompany:     in.TargetCompany,
		TargetPosition:    in.TargetPosition,
		TargetJD:          strings.TrimSpace(in.TargetJD),
		ResumeID:          resume.ID,
		ResumeVersionID:   version.ID,
		ResumeSnapshot:    version.Content,
		ResumeName:        resume.Name,
		Difficulty:        in.Difficulty,
		TotalQuestions:    in.TotalQuestions,
		Mode:              in.Mode,
		ExaminerStyle:     strings.TrimSpace(in.ExaminerStyle),
		TrainingFocus:     strings.TrimSpace(in.TrainingFocus),
		Status:            StatusPreparing,
		CurrentQuestionNo: 0,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(interview).Error
	}); err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *InterviewService) loadResumeSnapshot(userID, resumeID, versionID uint) (*models.Resume, *models.ResumeVersion, error) {
	var resume models.Resume
	if err := s.db.Where("id = ? AND user_id = ?", resumeID, userID).First(&resume).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrResumeNotFound
	} else if err != nil {
		return nil, nil, err
	}
	if versionID == 0 {
		versionID = resume.CurrentVersionID
	}
	if versionID == 0 {
		return nil, nil, ErrVersionNotFound
	}
	var version models.ResumeVersion
	if err := s.db.Where("id = ? AND resume_id = ?", versionID, resumeID).First(&version).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrVersionNotFound
	} else if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(version.Content) == "" {
		return nil, nil, errors.New("resume version content is empty")
	}
	return &resume, &version, nil
}

// SetMode updates the active interaction mode without resetting the conversation.
func (s *InterviewService) SetMode(userID, interviewID uint, mode string) (*models.Interview, error) {
	if _, ok := validInterviewModes[mode]; !ok {
		return nil, ErrInvalidMode
	}
	interview, err := s.Get(userID, interviewID)
	if err != nil {
		return nil, err
	}
	if interview.Status != StatusPreparing && interview.Status != StatusOngoing {
		return nil, ErrInterviewEnded
	}
	if err := s.db.Model(interview).Update("mode", mode).Error; err != nil {
		return nil, err
	}
	interview.Mode = mode
	return interview, nil
}

// Cancel cancels a session that has not entered the formal interview.
func (s *InterviewService) Cancel(userID, interviewID uint) (*models.Interview, error) {
	interview, err := s.Get(userID, interviewID)
	if err != nil {
		return nil, err
	}
	if interview.Status == StatusCancelled {
		return interview, nil
	}
	if interview.Status != StatusPreparing {
		return nil, ErrInterviewEnded
	}
	now := time.Now()
	result := s.db.Model(&models.Interview{}).
		Where("id = ? AND user_id = ? AND status = ?", interviewID, userID, StatusPreparing).
		Updates(map[string]interface{}{
			"status": StatusCancelled, "status_message": "", "ended_at": now,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrInterviewEnded
	}
	interview.Status = StatusCancelled
	interview.StatusMessage = ""
	interview.EndedAt = &now
	return interview, nil
}

// Get 获取面试详情（含所有消息）
func (s *InterviewService) Get(userID, id uint) (*models.Interview, error) {
	var interview models.Interview
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&interview).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInterviewNotFound
	}
	return &interview, err
}

// List 列出用户的所有面试
func (s *InterviewService) List(userID uint) ([]models.Interview, error) {
	var list []models.Interview
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListMessages 列出面试的所有消息
func (s *InterviewService) ListMessages(userID, interviewID uint) ([]models.InterviewMessage, error) {
	if _, err := s.Get(userID, interviewID); err != nil {
		return nil, err
	}
	var list []models.InterviewMessage
	err := s.db.Where("interview_id = ?", interviewID).Order("id ASC").Find(&list).Error
	return list, err
}
