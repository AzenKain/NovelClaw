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

// GeminiClient implements LLMClient for Google Gemini Native REST API.
type GeminiClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewGeminiClient creates a new GeminiClient.
func NewGeminiClient(apiKey string, timeout time.Duration) *GeminiClient {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProviderName returns the provider identifier.
func (c *GeminiClient) ProviderName() string {
	return "gemini_native"
}

// safetySettingsBlockNone disables safety filters to allow uncensored novel translation.
var safetySettingsBlockNone = []map[string]string{
	{"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_NONE"},
	{"category": "HARM_CATEGORY_HATE_SPEECH", "threshold": "BLOCK_NONE"},
	{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT", "threshold": "BLOCK_NONE"},
	{"category": "HARM_CATEGORY_DANGEROUS_CONTENT", "threshold": "BLOCK_NONE"},
	{"category": "HARM_CATEGORY_CIVIC_INTEGRITY", "threshold": "BLOCK_NONE"},
}

// convertToGeminiPayload converts a generic CompletionRequest into Gemini REST format.
func convertToGeminiPayload(req CompletionRequest) map[string]any {
	var systemInstruction map[string]any
	var contents []map[string]any

	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			systemInstruction = map[string]any{
				"parts": []map[string]string{
					{"text": msg.Content},
				},
			}
			continue
		}

		role := "user"
		if msg.Role == RoleAssistant {
			role = "model"
		}

		contents = append(contents, map[string]any{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}

	payload := map[string]any{
		"contents":       contents,
		"safetySettings": safetySettingsBlockNone,
		"generationConfig": map[string]any{
			"temperature": req.Temperature,
		},
	}

	if systemInstruction != nil {
		payload["systemInstruction"] = systemInstruction
	}
	if req.MaxTokens > 0 {
		payload["generationConfig"].(map[string]any)["maxOutputTokens"] = req.MaxTokens
	}

	if req.ReasoningEffort != "" {
		var budget int
		switch strings.ToLower(strings.TrimSpace(req.ReasoningEffort)) {
		case "off":
			budget = 0
		case "low":
			budget = 1024
		case "medium":
			budget = 2048
		case "high":
			budget = 4096
		case "xhigh":
			budget = 8192
		case "max":
			budget = 16384
		default:
			budget = -1
		}
		if budget >= 0 {
			genConfig := payload["generationConfig"].(map[string]any)
			genConfig["thinkingConfig"] = map[string]any{
				"thinkingBudget": budget,
			}
		}
	}

	if len(req.Tools) > 0 {
		var decls []map[string]any
		for _, tool := range req.Tools {
			decls = append(decls, map[string]any{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			})
		}
		payload["tools"] = []map[string]any{
			{"functionDeclarations": decls},
		}
	}

	return payload
}

// Generate sends a completion request to Gemini REST API.
func (c *GeminiClient) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	model = strings.TrimPrefix(model, "models/")

	apiURL := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, model, c.apiKey)
	payload := convertToGeminiPayload(req)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create gemini request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute gemini request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api error [status %d]: %s", resp.StatusCode, string(respBytes))
	}

	var rawResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string `json:"text"`
					Thought      bool   `json:"thought"`
					FunctionCall *struct {
						Name string         `json:"name"`
						Args map[string]any `json:"args"`
					} `json:"functionCall"`
				} `json:"parts"`
				Role string `json:"role"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(respBytes, &rawResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	if len(rawResp.Candidates) == 0 || len(rawResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response candidate from gemini")
	}

	var sb strings.Builder
	var thoughtSb strings.Builder
	var toolCalls []ToolCall

	for i, p := range rawResp.Candidates[0].Content.Parts {
		if p.Text != "" {
			if p.Thought {
				thoughtSb.WriteString(p.Text)
			} else {
				sb.WriteString(p.Text)
			}
		}
		if p.FunctionCall != nil {
			argsBytes, _ := json.Marshal(p.FunctionCall.Args)
			var call ToolCall
			call.ID = fmt.Sprintf("call_gemini_%d", i)
			call.Type = "function"
			call.Function.Name = p.FunctionCall.Name
			call.Function.Arguments = string(argsBytes)
			toolCalls = append(toolCalls, call)
		}
	}

	return &CompletionResponse{
		Model:        model,
		Content:      strings.TrimSpace(sb.String()),
		Thought:      strings.TrimSpace(thoughtSb.String()),
		ToolCalls:    toolCalls,
		FinishReason: rawResp.Candidates[0].FinishReason,
		PromptTokens: rawResp.UsageMetadata.PromptTokenCount,
		CompTokens:   rawResp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:  rawResp.UsageMetadata.TotalTokenCount,
	}, nil
}

// Stream streams response chunks from Gemini REST API.
func (c *GeminiClient) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	model = strings.TrimPrefix(model, "models/")

	apiURL := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", c.baseURL, model, c.apiKey)
	payload := convertToGeminiPayload(req)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create gemini stream http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http execute gemini stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini stream api error [status %d]: %s", resp.StatusCode, string(errBytes))
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
			var sse struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text    string `json:"text"`
							Thought bool   `json:"thought"`
						} `json:"parts"`
					} `json:"content"`
					FinishReason string `json:"finishReason"`
				} `json:"candidates"`
			}

			if err := json.Unmarshal([]byte(dataStr), &sse); err != nil {
				continue
			}

			if len(sse.Candidates) > 0 && len(sse.Candidates[0].Content.Parts) > 0 {
				part := sse.Candidates[0].Content.Parts[0]
				finishReason := sse.Candidates[0].FinishReason
				chunk := StreamChunk{
					FinishReason: finishReason,
				}
				if part.Thought {
					chunk.Thought = part.Text
				} else {
					chunk.Content = part.Text
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
