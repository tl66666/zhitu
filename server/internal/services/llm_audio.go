package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// Transcribe 调用 MiMo ASR 将音频转写为文本
//
// mimo ASR 不走 OpenAI 的 /v1/audio/transcriptions，而是通过 /v1/chat/completions
// 将音频 base64 编码后放在 input_audio 类型的 content 中传入。
//
// 模型：cfg.WhisperModel → "mimo-v2.5-asr"（语音识别）
// 用途：面试语音识别（用户语音回答转文字）
func (s *LLMService) Transcribe(ctx context.Context, audio io.Reader, filename string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrLLMNotConfigured
	}

	// 1. 读取音频并 base64 编码
	audioBytes, err := io.ReadAll(audio)
	if err != nil {
		return "", fmt.Errorf("read audio: %w", err)
	}
	audioBase64 := base64.StdEncoding.EncodeToString(audioBytes)
	mime := audioMIME(filename)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, audioBase64)

	// 2. 构造 ASR 请求
	reqBody := asrRequest{
		Model: s.cfg.WhisperModel,
		Messages: []asrMessage{{
			Role: "user",
			Content: []asrContentItem{{
				Type:       "input_audio",
				InputAudio: asrInputAudio{Data: dataURL},
			}},
		}},
		ASROptions: struct {
			Language string `json:"language"`
		}{Language: "auto"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// 3. POST /v1/chat/completions
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.openAIBaseURL()+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// MiMo OpenAI-compatible audio endpoints authenticate with api-key.
	// Keep Bearer as a compatibility header for other compatible gateways.
	req.Header.Set("api-key", s.cfg.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", s.wrapAPIError(resp)
	}

	// 4. 解析响应：choices[0].message.content 是识别出的文字
	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode asr response: %w", err)
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", ErrLLMEmptyResponse
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// Synthesize 调用 MiMo TTS 将文本合成为 WAV 音频字节
//
// mimo TTS 不走 OpenAI 的 /v1/audio/speech，而是通过 /v1/chat/completions
// 目标文本放在 role=assistant 的 content 中，audio 参数指定格式和音色。
// 响应中 message.audio.data 是 base64 编码的音频字节。
//
// 模型：cfg.TTSModel → "mimo-v2.5-tts"（语音生成）
// 用途：面试语音生成（AI 面试官朗读，前端用户可选开关）
func (s *LLMService) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if !s.IsConfigured() {
		return nil, ErrLLMNotConfigured
	}
	if text == "" {
		return nil, errors.New("tts input text is empty")
	}

	// 音色兜底：未配置或为 OpenAI 默认值时用 mimo 预置音色
	voice := s.cfg.TTSVoice
	if voice == "" || voice == "alloy" {
		voice = "mimo_default"
	}

	// 1. 构造 TTS 请求
	reqBody := ttsRequestMimo{
		Model: s.cfg.TTSModel,
		Messages: []ttsMessage{{
			Role:    "assistant",
			Content: text,
		}},
		Audio: ttsAudioConfig{
			Format: "wav",
			Voice:  voice,
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// 2. POST /v1/chat/completions
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.openAIBaseURL()+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// MiMo OpenAI-compatible audio endpoints authenticate with api-key.
	// Keep Bearer as a compatibility header for other compatible gateways.
	req.Header.Set("api-key", s.cfg.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, s.wrapAPIError(resp)
	}

	// 3. 解析响应：message.audio.data 是 base64 编码的 WAV 音频
	var result ttsResponseMimo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode tts response: %w", err)
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Audio.Data == "" {
		return nil, ErrLLMEmptyResponse
	}

	// 4. base64 解码为音频字节
	audio, err := base64.StdEncoding.DecodeString(result.Choices[0].Message.Audio.Data)
	if err != nil {
		return nil, fmt.Errorf("decode base64 audio: %w", err)
	}
	return audio, nil
}

// audioMIME 根据音频文件扩展名返回 MIME 类型
func audioMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	default:
		return "audio/wav"
	}
}

// openAIBaseURL 返回语音接口使用的 OpenAI 兼容 API 根地址。
// 当聊天使用 MiMo Anthropic 入口时，语音能力仍位于同域名的 /v1 下。
func (s *LLMService) openAIBaseURL() string {
	base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	if !s.isAnthropic() {
		return base
	}
	if marker := strings.Index(base, "/anthropic/"); marker >= 0 {
		return base[:marker] + "/v1"
	}
	if strings.HasSuffix(base, "/messages") {
		return strings.TrimSuffix(base, "/messages")
	}
	return base
}
