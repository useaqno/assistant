package llm

import (
	"context"
	"encoding/json"
)

// openAI implements the LLM interface against the Chat Completions API. It also
// serves OpenAI-compatible endpoints (Groq, OpenRouter, …) via a custom base.
type openAI struct {
	cfg  Config
	base string
}

func (o *openAI) headers() map[string]string {
	return map[string]string{"Authorization": "Bearer " + o.cfg.APIKey}
}

func (o *openAI) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := make([]map[string]any, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, map[string]any{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]any{"role": string(m.Role), "content": m.Content})
	}
	body := map[string]any{
		"model":       o.cfg.Model,
		"messages":    msgs,
		"max_tokens":  pick(req.MaxTokens, 2000),
		"temperature": req.Temperature,
	}
	if len(req.Tools) > 0 {
		body["tools"] = openAITools(req.Tools)
		body["tool_choice"] = "auto"
	}
	if req.JSONMode {
		body["response_format"] = map[string]string{"type": "json_object"}
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := postJSON(ctx, o.base+"/v1/chat/completions", o.headers(), body, &resp); err != nil {
		return CompletionResponse{}, err
	}
	var out CompletionResponse
	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		out.Text = msg.Content
		for _, tc := range msg.ToolCalls {
			args := map[string]any{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			out.ToolCalls = append(out.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Args: args})
		}
	}
	return out, nil
}

func (o *openAI) Stream(ctx context.Context, req CompletionRequest) (<-chan Chunk, error) {
	return emulateStream(o.Complete(ctx, req))
}

func (o *openAI) Embed(ctx context.Context, input []string) ([][]float32, error) {
	body := map[string]any{"model": o.cfg.Model, "input": input}
	var resp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := postJSON(ctx, o.base+"/v1/embeddings", o.headers(), body, &resp); err != nil {
		return nil, err
	}
	out := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

func openAITools(tools []Tool) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			},
		})
	}
	return out
}
