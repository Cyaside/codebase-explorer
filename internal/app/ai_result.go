package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func (s Service) buildFinalAIResult(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, aiContext provider.CondensedContext, execution fullai.Execution) (provider.Result, string, error) {
	if request.FullAI.Normalize().Mode.Enabled() {
		emitAnalyzeProgress(request, "ai-synthesis", "skipped", "full-ai function outputs supersede legacy synthesis")
		return projectProviderResultFromFullAI(analysis.GeneratedAt, execution), cache.StatusDisabled, nil
	}

	return s.buildAIResult(ctx, request, analysis, request.DeterministicOnly, aiContext)
}

func (s Service) buildAIResult(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, deterministicOnly bool, aiContext provider.CondensedContext) (provider.Result, string, error) {
	if deterministicOnly {
		emitAnalyzeProgress(request, "ai-synthesis", "skipped", "deterministic-only mode enabled")
		return provider.SkippedResult(analysis.GeneratedAt, "deterministic-only mode enabled"), cache.StatusDisabled, nil
	}

	providerConfig := s.providerConfig(request)
	if !providerConfig.Enabled() {
		emitAnalyzeProgress(request, "ai-synthesis", "disabled", "no provider configured")
		return provider.DisabledResult(analysis.GeneratedAt, "no provider configured"), cache.StatusDisabled, nil
	}

	if err := s.providers.Validate(providerConfig); err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled, nil
	}

	client, err := s.providers.ClientFor(providerConfig)
	if err != nil {
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled, nil
	}

	if s.cache.Enabled() {
		cacheKey, keyErr := cache.BuildProviderKey(providerConfig, aiContext, s.settings.AppVersion)
		if keyErr == nil {
			cached, loadErr := s.cache.LoadProvider(cacheKey)
			switch {
			case loadErr == nil:
				emitAnalyzeProgress(request, "provider-cache", cache.StatusHit, "reused cached provider synthesis")
				return cached.Result, cache.StatusHit, nil
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
				if errors.Is(synthErr, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
					emitAnalyzeProgress(request, "ai-synthesis", "canceled", "provider call canceled")
					return provider.Result{}, cache.StatusMiss, context.Canceled
				}
				emitAnalyzeProgress(request, "ai-synthesis", "fallback", synthErr.Error())
				return provider.FallbackResult(analysis.GeneratedAt, providerConfig, synthErr.Error()), cache.StatusMiss, nil
			}

			_ = s.cache.SaveProvider(cache.ProviderPayload{
				CachedAt:   time.Now().UTC(),
				Key:        cacheKey,
				Result:     result,
				AppVersion: s.settings.AppVersion,
			})
			emitAnalyzeProgress(request, "ai-synthesis", "succeeded", fmt.Sprintf("provider %s returned structured synthesis", result.Provider))
			return result, cache.StatusMiss, nil
		}
	}

	emitAnalyzeProgress(request, "ai-synthesis", "running", fmt.Sprintf("calling provider %s with model %s", providerConfig.Name, providerConfig.Model))
	result, err := client.Synthesize(ctx, provider.Request{
		Config:  providerConfig,
		Context: aiContext,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			emitAnalyzeProgress(request, "ai-synthesis", "canceled", "provider call canceled")
			return provider.Result{}, cache.StatusDisabled, context.Canceled
		}
		emitAnalyzeProgress(request, "ai-synthesis", "fallback", err.Error())
		return provider.FallbackResult(analysis.GeneratedAt, providerConfig, err.Error()), cache.StatusDisabled, nil
	}

	emitAnalyzeProgress(request, "ai-synthesis", "succeeded", fmt.Sprintf("provider %s returned structured synthesis", result.Provider))
	return result, cache.StatusDisabled, nil
}

func projectProviderResultFromFullAI(generatedAt time.Time, execution fullai.Execution) provider.Result {
	if execution.ExecutedCount == 0 {
		return emptyFullAIProviderResult(generatedAt, execution)
	}

	summary := functionOutput(execution, "summary")
	architecture := functionOutput(execution, "architecture")
	hotspots := functionOutput(execution, "hotspots-and-dependencies")
	recommendations := functionOutput(execution, "recommendations")

	result := provider.Result{
		SchemaVersion:           provider.ResultSchemaVersion,
		GeneratedAt:             generatedAt,
		Provider:                strings.TrimSpace(execution.Provider),
		Model:                   strings.TrimSpace(execution.Model),
		Status:                  provider.ResultStatusAvailable,
		Used:                    true,
		ProjectSummary:          summary.Summary,
		ArchitectureNarrative:   architecture.Summary,
		HotspotExplanations:     hotspotExplanationsFromFullAI(hotspots),
		ReadingPathExplanations: readingPathExplanationsFromFullAI(recommendations),
	}
	if result.ProjectSummary == "" {
		result.ProjectSummary = firstFindingClaim(summary)
	}
	if result.ArchitectureNarrative == "" {
		result.ArchitectureNarrative = firstFindingClaim(architecture)
	}
	if execution.Status == "partial" || execution.Status == "unverified" {
		result.FallbackReason = strings.TrimSpace(execution.Note)
	}
	return result
}

func emptyFullAIProviderResult(generatedAt time.Time, execution fullai.Execution) provider.Result {
	reason := strings.TrimSpace(execution.Note)
	if reason == "" {
		reason = "full-ai mode produced no executable function output"
	}

	switch execution.Status {
	case "disabled":
		return provider.DisabledResult(generatedAt, reason)
	case "skipped":
		return provider.SkippedResult(generatedAt, reason)
	default:
		return provider.FallbackResult(generatedAt, provider.Config{
			Name:  execution.Provider,
			Model: execution.Model,
		}, reason)
	}
}

func functionOutput(execution fullai.Execution, name string) fullai.FunctionOutput {
	for _, result := range execution.Results {
		if result.Name == name && (result.Status == "verified" || result.Status == "unverified") {
			return result.Output
		}
	}
	return fullai.FunctionOutput{}
}

func hotspotExplanationsFromFullAI(output fullai.FunctionOutput) []provider.HotspotExplanation {
	explanations := make([]provider.HotspotExplanation, 0, len(output.KeyFindings))
	for _, finding := range output.KeyFindings {
		path := firstEvidencePath(finding.EvidencePaths)
		if path == "" {
			continue
		}
		explanations = append(explanations, provider.HotspotExplanation{
			Path:        path,
			Explanation: finding.Claim,
		})
	}
	return explanations
}

func readingPathExplanationsFromFullAI(output fullai.FunctionOutput) []provider.ReadingPathExplanation {
	explanations := make([]provider.ReadingPathExplanation, 0, len(output.KeyFindings))
	for _, finding := range output.KeyFindings {
		path := firstEvidencePath(finding.EvidencePaths)
		if path == "" {
			continue
		}
		explanations = append(explanations, provider.ReadingPathExplanation{
			Path:      path,
			Rationale: finding.Claim,
		})
	}
	return explanations
}

func firstFindingClaim(output fullai.FunctionOutput) string {
	if len(output.KeyFindings) == 0 {
		return ""
	}
	return strings.TrimSpace(output.KeyFindings[0].Claim)
}

func firstEvidencePath(paths []string) string {
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
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
