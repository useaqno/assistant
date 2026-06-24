package llm

import "context"

// ollama implements the LLM interface against a local Ollama server.
type ollama struct {
	cfg  Config
	base string
}

func (o *ollama) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := make([]map[string]any, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, map[string]any{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]any{"role": string(m.Role), "content": m.Content})
	}
	body := map[string]any{
		"model":    o.cfg.Model,
		"messages": msgs,
		"stream":   false,
		"options":  map[string]any{"temperature": req.Temperature},
	}
	if len(req.Tools) > 0 {
		body["tools"] = openAITools(req.Tools) // Ollama follows the OpenAI tool shape
	}

	var resp struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string         `json:"name"`
					Arguments map[string]any `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	}
	if err := postJSON(ctx, o.base+"/api/chat", nil, body, &resp); err != nil {
		return CompletionResponse{}, err
	}
	out := CompletionResponse{Text: resp.Message.Content}
	for _, tc := range resp.Message.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCall{Name: tc.Function.Name, Args: tc.Function.Arguments})
	}
	return out, nil
}

func (o *ollama) Stream(ctx context.Context, req CompletionRequest) (<-chan Chunk, error) {
	return emulateStream(o.Complete(ctx, req))
}

func (o *ollama) Embed(ctx context.Context, input []string) ([][]float32, error) {
	out := make([][]float32, 0, len(input))
	for _, in := range input {
		var resp struct {
			Embedding []float32 `json:"embedding"`
		}
		body := map[string]any{"model": o.cfg.Model, "prompt": in}
		if err := postJSON(ctx, o.base+"/api/embeddings", nil, body, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Embedding)
	}
	return out, nil
}
