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
	"github.com/Cyaside/codebase-explorer/internal/repo"
	"github.com/Cyaside/codebase-explorer/internal/report"
)

type WriteRequest struct {
	OutputRoot        string
	DeterministicOnly bool
	ScanResult        repo.ScanResult
	Analysis          analyzer.Result
}

type WriteResult struct {
	BundlePath string
}

type Writer struct {
	version string
}

func NewWriter(version string) Writer {
	return Writer{version: version}
}

func (w Writer) Write(request WriteRequest) (WriteResult, error) {
	bundleName := fmt.Sprintf("%s_%s", time.Now().UTC().Format("2006-01-02_150405"), sanitizeName(request.Analysis.ProjectName))
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
	}
	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return WriteResult{}, fmt.Errorf("create bundle directory %q: %w", directory, err)
		}
	}

	files := map[string]string{
		filepath.Join(bundlePath, "README.md"):                 report.RootREADME(request.ScanResult, request.Analysis, bundleName),
		filepath.Join(bundlePath, "overview", "README.md"):     report.OverviewREADME(request.Analysis),
		filepath.Join(bundlePath, "architecture", "README.md"): report.ArchitectureREADME(request.Analysis),
		filepath.Join(bundlePath, "hotspots", "README.md"):     report.HotspotsREADME(request.Analysis),
		filepath.Join(bundlePath, "dependencies", "README.md"): report.DependenciesREADME(request.Analysis),
		filepath.Join(bundlePath, "reading-path", "README.md"): report.ReadingPathREADME(request.Analysis),
		filepath.Join(bundlePath, "changes", "README.md"):      report.ChangesREADME(),
		filepath.Join(bundlePath, "data", "README.md"):         report.DataREADME(),
	}

	for pathOnDisk, contents := range files {
		if err := os.WriteFile(pathOnDisk, []byte(contents), 0o644); err != nil {
			return WriteResult{}, fmt.Errorf("write report file %q: %w", pathOnDisk, err)
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
	contractMeta := map[string]any{
		"bundle_schema_version":   request.Analysis.SchemaVersion,
		"analysis_schema_version": request.Analysis.SchemaVersion,
		"metrics_schema_version":  request.Analysis.SchemaVersion,
		"files_schema_version":    request.Analysis.SchemaVersion,
		"modules_schema_version":  request.Analysis.SchemaVersion,
		"tool_version":            w.version,
	}
	if err := writeJSON(filepath.Join(bundlePath, "data", "contract.json"), contractMeta); err != nil {
		return WriteResult{}, err
	}
	cacheMeta := map[string]any{
		"version":            w.version,
		"cache_enabled":      false,
		"deterministic_only": request.DeterministicOnly,
		"generated_at":       request.Analysis.GeneratedAt,
		"note":               "phase 1 baseline writes cache metadata but does not reuse cache yet",
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

	return WriteResult{BundlePath: bundlePath}, nil
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
