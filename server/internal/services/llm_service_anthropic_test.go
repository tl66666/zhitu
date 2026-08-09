package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zhitu/server/internal/config"
)

func TestAnthropicChatUsesServerSideKeyAndMessagesProtocol(t *testing.T) {
	t.Parallel()

	var received anthropicRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/anthropic/v1/messages" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != "server-secret" {
			t.Errorf("X-API-Key = %q", got)
		}
		if got := r.Header.Get("Anthropic-Version"); got != "2023-06-01" {
			t.Errorf("Anthropic-Version = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"type":"message",
			"role":"assistant",
			"model":"mimo-v2.5",
			"content":[{"type":"thinking","thinking":"hidden"},{"type":"text","text":"OK"}],
			"stop_reason":"end_turn",
			"usage":{"input_tokens":10,"output_tokens":2}
		}`)
	}))
	defer server.Close()

	service := NewLLMService(&config.LLMConfig{
		Provider:   "anthropic",
		BaseURL:    server.URL + "/anthropic/v1/messages",
		APIKey:     "server-secret",
		ChatModel:  "mimo-v2.5",
		MaxTokens:  128,
		TimeoutSec: 5,
	})
	if got, want := service.openAIBaseURL(), server.URL+"/v1"; got != want {
		t.Fatalf("openAIBaseURL() = %q, want %q", got, want)
	}
	got, err := service.Chat(context.Background(), []ChatMessage{
		{Role: "system", Content: "system instruction"},
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if got != "OK" {
		t.Fatalf("Chat() = %q", got)
	}
	if received.System != "system instruction" {
		t.Errorf("system = %q", received.System)
	}
	if received.Thinking == nil || received.Thinking.Type != "disabled" {
		t.Errorf("thinking = %#v, want disabled", received.Thinking)
	}
	if len(received.Messages) != 1 ||
		received.Messages[0].Role != "user" ||
		received.Messages[0].Content != "hello" {
		t.Errorf("messages = %#v", received.Messages)
	}
}

func TestAnthropicChatStreamReadsTextDeltas(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `event: content_block_delta`)
		fmt.Fprintln(w, `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"你"}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `event: content_block_delta`)
		fmt.Fprintln(w, `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"好"}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `event: message_stop`)
		fmt.Fprintln(w, `data: {"type":"message_stop"}`)
	}))
	defer server.Close()

	service := NewLLMService(&config.LLMConfig{
		Provider:   "anthropic",
		BaseURL:    server.URL + "/messages",
		APIKey:     "server-secret",
		ChatModel:  "mimo-v2.5",
		MaxTokens:  128,
		TimeoutSec: 5,
	})
	var result strings.Builder
	err := service.ChatStream(
		context.Background(),
		[]ChatMessage{{Role: "user", Content: "hello"}},
		func(delta string) { result.WriteString(delta) },
	)
	if err != nil {
		t.Fatalf("ChatStream() error = %v", err)
	}
	if got := result.String(); got != "你好" {
		t.Fatalf("stream result = %q", got)
	}
}

func TestAnthropicChatStreamRejectsErrorEvent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"type":"error","error":{"type":"overloaded_error","message":"try again"}}`)
	}))
	defer server.Close()

	service := NewLLMService(&config.LLMConfig{
		Provider:   "anthropic",
		BaseURL:    server.URL + "/messages",
		APIKey:     "server-secret",
		ChatModel:  "mimo-v2.5",
		MaxTokens:  128,
		TimeoutSec: 5,
	})
	err := service.ChatStream(
		context.Background(),
		[]ChatMessage{{Role: "user", Content: "hello"}},
		nil,
	)
	if !errors.Is(err, ErrLLMStreamFailed) {
		t.Fatalf("ChatStream() error = %v, want ErrLLMStreamFailed", err)
	}
}

func TestAnthropicChatStreamRejectsIncompleteStream(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"partial"}}`)
	}))
	defer server.Close()

	service := NewLLMService(&config.LLMConfig{
		Provider:   "anthropic",
		BaseURL:    server.URL + "/messages",
		APIKey:     "server-secret",
		ChatModel:  "mimo-v2.5",
		MaxTokens:  128,
		TimeoutSec: 5,
	})
	err := service.ChatStream(
		context.Background(),
		[]ChatMessage{{Role: "user", Content: "hello"}},
		nil,
	)
	if !errors.Is(err, ErrLLMStreamFailed) {
		t.Fatalf("ChatStream() error = %v, want ErrLLMStreamFailed", err)
	}
}

func TestExtractJSONObjectWithSurroundingText(t *testing.T) {
	input := `分析结果如下： {"reply":"包含 { 花括号 } 和 \"引号\"","match":null} 请查收。`
	want := `{"reply":"包含 { 花括号 } 和 \"引号\"","match":null}`
	if got := extractJSONObject(input); got != want {
		t.Fatalf("extractJSONObject() = %q, want %q", got, want)
	}
}
