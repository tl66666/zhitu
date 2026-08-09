package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *LLMService) isAnthropic() bool {
	provider := strings.ToLower(strings.TrimSpace(s.cfg.Provider))
	return provider == "anthropic" || provider == "mimo-anthropic"
}

func (s *LLMService) anthropicEndpoint() string {
	base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	if strings.HasSuffix(base, "/messages") {
		return base
	}
	return base + "/messages"
}

func (s *LLMService) newAnthropicRequest(ctx context.Context, payload interface{}) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.anthropicEndpoint(),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.cfg.APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	return req, nil
}

func toAnthropicRequest(reqBody chatRequest) anthropicRequest {
	var systemParts []string
	messages := make([]ChatMessage, 0, len(reqBody.Messages))
	for _, message := range reqBody.Messages {
		switch strings.ToLower(message.Role) {
		case "system":
			if strings.TrimSpace(message.Content) != "" {
				systemParts = append(systemParts, message.Content)
			}
		case "assistant":
			messages = append(messages, ChatMessage{Role: "assistant", Content: message.Content})
		default:
			messages = append(messages, ChatMessage{Role: "user", Content: message.Content})
		}
	}
	return anthropicRequest{
		Model:       reqBody.Model,
		System:      strings.Join(systemParts, "\n\n"),
		Messages:    messages,
		Temperature: reqBody.Temperature,
		MaxTokens:   reqBody.MaxTokens,
		Stream:      reqBody.Stream,
		Thinking:    &anthropicThinking{Type: "disabled"},
	}
}

func (s *LLMService) doChatAnthropic(ctx context.Context, reqBody chatRequest) (*chatResponse, error) {
	req, err := s.newAnthropicRequest(ctx, toAnthropicRequest(reqBody))
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, s.wrapAPIError(resp)
	}

	var anthropicResult anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResult); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var content strings.Builder
	for _, block := range anthropicResult.Content {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}

	return &chatResponse{
		Choices: []chatResponseChoice{{
			Message: chatResponseMessage{
				Role:    "assistant",
				Content: content.String(),
			},
			FinishReason: anthropicResult.StopReason,
		}},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     anthropicResult.Usage.InputTokens,
			CompletionTokens: anthropicResult.Usage.OutputTokens,
			TotalTokens:      anthropicResult.Usage.InputTokens + anthropicResult.Usage.OutputTokens,
		},
	}, nil
}

func (s *LLMService) chatStreamAnthropic(
	ctx context.Context,
	reqBody chatRequest,
	onDelta func(delta string),
) error {
	anthropicBody := toAnthropicRequest(reqBody)
	anthropicBody.Stream = true
	req, err := s.newAnthropicRequest(ctx, anthropicBody)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return s.wrapAPIError(resp)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	sawText := false
	sawStop := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var event anthropicStreamEvent
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "error" {
			return fmt.Errorf(
				"%w: anthropic %s: %s",
				ErrLLMStreamFailed,
				event.Error.Type,
				truncate(event.Error.Message, 500),
			)
		}
		if event.Type == "message_stop" {
			sawStop = true
			break
		}
		if event.Type == "content_block_delta" &&
			event.Delta.Type == "text_delta" &&
			event.Delta.Text != "" {
			sawText = true
			if onDelta != nil {
				onDelta(event.Delta.Text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrLLMStreamFailed, err)
	}
	if !sawStop {
		return fmt.Errorf("%w: missing message_stop", ErrLLMStreamFailed)
	}
	if !sawText {
		return fmt.Errorf("%w: empty text stream", ErrLLMStreamFailed)
	}
	return nil
}
