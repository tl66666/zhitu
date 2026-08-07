package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

func TestCreateSupportsAllSceneHallScenesWithoutLLM(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:scene_hall_test?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Interview{}, &models.InterviewMessage{}, &models.Resume{}, &models.ResumeVersion{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	resume := models.Resume{UserID: 1, Name: "测试简历", TargetPosition: "后端工程师"}
	if err := db.Create(&resume).Error; err != nil {
		t.Fatalf("create resume: %v", err)
	}
	version := models.ResumeVersion{ResumeID: resume.ID, VersionLabel: "v1.0", Content: `{"personal":{"name":"候选人"},"project":[{"name":"支付平台"}]}`}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("create resume version: %v", err)
	}
	if err := db.Model(&resume).Update("current_version_id", version.ID).Error; err != nil {
		t.Fatalf("set current version: %v", err)
	}

	service := NewInterviewService(db, nil, nil, nil)
	scenes := []string{
		SceneTeaching,
		SceneCorporate,
		SceneGroup,
		SceneDefense,
		SceneClient,
		ScenePressure,
		ScenePublic,
		SceneMedical,
		SceneMedia,
		SceneRemote,
		SceneSystem,
		SceneAviation,
	}

	for _, scene := range scenes {
		t.Run(scene, func(t *testing.T) {
			interview, err := service.Create(context.Background(), 1, &CreateInterviewInput{
				Scene:          scene,
				TargetPosition: "测试岗位",
				TargetJD:       "负责支付系统开发，要求熟悉 Go 和分布式系统。",
				ResumeID:       resume.ID,
				TotalQuestions: 5,
				Mode:           ModeHybrid,
			})
			if err != nil {
				t.Fatalf("create scene %q: %v", scene, err)
			}
			if interview.Status != StatusPreparing {
				t.Fatalf("status = %q, want %q", interview.Status, StatusPreparing)
			}
			if interview.CurrentQuestionNo != 0 {
				t.Fatalf("current question = %d, want 0", interview.CurrentQuestionNo)
			}
			var count int64
			if err := db.Model(&models.InterviewMessage{}).Where("interview_id = ?", interview.ID).Count(&count).Error; err != nil {
				t.Fatalf("count messages: %v", err)
			}
			if count != 0 {
				t.Fatalf("message count = %d, want 0 before start", count)
			}
		})
	}
}

func TestStartGeneratesResumeAndJDGroundedFirstQuestion(t *testing.T) {
	requests := make(chan map[string]interface{}, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests <- payload
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"请结合支付平台项目说明你如何处理分布式一致性？"}}]}`))
	}))
	defer upstream.Close()

	db, err := gorm.Open(sqlite.Open("file:start_grounded_test?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Interview{}, &models.InterviewMessage{}, &models.Resume{}, &models.ResumeVersion{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	resume := models.Resume{UserID: 1, Name: "后端简历"}
	if err := db.Create(&resume).Error; err != nil {
		t.Fatalf("create resume: %v", err)
	}
	content := `{"personal":{"name":"候选人"},"project":[{"name":"支付平台","role":"负责人","description":"负责订单一致性和高并发架构"}]}`
	version := models.ResumeVersion{ResumeID: resume.ID, VersionLabel: "v1.0", Content: content}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("create resume version: %v", err)
	}
	if err := db.Model(&resume).Update("current_version_id", version.ID).Error; err != nil {
		t.Fatalf("set current version: %v", err)
	}

	service := NewInterviewService(db, NewLLMService(&config.LLMConfig{
		Provider: "mimo", BaseURL: upstream.URL, APIKey: "test-key", ChatModel: "test-chat",
	}), nil, nil)
	interview, err := service.Create(context.Background(), 1, &CreateInterviewInput{
		Scene: "tech", TargetPosition: "后端工程师", TargetJD: "负责支付系统开发，要求熟悉 Go 和分布式系统。", ResumeID: resume.ID,
		TotalQuestions: 5, Mode: ModeVoice,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	started, first, err := service.Start(context.Background(), 1, interview.ID, nil)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if started.Status != StatusOngoing || started.CurrentQuestionNo != 1 {
		t.Fatalf("started interview = status %q, question %d", started.Status, started.CurrentQuestionNo)
	}
	if first == nil || first.Content == "" {
		t.Fatal("Start() returned an empty first question")
	}

	payload := <-requests
	rawMessages, ok := payload["messages"].([]interface{})
	if !ok || len(rawMessages) == 0 {
		t.Fatalf("messages = %#v", payload["messages"])
	}
	system := rawMessages[0].(map[string]interface{})["content"].(string)
	for _, expected := range []string{"负责支付系统开发", "支付平台", "订单一致性"} {
		if !strings.Contains(system, expected) {
			t.Fatalf("system prompt missing %q:\n%s", expected, system)
		}
	}
}

func TestAttachResumeUsesOwnedCurrentVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:attach_resume_test?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Interview{},
		&models.Resume{},
		&models.ResumeVersion{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	interview := models.Interview{UserID: 1, Status: StatusOngoing}
	if err := db.Create(&interview).Error; err != nil {
		t.Fatalf("create interview: %v", err)
	}
	resume := models.Resume{UserID: 1, Name: "后端工程师简历", Scene: "manual"}
	if err := db.Create(&resume).Error; err != nil {
		t.Fatalf("create resume: %v", err)
	}
	content := `{"personal":{"name":"候选人"},"project":[{"name":"搜索平台","role":"负责人","description":"负责架构设计"}]}`
	version := models.ResumeVersion{
		ResumeID:     resume.ID,
		VersionLabel: "v1.0",
		Content:      content,
	}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("create resume version: %v", err)
	}
	if err := db.Model(&resume).Update("current_version_id", version.ID).Error; err != nil {
		t.Fatalf("set current version: %v", err)
	}

	service := NewInterviewService(db, nil, nil, nil)
	got, err := service.AttachResume(1, interview.ID, &AttachResumeInput{ResumeID: resume.ID})
	if err != nil {
		t.Fatalf("AttachResume() error = %v", err)
	}
	if got.ResumeSnapshot != content || got.ResumeName != resume.Name {
		t.Fatalf("attached resume = %#v", got)
	}

	var persisted models.Interview
	if err := db.First(&persisted, interview.ID).Error; err != nil {
		t.Fatalf("reload interview: %v", err)
	}
	if persisted.ResumeSnapshot != content || persisted.ResumeName != resume.Name {
		t.Fatalf("persisted resume snapshot/name = %q/%q", persisted.ResumeSnapshot, persisted.ResumeName)
	}

	if _, err := service.AttachResume(2, interview.ID, &AttachResumeInput{ResumeID: resume.ID}); !errors.Is(err, ErrInterviewNotFound) {
		t.Fatalf("other user AttachResume() error = %v, want ErrInterviewNotFound", err)
	}
}

func TestSummarizeResumeTruncatesPromptContent(t *testing.T) {
	content := `{"custom":[{"title":"补充经历","content":"` +
		strings.Repeat("项", maxResumePromptRunes+100) +
		`"}]}`
	summary := summarizeResume(content)
	if !strings.HasSuffix(summary, "（简历内容已截断）") {
		t.Fatalf("summary was not truncated")
	}
	if got := len([]rune(summary)); got > maxResumePromptRunes+20 {
		t.Fatalf("summary rune count = %d", got)
	}
}

func TestTranscribeVoiceDoesNotAdvanceInterview(t *testing.T) {
	requests := make(chan map[string]interface{}, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests <- payload
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"这是转写后的回答"}}]}`))
	}))
	defer upstream.Close()

	db, err := gorm.Open(sqlite.Open("file:transcribe_voice_test?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Interview{}, &models.InterviewMessage{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	interview := models.Interview{
		UserID:            1,
		Status:            StatusOngoing,
		CurrentQuestionNo: 2,
		TotalQuestions:    5,
	}
	if err := db.Create(&interview).Error; err != nil {
		t.Fatalf("create interview: %v", err)
	}

	llm := NewLLMService(&config.LLMConfig{
		Provider:     "mimo",
		BaseURL:      upstream.URL,
		APIKey:       "test-key",
		WhisperModel: "mimo-v2.5-asr",
	})
	service := NewInterviewService(db, llm, nil, nil)
	text, err := service.TranscribeVoice(
		context.Background(),
		1,
		interview.ID,
		bytes.NewReader([]byte("RIFF-test-wav")),
		"answer.wav",
	)
	if err != nil {
		t.Fatalf("TranscribeVoice() error = %v", err)
	}
	if text != "这是转写后的回答" {
		t.Fatalf("TranscribeVoice() = %q", text)
	}

	payload := <-requests
	if payload["model"] != "mimo-v2.5-asr" {
		t.Fatalf("model = %#v", payload["model"])
	}
	messages, ok := payload["messages"].([]interface{})
	if !ok || len(messages) != 1 {
		t.Fatalf("messages = %#v", payload["messages"])
	}
	messagePayload := messages[0].(map[string]interface{})
	content := messagePayload["content"].([]interface{})
	audioItem := content[0].(map[string]interface{})
	inputAudio := audioItem["input_audio"].(map[string]interface{})
	if data, _ := inputAudio["data"].(string); !strings.HasPrefix(data, "data:audio/wav;base64,") {
		t.Fatalf("audio data url = %q", data)
	}

	var count int64
	if err := db.Model(&models.InterviewMessage{}).Where("interview_id = ?", interview.ID).Count(&count).Error; err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 0 {
		t.Fatalf("message count = %d, want 0", count)
	}
	var persisted models.Interview
	if err := db.First(&persisted, interview.ID).Error; err != nil {
		t.Fatalf("reload interview: %v", err)
	}
	if persisted.CurrentQuestionNo != 2 {
		t.Fatalf("current question = %d, want 2", persisted.CurrentQuestionNo)
	}
}

func TestStorageURLWithAbsoluteBase(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "uploads")
	audioPath := filepath.Join(baseDir, "audio", "answer.wav")
	got, err := storageURL(baseDir, audioPath)
	if err != nil {
		t.Fatalf("storageURL() error = %v", err)
	}
	if got != "/static/audio/answer.wav" {
		t.Fatalf("storageURL() = %q", got)
	}

	if _, err := storageURL(baseDir, filepath.Join(filepath.Dir(baseDir), "outside.wav")); err == nil {
		t.Fatal("storageURL() accepted path outside storage base")
	}
}

func TestBuildNextQuestionInstructionUsesLatestAnswer(t *testing.T) {
	history := []models.InterviewMessage{
		{Role: "assistant", Content: "请介绍一个你负责的项目。"},
		{Role: "user", Content: "我负责过旧项目。"},
		{Role: "assistant", Content: "你为什么在新系统中选择 Redis？"},
		{Role: "user", Content: "因为缓存命中率达到 92%，但当时没有处理热 key。"},
	}

	instruction := buildNextQuestionInstruction(history, 3)
	for _, expected := range []string{
		"你为什么在新系统中选择 Redis",
		"缓存命中率达到 92%",
		"没有处理热 key",
		"不得无视回答机械切换到通用题库",
		"所有逐题评价统一留到面试结束后的总体报告",
	} {
		if !strings.Contains(instruction, expected) {
			t.Fatalf("instruction missing %q:\n%s", expected, instruction)
		}
	}
	if strings.Contains(instruction, "我负责过旧项目") {
		t.Fatalf("instruction included an older answer:\n%s", instruction)
	}
}

func TestBuildNextQuestionInstructionLimitsAnswerLength(t *testing.T) {
	longAnswer := strings.Repeat("答", maxFollowupPromptRunes+100)
	instruction := buildNextQuestionInstruction([]models.InterviewMessage{
		{Role: "assistant", Content: "上一题"},
		{Role: "user", Content: longAnswer},
	}, 2)
	if !strings.Contains(instruction, "…") {
		t.Fatalf("long answer was not truncated")
	}
	if strings.Contains(instruction, strings.Repeat("答", maxFollowupPromptRunes+1)) {
		t.Fatalf("instruction kept more than the configured answer limit")
	}
}
