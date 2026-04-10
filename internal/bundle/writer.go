package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
	"github.com/Cyaside/codebase-explorer/internal/report"
	"github.com/Cyaside/codebase-explorer/internal/viewer"
)

type WriteRequest struct {
	OutputRoot        string
	DeterministicOnly bool
	ScanResult        repo.ScanResult
	Analysis          analyzer.Result
	Changes           changes.Result
	Warnings          []string
	Cache             CacheMeta
	AIContext         provider.CondensedContext
	AIResult          provider.Result
	FullAIPlan        fullai.Plan
	FullAIEvidence    fullai.Evidence
	FullAIFunctions   fullai.Functions
	FullAIExecution   fullai.Execution
	FullAISummary     fullai.Summary
}

type WriteResult struct {
	BundlePath     string
	PrunedBundles  int
	RetentionLimit int
}

type Writer struct {
	version    string
	keepLatest int
}

type CacheMeta struct {
	Enabled             bool
	Root                string
	DeterministicStatus string
	ProviderStatus      string
}

func NewWriter(version string, keepLatest int) Writer {
	return Writer{
		version:    version,
		keepLatest: keepLatest,
	}
}

func (w Writer) Write(request WriteRequest) (WriteResult, error) {
	bundleName, err := nextBundleName(request.OutputRoot, request.Analysis.ProjectName, time.Now().UTC())
	if err != nil {
		return WriteResult{}, fmt.Errorf("allocate bundle name: %w", err)
	}
	bundlePath := filepath.Join(request.OutputRoot, bundleName)

	directories := []string{
		bundlePath,
		filepath.Join(bundlePath, "overview"),
		filepath.Join(bundlePath, "architecture"),
		filepath.Join(bundlePath, "hotspots"),
		filepath.Join(bundlePath, "dependencies"),
		filepath.Join(bundlePath, "reading-path"),
		filepath.Join(bundlePath, "changes"),
		filepath.Join(bundlePath, "data"),
		filepath.Join(bundlePath, "ui"),
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return WriteResult{}, fmt.Errorf("create bundle directory %q: %w", directory, err)
		}
	}

	files := map[string]string{
		filepath.Join(bundlePath, "README.md"):                            report.RootREADME(request.ScanResult, request.Analysis, request.AIResult, request.Warnings, bundleName),
		filepath.Join(bundlePath, "overview", "README.md"):                report.OverviewREADME(request.Analysis, request.AIResult),
		filepath.Join(bundlePath, "architecture", "README.md"):            report.ArchitectureREADME(request.Analysis, request.AIResult),
		filepath.Join(bundlePath, "architecture", "module-graph.mmd"):     report.ArchitectureMermaid(request.Analysis),
		filepath.Join(bundlePath, "hotspots", "README.md"):                report.HotspotsREADME(request.Analysis, request.AIResult, request.Changes),
		filepath.Join(bundlePath, "dependencies", "README.md"):            report.DependenciesREADME(request.Analysis),
		filepath.Join(bundlePath, "dependencies", "dependency-graph.mmd"): report.DependenciesMermaid(request.Analysis),
		filepath.Join(bundlePath, "reading-path", "README.md"):            report.ReadingPathREADME(request.Analysis, request.AIResult),
		filepath.Join(bundlePath, "changes", "README.md"):                 report.ChangesREADME(request.Changes),
		filepath.Join(bundlePath, "data", "README.md"):                    report.DataREADME(),
	}

	for pathOnDisk, contents := range files {
		if err := os.WriteFile(pathOnDisk, []byte(contents), 0o644); err != nil {
			return WriteResult{}, fmt.Errorf("write report file %q: %w", pathOnDisk, err)
		}
	}

	viewerFiles, err := viewer.Files(buildViewerData(request, bundleName))
	if err != nil {
		return WriteResult{}, fmt.Errorf("prepare viewer files: %w", err)
	}
	for relativePath, contents := range viewerFiles {
		pathOnDisk := filepath.Join(bundlePath, "ui", relativePath)
		if err := os.WriteFile(pathOnDisk, contents, 0o644); err != nil {
			return WriteResult{}, fmt.Errorf("write viewer file %q: %w", pathOnDisk, err)
		}
	}

	if err := writeJSON(filepath.Join(bundlePath, "data", "analysis.json"), request.Analysis); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "metrics.json"), request.Analysis.Metrics); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "files.json"), request.ScanResult.Files); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "modules.json"), request.Analysis.Modules); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "ai-context.json"), request.AIContext); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "ai-result.json"), request.AIResult); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "full-ai-plan.json"), request.FullAIPlan); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "full-ai-meta.json"), request.FullAISummary); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "full-ai-evidence.json"), request.FullAIEvidence); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "full-ai-functions.json"), request.FullAIFunctions); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "full-ai-execution.json"), request.FullAIExecution); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "changes", "issue-correlation.json"), request.Changes); err != nil {
		return WriteResult{}, err
	}
	contractMeta := map[string]any{
		"bundle_schema_version":            request.Analysis.SchemaVersion,
		"analysis_schema_version":          request.Analysis.SchemaVersion,
		"metrics_schema_version":           request.Analysis.SchemaVersion,
		"files_schema_version":             request.Analysis.SchemaVersion,
		"modules_schema_version":           request.Analysis.SchemaVersion,
		"changes_schema_version":           request.Changes.SchemaVersion,
		"ai_context_schema_version":        request.AIContext.SchemaVersion,
		"ai_result_schema_version":         request.AIResult.SchemaVersion,
		"full_ai_plan_schema_version":      request.FullAIPlan.SchemaVersion,
		"full_ai_evidence_schema_version":  request.FullAIEvidence.SchemaVersion,
		"full_ai_functions_schema_version": request.FullAIFunctions.SchemaVersion,
		"full_ai_execution_schema_version": request.FullAIExecution.SchemaVersion,
		"full_ai_meta_schema_version":      request.FullAISummary.SchemaVersion,
		"full_ai_mode":                     request.FullAISummary.Mode,
		"tool_version":                     w.version,
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "contract.json"), contractMeta); err != nil {
		return WriteResult{}, err
	}
	cacheMeta := map[string]any{
		"version":                w.version,
		"cache_enabled":          request.Cache.Enabled,
		"cache_root":             request.Cache.Root,
		"deterministic_status":   request.Cache.DeterministicStatus,
		"provider_status":        request.Cache.ProviderStatus,
		"deterministic_only":     request.DeterministicOnly,
		"full_ai_mode":           request.FullAISummary.Mode,
		"generated_at":           request.Analysis.GeneratedAt,
		"output_retention_limit": w.keepLatest,
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "cache-meta.json"), cacheMeta); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "hotspots", "hotspots.json"), request.Analysis.Hotspots); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "dependencies", "dependencies.json"), request.Analysis.DependencyRisks); err != nil {
		return WriteResult{}, err
	}
	if err := writeJSON(filepath.Join(bundlePath, "reading-path", "reading-path.json"), request.Analysis.ReadingPath); err != nil {
		return WriteResult{}, err
	}

	prunedBundles, err := pruneBundles(request.OutputRoot, bundlePath, w.keepLatest)
	if err != nil {
		return WriteResult{}, err
	}

	return WriteResult{
		BundlePath:     bundlePath,
		PrunedBundles:  prunedBundles,
		RetentionLimit: w.keepLatest,
	}, nil
}

func writeJSON(pathOnDisk string, value any) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json for %q: %w", pathOnDisk, err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(pathOnDisk, contents, 0o644); err != nil {
		return fmt.Errorf("write json file %q: %w", pathOnDisk, err)
	}
	return nil
}

var unsafeNameCharacters = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func sanitizeName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "repository"
	}
	cleaned := unsafeNameCharacters.ReplaceAllString(trimmed, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		return "repository"
	}
	return strings.ToLower(cleaned)
}

func buildViewerData(request WriteRequest, bundleName string) viewer.BundleData {
	return viewer.BundleData{
		BundleName:  bundleName,
		GeneratedAt: request.Analysis.GeneratedAt,
		Project: viewer.ProjectData{
			Name:         request.Analysis.ProjectName,
			Type:         request.Analysis.ProjectType,
			AnalyzedPath: request.Analysis.AnalyzedPath,
			Summary:      request.Analysis.Summary,
			ProviderMode: request.Analysis.Provider,
		},
		Metrics:              request.Analysis.Metrics,
		Warnings:             request.Warnings,
		Languages:            request.Analysis.Languages,
		ImportantDirectories: request.Analysis.ImportantDirectories,
		EntryPoints:          request.Analysis.EntryPoints,
		CoreModules:          request.Analysis.CoreModules,
		Modules:              request.Analysis.Modules,
		Hotspots:             request.Analysis.Hotspots,
		Dependencies:         request.Analysis.DependencyRisks,
		ReadingPath:          request.Analysis.ReadingPath,
		Changes:              request.Changes,
		AI:                   request.AIResult,
		FullAI:               request.FullAISummary,
		FullAIExecution:      request.FullAIExecution,
		Mermaid: viewer.MermaidData{
			Architecture: report.ArchitectureMermaid(request.Analysis),
			Dependencies: report.DependenciesMermaid(request.Analysis),
		},
		Links: viewer.LinkData{
			Root:                "README.md",
			Overview:            "overview/README.md",
			Architecture:        "architecture/README.md",
			Dependencies:        "dependencies/README.md",
			Hotspots:            "hotspots/README.md",
			ReadingPath:         "reading-path/README.md",
			Changes:             "changes/README.md",
			ArchitectureDiagram: "architecture/module-graph.mmd",
			DependencyDiagram:   "dependencies/dependency-graph.mmd",
		},
	}
}
