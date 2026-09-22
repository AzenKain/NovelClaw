package llm

import (
	"context"
)

// Role represents the role of a message author.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents an LLM conversation message.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolDefinition defines a function tool for LLM function calling.
type ToolDefinition struct {
	Type     string         `json:"type"`
	Function FunctionSchema `json:"function"`
}

// FunctionSchema describes the function parameters using JSON Schema.
type FunctionSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall represents a tool call emitted by an LLM.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// CompletionRequest represents an LLM completion request.
type CompletionRequest struct {
	Model           string           `json:"model"`
	Messages        []Message        `json:"messages"`
	Temperature     float32          `json:"temperature,omitempty"`
	TopP            float32          `json:"top_p,omitempty"`
	MaxTokens       int              `json:"max_tokens,omitempty"`
	Stream          bool             `json:"stream,omitempty"`
	Tools           []ToolDefinition `json:"tools,omitempty"`
	ReasoningEffort string           `json:"reasoning_effort,omitempty"`
}

// CompletionResponse represents an LLM completion response.
type CompletionResponse struct {
	ID              string     `json:"id"`
	Model           string     `json:"model"`
	Content         string     `json:"content"`
	Thought         string     `json:"thought,omitempty"`
	ToolCalls       []ToolCall `json:"tool_calls,omitempty"`
	FinishReason    string     `json:"finish_reason"`
	PromptTokens    int        `json:"prompt_tokens"`
	CompTokens      int        `json:"comp_tokens"`
	TotalTokens     int        `json:"total_tokens"`
	ReasoningTokens int        `json:"reasoning_tokens,omitempty"`
}

// StreamChunk represents a chunk received during streaming.
type StreamChunk struct {
	Content         string     `json:"content"`
	Thought         string     `json:"thought,omitempty"`
	ToolCalls       []ToolCall `json:"tool_calls,omitempty"`
	FinishReason    string     `json:"finish_reason,omitempty"`
	ReasoningTokens int        `json:"reasoning_tokens,omitempty"`
	Error           error      `json:"error,omitempty"`
}

// LLMClient provides an interface for interacting with LLM providers.
type LLMClient interface {
	Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	ProviderName() string
}
