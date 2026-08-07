package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhitu/server/internal/models"
	"github.com/zhitu/server/internal/utils"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SendVoice 用户发送语音回答，先 Whisper 转写，再走文字逻辑
// 返回 AI 下一题的消息（含 TTS audio_url）
func (s *InterviewService) SendVoice(ctx context.Context, userID, interviewID uint, audio io.Reader, filename string, onDelta func(string)) (*models.InterviewMessage, error) {
	return s.SendVoiceWithDuration(ctx, userID, interviewID, audio, filename, 0, onDelta)
}

// SendVoiceWithDuration is the no-preview voice answer path used by voice and hybrid rooms.
func (s *InterviewService) SendVoiceWithDuration(ctx context.Context, userID, interviewID uint, audio io.Reader, filename string, durationSec int, onDelta func(string)) (*models.InterviewMessage, error) {
	// 1. 保存音频文件到磁盘
	_, absPath, err := utils.SaveFile(audio, s.storage.AudioDir, filename)
	if err != nil {
		return nil, fmt.Errorf("save audio: %w", err)
	}

	// 2. Whisper 转写（重新读取磁盘文件）
	transcribed, err := s.transcribeFromPath(ctx, absPath)
	if err != nil {
		return nil, fmt.Errorf("transcribe: %w", err)
	}

	// 3. 存用户回答（含音频 URL）
	interview, err := s.Get(userID, interviewID)
	if err != nil {
		return nil, err
	}
	if interview.Status != StatusOngoing {
		if interview.Status == StatusPreparing {
			return nil, ErrInterviewPreparing
		}
		return nil, ErrInterviewEnded
	}
	if strings.TrimSpace(interview.ResumeSnapshot) == "" {
		return nil, ErrResumeRequired
	}
	if interview.Mode != ModeVoice && interview.Mode != ModeHybrid {
		return nil, ErrModeNotAllowed
	}

	audioURL, err := storageURL(s.storage.BaseDir, absPath)
	if err != nil {
		return nil, err
	}
	userMsg := &models.InterviewMessage{
		InterviewID: interviewID,
		Role:        "user",
		Content:     transcribed,
		AudioURL:    audioURL,
		InputMode:   ModeVoice,
		QuestionNo:  interview.CurrentQuestionNo,
		DurationSec: durationSec,
	}
	if err := s.db.Create(userMsg).Error; err != nil {
		return nil, err
	}

	if interview.CurrentQuestionNo >= interview.TotalQuestions {
		return nil, s.endAndGenerateReport(ctx, interview)
	}

	// 4. AI 生成下一题
	return s.askNextQuestionWithStream(ctx, interview, interview.CurrentQuestionNo+1, onDelta)
}

// TranscribeVoice 仅转写语音草稿，不创建回答消息，也不推进面试进度。
func (s *InterviewService) TranscribeVoice(ctx context.Context, userID, interviewID uint, audio io.Reader, filename string) (string, error) {
	interview, err := s.Get(userID, interviewID)
	if err != nil {
		return "", err
	}
	if interview.Status != StatusOngoing {
		return "", ErrInterviewEnded
	}

	transcribed, err := s.llm.Transcribe(ctx, audio, filename)
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}
	return transcribed, nil
}

// transcribeFromPath 转写用户上传的语音回答
// 模型：cfg.WhisperModel → "mimo-v2.5-asr"（语音识别）
func (s *InterviewService) transcribeFromPath(ctx context.Context, absPath string) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("open audio file: %w", err)
	}
	defer f.Close()
	return s.llm.Transcribe(ctx, f, filepath.Base(absPath))
}

// GetTTS 获取某条 AI 提问的 TTS 音频
// 若消息已有 audio_url 则直接返回路径，否则现场合成并保存
//
// 模型：cfg.TTSModel → "mimo-v2.5-tts"（语音生成）
// 用途：面试语音生成（AI 面试官朗读，前端用户可选开关）
func (s *InterviewService) GetTTS(ctx context.Context, userID, interviewID, messageID uint) ([]byte, string, error) {
	if _, err := s.Get(userID, interviewID); err != nil {
		return nil, "", err
	}

	var msg models.InterviewMessage
	err := s.db.Where("id = ? AND interview_id = ? AND role = ?", messageID, interviewID, "assistant").First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", ErrMessageNotFound
	}
	if err != nil {
		return nil, "", err
	}

	// 若已有 audio_url，读取磁盘文件返回
	if msg.AudioURL != "" {
		storedPath := strings.TrimPrefix(msg.AudioURL, "/static/")
		full := storedPath
		if !filepath.IsAbs(storedPath) {
			full = filepath.Join(s.storage.BaseDir, filepath.FromSlash(storedPath))
		}
		data, err := os.ReadFile(full)
		if err == nil {
			return data, filepath.Base(full), nil
		}
		// 读取失败则重新合成
	}

	// 现场合成
	audio, err := s.llm.Synthesize(ctx, msg.Content)
	if err != nil {
		return nil, "", fmt.Errorf("synthesize: %w", err)
	}

	// 保存到磁盘
	filename := fmt.Sprintf("tts_%d.wav", msg.ID)
	_, absPath, err := utils.SaveBytes(audio, s.storage.TTSDir, filename)
	if err != nil {
		return nil, "", fmt.Errorf("save tts: %w", err)
	}

	// 更新消息的 audio_url
	audioURL, err := storageURL(s.storage.BaseDir, absPath)
	if err != nil {
		return nil, "", err
	}
	s.db.Model(&msg).Update("audio_url", audioURL)

	return audio, filename, nil
}

func storageURL(baseDir, absPath string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("resolve storage base: %w", err)
	}
	fileAbs, err := filepath.Abs(absPath)
	if err != nil {
		return "", fmt.Errorf("resolve stored file: %w", err)
	}
	rel, err := filepath.Rel(baseAbs, fileAbs)
	if err != nil {
		return "", fmt.Errorf("make storage url: %w", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("stored file is outside storage base")
	}
	return "/static/" + filepath.ToSlash(rel), nil
}
