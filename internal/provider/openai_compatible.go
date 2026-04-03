package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type OpenAICompatibleClient struct {
	httpClient HTTPDoer
}

func NewOpenAICompatibleClient(httpClient HTTPDoer) OpenAICompatibleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return OpenAICompatibleClient{httpClient: httpClient}
}

func (c OpenAICompatibleClient) Synthesize(ctx context.Context, request Request) (Result, error) {
	payload, err := marshalChatCompletionRequest(request)
	if err != nil {
		return Result{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsURL(request.Config), bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("build synthesis request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+strings.TrimSpace(request.Config.APIKey))
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return Result{}, fmt.Errorf("send synthesis request: %w", err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(httpResponse.Body, 2<<20))
	if err != nil {
		return Result{}, fmt.Errorf("read synthesis response: %w", err)
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return Result{}, httpStatusError(httpResponse.StatusCode, responseBody)
	}

	var response openAICompatibleChatResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return Result{}, fmt.Errorf("decode synthesis response: %w", err)
	}

	content, err := response.messageContent()
	if err != nil {
		return Result{}, err
	}

	result, err := parseSynthesisResult(content)
	if err != nil {
		return Result{}, err
	}

	result.SchemaVersion = ResultSchemaVersion
	result.GeneratedAt = time.Now().UTC()
	result.Provider = normalizeName(request.Config.Name)
	result.Model = strings.TrimSpace(request.Config.Model)
	result.Status = ResultStatusAvailable
	result.Used = true
	result.FallbackReason = ""

	return normalizeResultText(result), nil
}

type openAICompatibleChatRequest struct {
	Model    string                        `json:"model"`
	Messages []openAICompatibleChatMessage `json:"messages"`
}

type openAICompatibleChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatibleChatResponse struct {
	Choices []openAICompatibleChoice `json:"choices"`
}

type openAICompatibleChoice struct {
	Message openAICompatibleMessage `json:"message"`
}

type openAICompatibleMessage struct {
	Content string `json:"content"`
}

func marshalChatCompletionRequest(request Request) ([]byte, error) {
	systemPrompt, userPrompt, err := buildSynthesisPrompt(request.Context)
	if err != nil {
		return nil, err
	}

	payload := openAICompatibleChatRequest{
		Model: strings.TrimSpace(request.Config.Model),
		Messages: []openAICompatibleChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal synthesis request: %w", err)
	}
	return body, nil
}

func (r openAICompatibleChatResponse) messageContent() (string, error) {
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("synthesis response did not include any choices")
	}

	content := strings.TrimSpace(r.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("synthesis response returned an empty message")
	}

	return content, nil
}

func chatCompletionsURL(config Config) string {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" && normalizeName(config.Name) == "openai" {
		baseURL = defaultOpenAIBaseURL
	}
	return baseURL + "/chat/completions"
}

func httpStatusError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("synthesis request failed with status %d", statusCode)
	}
	return fmt.Errorf("synthesis request failed with status %d: %s", statusCode, message)
}

func normalizeResultText(result Result) Result {
	result.ProjectSummary = strings.TrimSpace(result.ProjectSummary)
	result.ArchitectureNarrative = strings.TrimSpace(result.ArchitectureNarrative)
	result.HotspotExplanations = trimHotspotExplanations(result.HotspotExplanations)
	result.ReadingPathExplanations = trimReadingPathExplanations(result.ReadingPathExplanations)
	return result
}

func trimHotspotExplanations(items []HotspotExplanation) []HotspotExplanation {
	trimmed := make([]HotspotExplanation, 0, len(items))
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		explanation := strings.TrimSpace(item.Explanation)
		if path == "" || explanation == "" {
			continue
		}
		trimmed = append(trimmed, HotspotExplanation{
			Path:        path,
			Explanation: explanation,
		})
	}
	return trimmed
}

func trimReadingPathExplanations(items []ReadingPathExplanation) []ReadingPathExplanation {
	trimmed := make([]ReadingPathExplanation, 0, len(items))
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		rationale := strings.TrimSpace(item.Rationale)
		if path == "" || rationale == "" {
			continue
		}
		trimmed = append(trimmed, ReadingPathExplanation{
			Path:      path,
			Rationale: rationale,
		})
	}
	return trimmed
}
