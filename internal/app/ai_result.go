package app

import (
	"context"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func (s Service) buildFinalAIResult(_ context.Context, request AnalyzeRequest, analysis analyzer.Result, _ provider.CondensedContext, execution fullai.Execution) (provider.Result, string, error) {
	emitAnalyzeProgress(request, "ai-synthesis", "skipped", "AI function outputs supply the result")
	return projectProviderResultFromFullAI(analysis.GeneratedAt, execution), cache.StatusDisabled, nil
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
		result.Status = provider.ResultStatusPartial
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
		if result.Name == name && result.Verified {
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
		return (provider.Config{
			Name:    request.ProviderOverride.Name,
			Model:   request.ProviderOverride.Model,
			APIKey:  request.ProviderOverride.APIKey,
			BaseURL: request.ProviderOverride.BaseURL,
		}).Canonical()
	}

	return (provider.Config{
		Name:    s.settings.Provider.Name,
		Model:   s.settings.Provider.Model,
		APIKey:  s.settings.Provider.APIKey,
		BaseURL: s.settings.Provider.BaseURL,
	}).Canonical()
}
