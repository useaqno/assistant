// Package llm is the provider-agnostic AI layer (docs/context.md §6). A single
// interface hides Anthropic / OpenAI / OpenAI-compatible / Ollama behind a
// normalized tool-calling shape, so the rest of the daemon never depends on a
// specific provider.
package llm

import "context"

// Role is a chat role.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSystem    Role = "system"
)

// Message is one turn in a completion request.
type Message struct {
	Role    Role
	Content string
	// ToolCalls is set on assistant turns that invoked tools.
	ToolCalls []ToolCall
	// ToolCallID/Name link a tool result back to its call.
	ToolCallID string
	Name       string
}

// Tool is a normalized tool definition exposed to the model.
type Tool struct {
	Name        string
	Description string
	// Parameters is a JSON Schema object describing the arguments.
	Parameters map[string]any
}

// ToolCall is the model's request to invoke a tool.
type ToolCall struct {
	ID   string
	Name string
	// Args is the raw JSON arguments object.
	Args map[string]any
}

// CompletionRequest is the normalized request across providers.
type CompletionRequest struct {
	System      string
	Messages    []Message
	Tools       []Tool
	JSONMode    bool
	MaxTokens   int
	Temperature float32
}

// CompletionResponse is the normalized response.
type CompletionResponse struct {
	Text      string
	ToolCalls []ToolCall
}

// Chunk is a streamed token (or a terminal tool-call).
type Chunk struct {
	Text     string
	Done     bool
	ToolCall *ToolCall
	Err      error
}

// LLM abstracts a chat-completions provider.
type LLM interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan Chunk, error)
	Embed(ctx context.Context, input []string) ([][]float32, error)
}

// Config carries the resolved provider settings.
type Config struct {
	Provider    string // anthropic | openai | openai_compatible | ollama
	Model       string
	BaseURL     string
	APIKey      string
	MaxTokens   int
	Temperature float32
}
