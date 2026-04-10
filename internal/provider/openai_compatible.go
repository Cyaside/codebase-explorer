package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

const maxPromptCompletionAttempts = 3

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
	systemPrompt, userPrompt, err := buildSynthesisPrompt(request.Context)
	if err != nil {
		return Result{}, err
	}
	promptResult, err := c.Complete(ctx, PromptRequest{
		Config:       request.Config,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return Result{}, err
	}
	result, err := parseSynthesisResult(promptResult.Content)
	if err != nil {
		return Result{}, err
	}

	result.SchemaVersion = ResultSchemaVersion
	result.GeneratedAt = promptResult.GeneratedAt
	result.Provider = promptResult.Provider
	result.Model = promptResult.Model
	result.Status = ResultStatusAvailable
	result.Used = true
	result.FallbackReason = ""

	return normalizeResultText(result), nil
}

func (c OpenAICompatibleClient) Complete(ctx context.Context, request PromptRequest) (PromptResult, error) {
	var lastErr error
	for attempt := range maxPromptCompletionAttempts {
		result, retryAfter, err := c.completeOnce(ctx, request)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if attempt == maxPromptCompletionAttempts-1 || !retryablePromptError(err) {
			return PromptResult{}, err
		}
		if err := sleepForPromptRetry(ctx, promptRetryDelay(attempt, retryAfter)); err != nil {
			return PromptResult{}, err
		}
	}

	return PromptResult{}, lastErr
}

func (c OpenAICompatibleClient) completeOnce(ctx context.Context, request PromptRequest) (PromptResult, *time.Duration, error) {
	payload, err := marshalPromptCompletionRequest(request)
	if err != nil {
		return PromptResult{}, nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsURL(request.Config), bytes.NewReader(payload))
	if err != nil {
		return PromptResult{}, nil, fmt.Errorf("build prompt request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+strings.TrimSpace(request.Config.APIKey))
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	setOpenRouterHeaders(httpRequest, request.Config)

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return PromptResult{}, nil, fmt.Errorf("send prompt request: %w", err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(httpResponse.Body, 2<<20))
	if err != nil {
		return PromptResult{}, nil, fmt.Errorf("read prompt response: %w", err)
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return PromptResult{}, retryAfterDuration(httpResponse.Header.Get("Retry-After")), httpStatusError(httpResponse.StatusCode, responseBody)
	}

	var response openAICompatibleChatResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return PromptResult{}, nil, fmt.Errorf("decode prompt response: %w", err)
	}

	content, err := response.messageContent()
	if err != nil {
		return PromptResult{}, nil, err
	}

	return PromptResult{
		GeneratedAt: time.Now().UTC(),
		Provider:    normalizeName(request.Config.Name),
		Model:       strings.TrimSpace(request.Config.Model),
		Status:      ResultStatusAvailable,
		Used:        true,
		Content:     strings.TrimSpace(content),
	}, nil, nil
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

func marshalPromptCompletionRequest(request PromptRequest) ([]byte, error) {
	payload := openAICompatibleChatRequest{
		Model: strings.TrimSpace(request.Config.Model),
		Messages: []openAICompatibleChatMessage{
			{Role: "system", Content: strings.TrimSpace(request.SystemPrompt)},
			{Role: "user", Content: strings.TrimSpace(request.UserPrompt)},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal prompt request: %w", err)
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

type promptStatusError struct {
	statusCode int
	body       []byte
}

func (e promptStatusError) Error() string {
	message := strings.TrimSpace(string(e.body))
	if message == "" {
		return fmt.Sprintf("synthesis request failed with status %d", e.statusCode)
	}
	return fmt.Sprintf("synthesis request failed with status %d: %s", e.statusCode, message)
}

func httpStatusError(statusCode int, body []byte) error {
	return promptStatusError{statusCode: statusCode, body: body}
}

func retryablePromptError(err error) bool {
	var statusErr promptStatusError
	if errors.As(err, &statusErr) {
		return statusErr.statusCode == http.StatusTooManyRequests || statusErr.statusCode >= http.StatusInternalServerError
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "did not include any choices") || strings.Contains(message, "returned an empty message")
}

func promptRetryDelay(attempt int, retryAfter *time.Duration) time.Duration {
	if retryAfter != nil {
		return *retryAfter
	}
	switch attempt {
	case 0:
		return 2 * time.Second
	default:
		return 6 * time.Second
	}
}

func sleepForPromptRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryAfterDuration(value string) *time.Duration {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if seconds, err := strconv.Atoi(trimmed); err == nil {
		duration := time.Duration(seconds) * time.Second
		return &duration
	}
	if retryAt, err := http.ParseTime(trimmed); err == nil {
		duration := time.Until(retryAt)
		if duration < 0 {
			duration = 0
		}
		return &duration
	}
	return nil
}

func setOpenRouterHeaders(request *http.Request, config Config) {
	if !strings.Contains(strings.ToLower(strings.TrimSpace(config.BaseURL)), "openrouter.ai") {
		return
	}
	request.Header.Set("HTTP-Referer", "http://localhost")
	request.Header.Set("X-Title", "Codebase Explorer")
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
