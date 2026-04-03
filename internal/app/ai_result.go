package app

import (
	"context"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func (s Service) buildAIResult(ctx context.Context, analysis analyzer.Result, deterministicOnly bool, aiContext provider.CondensedContext) provider.Result {
	if deterministicOnly {
		return provider.SkippedResult(analysis.GeneratedAt, "deterministic-only mode enabled")
	}

	providerConfig := s.providerConfig()
	if !providerConfig.Enabled() {
		return provider.DisabledResult(analysis.GeneratedAt, "no provider configured")
	}

	if err := s.providers.Validate(providerConfig); err != nil {
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

	client, err := s.providers.ClientFor(providerConfig)
	if err != nil {
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

	result, err := client.Synthesize(ctx, provider.Request{
		Config:  providerConfig,
		Context: aiContext,
	})
	if err != nil {
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error())
	}

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
