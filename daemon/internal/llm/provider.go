package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewClient builds a provider client from config. The bool reports whether a
// usable client could be constructed (e.g. a cloud provider needs an API key).
func NewClient(cfg Config) (LLM, bool) {
	switch cfg.Provider {
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, false
		}
		return &anthropic{cfg: cfg, base: orDefault(cfg.BaseURL, "https://api.anthropic.com")}, true
	case "openai":
		if cfg.APIKey == "" {
			return nil, false
		}
		return &openAI{cfg: cfg, base: orDefault(cfg.BaseURL, "https://api.openai.com")}, true
	case "openai_compatible":
		if cfg.BaseURL == "" {
			return nil, false
		}
		return &openAI{cfg: cfg, base: cfg.BaseURL}, true
	case "ollama":
		return &ollama{cfg: cfg, base: orDefault(cfg.BaseURL, "http://localhost:11434")}, true
	default:
		return nil, false
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

var httpClient = &http.Client{Timeout: 90 * time.Second}

// postJSON sends a JSON body and decodes the JSON response into out.
func postJSON(ctx context.Context, url string, headers map[string]string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("%s: %d %s", url, res.StatusCode, truncate(string(data), 300))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// emulateStream turns a one-shot completion into a Chunk channel (used by
// providers without a streaming path wired yet).
func emulateStream(resp CompletionResponse, err error) (<-chan Chunk, error) {
	if err != nil {
		return nil, err
	}
	ch := make(chan Chunk, 4)
	go func() {
		defer close(ch)
		if resp.Text != "" {
			ch <- Chunk{Text: resp.Text}
		}
		for i := range resp.ToolCalls {
			tc := resp.ToolCalls[i]
			ch <- Chunk{ToolCall: &tc}
		}
		ch <- Chunk{Done: true}
	}()
	return ch, nil
}
