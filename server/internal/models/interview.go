package models

import "time"

// Interview 面试会话主表
type Interview struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	UserID            uint       `gorm:"index;not null" json:"user_id"`
	Scene             string     `gorm:"size:20;not null" json:"scene"` // tech/behavior/pressure/hr/group/teaching
	TargetCompany     string     `gorm:"size:100" json:"target_company"`
	TargetPosition    string     `gorm:"size:100" json:"target_position"`
	TargetJD          string     `gorm:"type:text" json:"target_jd"`
	ResumeID          uint       `gorm:"index;default:0" json:"resume_id"`
	ResumeVersionID   uint       `gorm:"index;default:0" json:"resume_version_id"`
	ResumeSnapshot    string     `gorm:"type:text" json:"resume_snapshot"` // 面试开始前固化的简历快照（版本内容 JSON）
	ResumeName        string     `gorm:"size:100" json:"resume_name"`      // 简历名称（用于前端展示）
	ExaminerStyle     string     `gorm:"size:30" json:"examiner_style"`
	TrainingFocus     string     `gorm:"size:500" json:"training_focus"`
	Difficulty        string     `gorm:"size:20;default:mid" json:"difficulty"` // junior/mid/senior/mixed
	TotalQuestions    int        `gorm:"default:8" json:"total_questions"`
	Mode              string     `gorm:"size:20;default:hybrid" json:"mode"`            // text/voice/hybrid
	Status            string     `gorm:"size:20;default:preparing;index" json:"status"` // preparing/starting/ongoing/reviewing/completed/report_failed/cancelled
	StatusMessage     string     `gorm:"size:500" json:"status_message"`
	CurrentQuestionNo int        `gorm:"default:0" json:"current_question_no"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Interview) TableName() string { return "interviews" }

// InterviewMessage 面试问答记录
type InterviewMessage struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	InterviewID  uint      `gorm:"index;not null" json:"interview_id"`
	Role         string    `gorm:"size:10;not null" json:"role"` // assistant / user
	Content      string    `gorm:"type:text" json:"content"`
	AudioURL     string    `gorm:"size:500" json:"audio_url"`    // user=录音，assistant=TTS
	InputMode    string    `gorm:"size:20" json:"input_mode"`    // text/voice（用户回答实际使用的通道）
	QuestionType string    `gorm:"size:20" json:"question_type"` // project/algorithm/system_design/behavior/open/followup
	QuestionNo   int       `gorm:"default:0" json:"question_no"`
	DurationSec  int       `json:"duration_sec"`
	CreatedAt    time.Time `json:"created_at"`
}

func (InterviewMessage) TableName() string { return "interview_messages" }

// InterviewScore 面试五维度评分
type InterviewScore struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	InterviewID uint      `gorm:"index;not null" json:"interview_id"`
	Dimension   string    `gorm:"size:20;not null" json:"dimension"` // professional/expression/logic/adaptability/pace
	Score       int       `json:"score"`                             // 0-100
	Comment     string    `gorm:"type:text" json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

func (InterviewScore) TableName() string { return "interview_scores" }

// InterviewReport 面试复盘报告
type InterviewReport struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	InterviewID      uint      `gorm:"uniqueIndex;not null" json:"interview_id"`
	OverallScore     int       `json:"overall_score"` // 0-100
	Summary          string    `gorm:"type:text" json:"summary"`
	Highlights       string    `gorm:"type:text" json:"highlights"`        // JSON 数组
	Improvements     string    `gorm:"type:text" json:"improvements"`      // JSON 数组
	Recommendations  string    `gorm:"type:text" json:"recommendations"`   // JSON 数组
	WordCloud        string    `gorm:"type:text" json:"word_cloud"`        // JSON 对象
	QuestionFeedback string    `gorm:"type:text" json:"question_feedback"` // JSON 数组，每题的逐题评价
	CreatedAt        time.Time `json:"created_at"`
}

func (InterviewReport) TableName() string { return "interview_reports" }
