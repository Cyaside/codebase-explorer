package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func (s Service) buildAIResult(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, deterministicOnly bool, aiContext provider.CondensedContext) (provider.Result, string) {
	if deterministicOnly {
		emitAnalyzeProgress(request, "ai-synthesis", "skipped", "deterministic-only mode enabled")
		return provider.SkippedResult(analysis.GeneratedAt, "deterministic-only mode enabled"), cache.StatusDisabled
	}

	providerConfig := s.providerConfig(request)
	if !providerConfig.Enabled() {
		emitAnalyzeProgress(request, "ai-synthesis", "disabled", "no provider configured")
		return provider.DisabledResult(analysis.GeneratedAt, "no provider configured"), cache.StatusDisabled
	}

	if err := s.providers.Validate(providerConfig); err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled
	}

	client, err := s.providers.ClientFor(providerConfig)
	if err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled
	}

	if s.cache.Enabled() {
		cacheKey, keyErr := cache.BuildProviderKey(providerConfig, aiContext, s.settings.AppVersion)
		if keyErr == nil {
			cached, loadErr := s.cache.LoadProvider(cacheKey)
			switch {
			case loadErr == nil:
				emitAnalyzeProgress(request, "provider-cache", cache.StatusHit, "reused cached provider synthesis")
				return cached.Result, cache.StatusHit
			case errors.Is(loadErr, cache.ErrCacheMiss):
				emitAnalyzeProgress(request, "provider-cache", cache.StatusMiss, "provider synthesis cache miss")
			default:
				emitAnalyzeProgress(request, "provider-cache", "bypass", "provider cache lookup failed; continuing with live synthesis")
			}

			emitAnalyzeProgress(request, "ai-synthesis", "running", fmt.Sprintf("calling provider %s with model %s", providerConfig.Name, providerConfig.Model))
			result, synthErr := client.Synthesize(ctx, provider.Request{
				Config:  providerConfig,
				Context: aiContext,
			})
			if synthErr != nil {
				emitAnalyzeProgress(request, "ai-synthesis", "fallback", synthErr.Error())
				return provider.FallbackResult(analysis.GeneratedAt, providerConfig, synthErr.Error()), cache.StatusMiss
			}

			_ = s.cache.SaveProvider(cache.ProviderPayload{
				CachedAt:   time.Now().UTC(),
				Key:        cacheKey,
				Result:     result,
				AppVersion: s.settings.AppVersion,
			})
			emitAnalyzeProgress(request, "ai-synthesis", "succeeded", fmt.Sprintf("provider %s returned structured synthesis", result.Provider))
			return result, cache.StatusMiss
		}
	}

	emitAnalyzeProgress(request, "ai-synthesis", "running", fmt.Sprintf("calling provider %s with model %s", providerConfig.Name, providerConfig.Model))
	result, err := client.Synthesize(ctx, provider.Request{
		Config:  providerConfig,
		Context: aiContext,
	})
	if err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled
	}

	emitAnalyzeProgress(request, "ai-synthesis", "succeeded", fmt.Sprintf("provider %s returned structured synthesis", result.Provider))
	return result, cache.StatusDisabled
}

func (s Service) providerConfig(request AnalyzeRequest) provider.Config {
	if request.ProviderOverride != nil {
		return provider.Config{
			Name:    request.ProviderOverride.Name,
			Model:   request.ProviderOverride.Model,
			APIKey:  request.ProviderOverride.APIKey,
			BaseURL: request.ProviderOverride.BaseURL,
		}
	}

	return provider.Config{
		Name:    s.settings.Provider.Name,
		Model:   s.settings.Provider.Model,
		APIKey:  s.settings.Provider.APIKey,
		BaseURL: s.settings.Provider.BaseURL,
	}
}
