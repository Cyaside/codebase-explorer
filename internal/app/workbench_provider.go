package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/provider"
)

const providerDiagnosticsPrompt = "Reply with exactly: Codebase Explorer provider test ok"

type workbenchProviderDiagnosticPayload struct {
	Provider provider.Config `json:"provider"`
}

type workbenchProviderModelsResponse struct {
	Provider  string   `json:"provider"`
	BaseURL   string   `json:"base_url"`
	Count     int      `json:"count"`
	Models    []string `json:"models"`
	LatencyMS int64    `json:"latency_ms"`
}

type workbenchProviderTestResponse struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	LatencyMS int64  `json:"latency_ms"`
}

func (s Service) handleWorkbenchProviderModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := decodeWorkbenchProviderDiagnosticPayload(r)
	if err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateProviderForModelList(s.providers, payload.Provider); err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}

	startedAt := time.Now()
	models, err := fetchProviderModels(r.Context(), payload.Provider)
	if err != nil {
		writeWorkbenchError(w, http.StatusBadGateway, err)
		return
	}

	writeWorkbenchJSON(w, http.StatusOK, workbenchProviderModelsResponse{
		Provider:  strings.TrimSpace(payload.Provider.Name),
		BaseURL:   providerBaseURL(payload.Provider),
		Count:     len(models),
		Models:    models,
		LatencyMS: time.Since(startedAt).Milliseconds(),
	})
}

func (s Service) handleWorkbenchProviderTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := decodeWorkbenchProviderDiagnosticPayload(r)
	if err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.providers.Validate(payload.Provider); err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}

	client, err := s.providers.ClientFor(payload.Provider)
	if err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}

	startedAt := time.Now()
	result, err := client.Complete(r.Context(), provider.PromptRequest{
		Config:       payload.Provider,
		SystemPrompt: "You are a strict diagnostics responder.",
		UserPrompt:   providerDiagnosticsPrompt,
	})
	if err != nil {
		writeWorkbenchError(w, http.StatusBadGateway, err)
		return
	}

	writeWorkbenchJSON(w, http.StatusOK, workbenchProviderTestResponse{
		Provider:  result.Provider,
		Model:     result.Model,
		Status:    result.Status,
		Output:    result.Content,
		LatencyMS: time.Since(startedAt).Milliseconds(),
	})
}

func decodeWorkbenchProviderDiagnosticPayload(r *http.Request) (workbenchProviderDiagnosticPayload, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return workbenchProviderDiagnosticPayload{}, err
	}

	var payload workbenchProviderDiagnosticPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return workbenchProviderDiagnosticPayload{}, fmt.Errorf("decode request: %w", err)
	}
	return payload, nil
}

func validateProviderForModelList(registry provider.Registry, config provider.Config) error {
	if !config.Enabled() {
		return fmt.Errorf("provider mode is required")
	}
	descriptor, found := registry.Describe(config.Name)
	if !found {
		return registry.Validate(config)
	}
	if descriptor.RequiresAPIKey && strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("provider %q requires an API key", descriptor.Name)
	}
	if descriptor.RequiresBaseURL && strings.TrimSpace(config.BaseURL) == "" {
		return fmt.Errorf("provider %q requires a base URL", descriptor.Name)
	}
	return nil
}

func fetchProviderModels(ctx context.Context, config provider.Config) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, providerBaseURL(config)+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("build model list request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(config.APIKey))
	request.Header.Set("Accept", "application/json")
	setWorkbenchOpenRouterHeaders(request, config)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send model list request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read model list response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("model list request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, fmt.Errorf("decode model list response: %w", err)
	}

	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		models = append(models, id)
	}
	return models, nil
}

func providerBaseURL(config provider.Config) string {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" && strings.EqualFold(strings.TrimSpace(config.Name), "openai") {
		return "https://api.openai.com/v1"
	}
	return baseURL
}

func setWorkbenchOpenRouterHeaders(request *http.Request, config provider.Config) {
	if !strings.Contains(strings.ToLower(strings.TrimSpace(config.BaseURL)), "openrouter.ai") {
		return
	}
	request.Header.Set("HTTP-Referer", "http://localhost")
	request.Header.Set("X-Title", "Codebase Explorer")
}
