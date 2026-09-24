package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

const maxPromptCompletionAttempts = 3
const promptHTTPTimeout = 180 * time.Second

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type OpenAICompatibleClient struct {
	httpClient HTTPDoer
	limiter    *promptLimiter
}

type promptLimiter struct {
	serial atomic.Bool
	mu     sync.Mutex
}

func NewOpenAICompatibleClient(httpClient HTTPDoer) OpenAICompatibleClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: promptHTTPTimeout}
	}
	return OpenAICompatibleClient{httpClient: httpClient, limiter: &promptLimiter{}}
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
		var result PromptResult
		var retryAfter *time.Duration
		var err error
		if c.limiter != nil && c.limiter.serial.Load() {
			c.limiter.mu.Lock()
			result, retryAfter, err = c.completeOnce(ctx, request)
			c.limiter.mu.Unlock()
		} else {
			result, retryAfter, err = c.completeOnce(ctx, request)
		}
		if err == nil {
			return result, nil
		}
		var statusErr promptStatusError
		if c.limiter != nil && errors.As(err, &statusErr) && statusErr.statusCode == http.StatusTooManyRequests {
			c.limiter.serial.Store(true)
		}
		lastErr = err
		if ctx.Err() != nil {
			return PromptResult{}, ctx.Err()
		}
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
		GeneratedAt:  time.Now().UTC(),
		Provider:     request.Config.Canonical().Name,
		Model:        strings.TrimSpace(request.Config.Model),
		Status:       ResultStatusAvailable,
		Used:         true,
		Content:      strings.TrimSpace(content),
		PromptTokens: response.Usage.PromptTokens,
		OutputTokens: response.Usage.CompletionTokens,
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
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
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
}

func (e promptStatusError) Error() string {
	return fmt.Sprintf("provider request failed with HTTP status %d", e.statusCode)
}

func httpStatusError(statusCode int, _ []byte) error {
	return promptStatusError{statusCode: statusCode}
}

func retryablePromptError(err error) bool {
	var statusErr promptStatusError
	if errors.As(err, &statusErr) {
		return statusErr.statusCode == http.StatusTooManyRequests || statusErr.statusCode >= http.StatusInternalServerError
	}
	var networkError net.Error
	if errors.As(err, &networkError) {
		return networkError.Timeout() || networkError.Temporary()
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
		duration := min(time.Duration(seconds)*time.Second, 30*time.Second)
		return &duration
	}
	if retryAt, err := http.ParseTime(trimmed); err == nil {
		duration := min(time.Until(retryAt), 30*time.Second)
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
