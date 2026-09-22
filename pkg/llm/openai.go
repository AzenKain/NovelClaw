package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient implements LLMClient for OpenAI-compatible endpoints.
type OpenAIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAIClient.
func NewOpenAIClient(baseURL, apiKey string, timeout time.Duration) *OpenAIClient {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	endpoint := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(endpoint, "/v1") {
		endpoint += "/v1"
	}
	return &OpenAIClient{
		baseURL: endpoint,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProviderName returns the provider identifier.
func (c *OpenAIClient) ProviderName() string {
	return "openai_compatible"
}

func normalizeOpenAIReasoningEffort(effort string) string {
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "low":
		return "low"
	case "medium":
		return "medium"
	case "high", "xhigh", "max":
		return "high"
	default:
		return ""
	}
}

// Generate sends a completion request and returns the response.
func (c *OpenAIClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	req.Stream = false
	req.ReasoningEffort = normalizeOpenAIReasoningEffort(req.ReasoningEffort)
	chatURL := c.baseURL + "/chat/completions"

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http execute: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error [status %d]: %s", resp.StatusCode, string(respBytes))
	}

	var rawResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role             string     `json:"role"`
				Content          string     `json:"content"`
				ReasoningContent string     `json:"reasoning_content"`
				Reasoning        string     `json:"reasoning"`
				ToolCalls        []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			CompTokens   int `json:"completion_tokens"`
			TotalTokens  int `json:"total_tokens"`
			CompDetails  struct {
				ReasoningTokens int `json:"reasoning_tokens"`
			} `json:"completion_tokens_details"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBytes, &rawResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(rawResp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices in api response")
	}

	choice := rawResp.Choices[0]
	thought := choice.Message.ReasoningContent
	if thought == "" {
		thought = choice.Message.Reasoning
	}

	return &CompletionResponse{
		ID:              rawResp.ID,
		Model:           rawResp.Model,
		Content:         choice.Message.Content,
		Thought:         thought,
		ToolCalls:       choice.Message.ToolCalls,
		FinishReason:    choice.FinishReason,
		PromptTokens:    rawResp.Usage.PromptTokens,
		CompTokens:      rawResp.Usage.CompTokens,
		TotalTokens:     rawResp.Usage.TotalTokens,
		ReasoningTokens: rawResp.Usage.CompDetails.ReasoningTokens,
	}, nil
}

// Stream streams completion chunks using Server-Sent Events.
func (c *OpenAIClient) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	req.Stream = true
	req.ReasoningEffort = normalizeOpenAIReasoningEffort(req.ReasoningEffort)
	chatURL := c.baseURL + "/chat/completions"

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http execute stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api stream error [status %d]: %s", resp.StatusCode, string(errBytes))
	}

	ch := make(chan StreamChunk, 20)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err()}
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}

			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if dataStr == "[DONE]" {
				return
			}

			var sse struct {
				Choices []struct {
					Delta struct {
						Content          string     `json:"content"`
						ReasoningContent string     `json:"reasoning_content"`
						Reasoning        string     `json:"reasoning"`
						ToolCalls        []ToolCall `json:"tool_calls"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Usage struct {
					CompDetails struct {
						ReasoningTokens int `json:"reasoning_tokens"`
					} `json:"completion_tokens_details"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(dataStr), &sse); err != nil {
				continue
			}

			if len(sse.Choices) > 0 {
				delta := sse.Choices[0].Delta
				finishReason := sse.Choices[0].FinishReason
				thought := delta.ReasoningContent
				if thought == "" {
					thought = delta.Reasoning
				}
				chunk := StreamChunk{
					Content:         delta.Content,
					Thought:         thought,
					ToolCalls:       delta.ToolCalls,
					FinishReason:    finishReason,
					ReasoningTokens: sse.Usage.CompDetails.ReasoningTokens,
				}
				select {
				case <-ctx.Done():
					return
				case ch <- chunk:
				}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
			case ch <- StreamChunk{Error: err}:
			}
		}
	}()

	return ch, nil
}
