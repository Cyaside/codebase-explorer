package app

import (
	"context"
	"fmt"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func (s Service) buildAIResult(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, deterministicOnly bool, aiContext provider.CondensedContext) provider.Result {
	if deterministicOnly {
		emitAnalyzeProgress(request, "ai-synthesis", "skipped", "deterministic-only mode enabled")
		return provider.SkippedResult(analysis.GeneratedAt, "deterministic-only mode enabled")
	}

	providerConfig := s.providerConfig()
	if !providerConfig.Enabled() {
		emitAnalyzeProgress(request, "ai-synthesis", "disabled", "no provider configured")
		return provider.DisabledResult(analysis.GeneratedAt, "no provider configured")
	}

	if err := s.providers.Validate(providerConfig); err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

	client, err := s.providers.ClientFor(providerConfig)
	if err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

	emitAnalyzeProgress(request, "ai-synthesis", "running", fmt.Sprintf("calling provider %s with model %s", providerConfig.Name, providerConfig.Model))
	result, err := client.Synthesize(ctx, provider.Request{
		Config:  providerConfig,
		Context: aiContext,
	})
	if err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

	emitAnalyzeProgress(request, "ai-synthesis", "succeeded", fmt.Sprintf("provider %s returned structured synthesis", result.Provider))
	return result
}

func (s Service) providerConfig() provider.Config {
	return provider.Config{
		Name:    s.settings.Provider.Name,
		Model:   s.settings.Provider.Model,
		APIKey:  s.settings.Provider.APIKey,
		BaseURL: s.settings.Provider.BaseURL,
	}
}
