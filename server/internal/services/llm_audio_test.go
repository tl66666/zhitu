package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zhitu/server/internal/config"
)

func TestSynthesizeUsesMiMoTTSProtocol(t *testing.T) {
	t.Parallel()

	var received ttsRequestMimo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("api-key"); got != "server-secret" {
			t.Errorf("api-key = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer server-secret" {
			t.Errorf("Authorization = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"audio":{"data":"d2F2"}}}]}`)
	}))
	defer server.Close()

	service := NewLLMService(&config.LLMConfig{
		Provider:   "anthropic",
		BaseURL:    server.URL + "/anthropic/v1/messages",
		APIKey:     "server-secret",
		TTSModel:   "mimo-v2.5-tts",
		TTSVoice:   "Chloe",
		TimeoutSec: 5,
	})
	audio, err := service.Synthesize(context.Background(), "您好，面试现在开始。")
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if got := string(audio); got != "wav" {
		t.Fatalf("Synthesize() = %q", got)
	}
	if received.Model != "mimo-v2.5-tts" {
		t.Errorf("model = %q", received.Model)
	}
	if len(received.Messages) != 1 ||
		received.Messages[0].Role != "assistant" ||
		received.Messages[0].Content != "您好，面试现在开始。" {
		t.Errorf("messages = %#v", received.Messages)
	}
	if received.Audio.Format != "wav" || received.Audio.Voice != "Chloe" {
		t.Errorf("audio = %#v", received.Audio)
	}
}
