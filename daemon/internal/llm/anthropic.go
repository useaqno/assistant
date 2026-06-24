package llm

import (
	"context"
	"errors"
)

// anthropic implements the LLM interface against the Messages API.
type anthropic struct {
	cfg  Config
	base string
}

func (a *anthropic) headers() map[string]string {
	return map[string]string{
		"x-api-key":         a.cfg.APIKey,
		"anthropic-version": "2023-06-01",
	}
}

func (a *anthropic) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	body := map[string]any{
		"model":       a.cfg.Model,
		"max_tokens":  pick(req.MaxTokens, 2000),
		"temperature": req.Temperature,
		"messages":    anthropicMessages(req.Messages),
	}
	if req.System != "" {
		body["system"] = req.System
	}
	if len(req.Tools) > 0 {
		body["tools"] = anthropicTools(req.Tools)
	}

	var resp struct {
		Content []struct {
			Type  string         `json:"type"`
			Text  string         `json:"text"`
			ID    string         `json:"id"`
			Name  string         `json:"name"`
			Input map[string]any `json:"input"`
		} `json:"content"`
	}
	if err := postJSON(ctx, a.base+"/v1/messages", a.headers(), body, &resp); err != nil {
		return CompletionResponse{}, err
	}
	var out CompletionResponse
	for _, c := range resp.Content {
		switch c.Type {
		case "text":
			out.Text += c.Text
		case "tool_use":
			out.ToolCalls = append(out.ToolCalls, ToolCall{ID: c.ID, Name: c.Name, Args: c.Input})
		}
	}
	return out, nil
}

func (a *anthropic) Stream(ctx context.Context, req CompletionRequest) (<-chan Chunk, error) {
	return emulateStream(a.Complete(ctx, req))
}

func (a *anthropic) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, errors.New("anthropic: embeddings not supported; use a local embeddings model")
}

func anthropicMessages(msgs []Message) []map[string]any {
	out := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		role := string(m.Role)
		if role == "system" {
			continue // system goes in its own field
		}
		out = append(out, map[string]any{"role": role, "content": m.Content})
	}
	return out
}

func anthropicTools(tools []Tool) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		out = append(out, map[string]any{
			"name":         t.Name,
			"description":  t.Description,
			"input_schema": t.Parameters,
		})
	}
	return out
}

func pick(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}
