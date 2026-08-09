package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

func TestResumeCopilotChatRetriesInvalidJSON(t *testing.T) {
	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := "not-json"
		if requests.Add(1) == 2 {
			content = `{"reply":"重试成功","match":null,"project":null,"prediction":null,"memory_summary":""}`
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []interface{}{map[string]interface{}{
				"message": map[string]string{"role": "assistant", "content": content},
			}},
		})
	}))
	defer upstream.Close()

	llm := NewLLMService(&config.LLMConfig{
		Provider: "openai", BaseURL: upstream.URL, APIKey: "test-key",
		ChatModel: "test-model", MaxTokens: 256, TimeoutSec: 5,
	})
	service := NewResumeCopilotService(llm, nil, nil)
	result, err := service.Chat(context.Background(), 7, &CopilotInput{
		Task: CopilotTaskCareerChat, DraftContent: "候选人有产品与研发经历",
		Messages: []CopilotMessage{{Role: "user", Content: "请概括方向"}},
	})
	if err != nil {
		t.Fatalf("chat after retry: %v", err)
	}
	if result.Reply != "重试成功" {
		t.Fatalf("reply = %q", result.Reply)
	}
	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want 2", requests.Load())
	}
}

func TestCompleteCopilotLLMResponseRequiresTaskResult(t *testing.T) {
	base := &copilotLLMResponse{Reply: "已完成"}
	if completeCopilotLLMResponse(CopilotTaskProjectOptimize, base) {
		t.Fatal("project task accepted a missing project result")
	}
	base.Project = &CopilotProjectResult{CurrentIssues: []string{"缺少量化结果"}}
	if !completeCopilotLLMResponse(CopilotTaskProjectOptimize, base) {
		t.Fatal("project task rejected a complete project result")
	}
	if !completeCopilotLLMResponse(CopilotTaskCareerChat, &copilotLLMResponse{Reply: "回答"}) {
		t.Fatal("career chat rejected a textual reply")
	}
}

func TestCopilotTaskOutputSchemaIncludesProjectShape(t *testing.T) {
	schema := copilotTaskOutputSchema(CopilotTaskProjectOptimize)
	for _, field := range []string{"current_issues", "star_analysis", "rewritten_description", "rewritten_tech_stack"} {
		if !strings.Contains(schema, field) {
			t.Fatalf("project schema missing %q: %s", field, schema)
		}
	}
}

func TestEmitCopilotReplyChunksLargeDelta(t *testing.T) {
	value := strings.Repeat("流式回复。", 12)
	var chunks []string
	emitCopilotReply(context.Background(), value, func(delta string) {
		chunks = append(chunks, delta)
	})
	if len(chunks) < 2 {
		t.Fatalf("chunks = %d, want multiple chunks", len(chunks))
	}
	var rebuilt strings.Builder
	for _, chunk := range chunks {
		if got := len([]rune(chunk)); got > copilotReplyChunkRunes {
			t.Fatalf("chunk rune length = %d, want <= %d", got, copilotReplyChunkRunes)
		}
		rebuilt.WriteString(chunk)
	}
	if rebuilt.String() != value {
		t.Fatalf("rebuilt reply = %q, want %q", rebuilt.String(), value)
	}
}

func newCopilotTestService(t *testing.T) (*ResumeCopilotService, *gorm.DB, *models.Resume, *models.ResumeVersion) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.Resume{}, &models.ResumeVersion{}, &models.ResumeAIOperation{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	resume := &models.Resume{UserID: 7, Name: "后端简历", TargetPosition: "后端工程师", TargetJD: "熟悉 Go"}
	if err := db.Create(resume).Error; err != nil {
		t.Fatalf("create resume: %v", err)
	}
	version := &models.ResumeVersion{
		ResumeID: resume.ID, VersionLabel: "v1.0",
		Content: `{"project":[{"name":"支付平台","description":"负责接口开发","tech_stack":["Go"]}]}`,
	}
	if err := db.Create(version).Error; err != nil {
		t.Fatalf("create version: %v", err)
	}
	resume.CurrentVersionID = version.ID
	if err := db.Model(resume).Update("current_version_id", version.ID).Error; err != nil {
		t.Fatalf("set current version: %v", err)
	}
	return NewResumeCopilotService(nil, NewResumeService(db, nil), nil), db, resume, version
}

func TestResumeCopilotChatValidatesTaskAndContext(t *testing.T) {
	service, _, resume, _ := newCopilotTestService(t)

	if _, err := service.Chat(context.Background(), 7, &CopilotInput{Task: "unknown", ResumeID: resume.ID}); !errors.Is(err, ErrCopilotInvalidTask) {
		t.Fatalf("invalid task error = %v", err)
	}
	if _, err := service.Chat(context.Background(), 7, &CopilotInput{
		Task: CopilotTaskJDMatch, DraftContent: "候选人有 Go 项目",
	}); !errors.Is(err, ErrCopilotJDRequired) {
		t.Fatalf("missing JD error = %v", err)
	}
	if _, err := service.Chat(context.Background(), 7, &CopilotInput{
		Task: CopilotTaskCareerChat, DraftContent: strings.Repeat("简历", maxCopilotContentRunes+1),
	}); !errors.Is(err, ErrCopilotContentTooLong) {
		t.Fatalf("oversized draft error = %v", err)
	}
	if _, err := service.Chat(context.Background(), 8, &CopilotInput{Task: CopilotTaskCareerChat, ResumeID: resume.ID}); !errors.Is(err, ErrResumeNotFound) {
		t.Fatalf("cross-user resume error = %v", err)
	}
	if _, err := service.Chat(context.Background(), 7, &CopilotInput{
		Task: CopilotTaskProjectOptimize, DraftContent: `{"project":[{"name":"支付平台"}]}`, ProjectIndex: intPtr(1),
	}); !errors.Is(err, ErrCopilotProjectRange) {
		t.Fatalf("out-of-range project error = %v", err)
	}
	if _, err := service.Chat(context.Background(), 7, &CopilotInput{
		Task: CopilotTaskProjectOptimize, DraftContent: `{"project":[]}`, ProjectIndex: intPtr(0),
	}); !errors.Is(err, ErrCopilotProjectEmpty) {
		t.Fatalf("empty project error = %v", err)
	}
}

func TestNormalizeMatchResultUsesRequirementWeightsAndClamps(t *testing.T) {
	result := &CopilotMatchResult{
		MatchScore: 999,
		RequirementMap: []CopilotRequirement{
			{Priority: "required", Status: "matched"},
			{Priority: "preferred", Status: "missing"},
		},
	}
	normalizeMatchResult(result)
	if result.MatchScore != 60 {
		t.Fatalf("normalized score = %d, want 60", result.MatchScore)
	}
	result.RequirementMap = nil
	result.MatchScore = -10
	normalizeMatchResult(result)
	if result.MatchScore != 0 {
		t.Fatalf("clamped score = %d, want 0", result.MatchScore)
	}
}

func TestResumeCopilotApplyCreatesVersionAndRejectsStaleProposal(t *testing.T) {
	service, db, resume, version := newCopilotTestService(t)
	updated, err := service.Apply(context.Background(), 7, &CopilotApplyInput{
		ResumeID: resume.ID, BaseVersionID: version.ID, Content: version.Content,
		ProjectIndex: intPtr(0), ReplacementDescription: "使用 Go 重构支付接口，降低延迟", ReplacementTechStack: []string{"Go", "MySQL"},
		ChangeNote: "Copilot 项目优化",
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if updated.ID == version.ID || updated.ChangeNote != "Copilot 项目优化" {
		t.Fatalf("new version = %#v", updated)
	}
	var content ResumeContent
	if err := json.Unmarshal([]byte(updated.Content), &content); err != nil {
		t.Fatalf("decode applied content: %v", err)
	}
	if content.Project[0].Description != "使用 Go 重构支付接口，降低延迟" || len(content.Project[0].TechStack) != 2 {
		t.Fatalf("applied project = %#v", content.Project[0])
	}
	if _, err := service.Apply(context.Background(), 7, &CopilotApplyInput{
		ResumeID: resume.ID, BaseVersionID: version.ID, Content: version.Content,
	}); err == nil || err.Error() != "resume has changed, please refresh the copilot proposal" {
		t.Fatalf("stale proposal error = %v", err)
	}
	var persisted models.Resume
	if err := db.First(&persisted, resume.ID).Error; err != nil {
		t.Fatalf("reload resume: %v", err)
	}
	if persisted.CurrentVersionID != updated.ID {
		t.Fatalf("current version = %d, want %d", persisted.CurrentVersionID, updated.ID)
	}
}

func intPtr(value int) *int { return &value }
