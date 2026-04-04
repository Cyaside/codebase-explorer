package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/viewer"
)

const viewerDataPrefix = "window.CODEARCH_VIEWER_DATA = "

type workbenchBundleSummary struct {
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	GeneratedAt      time.Time `json:"generated_at"`
	ProjectName      string    `json:"project_name"`
	ProjectType      string    `json:"project_type"`
	AnalyzedPath     string    `json:"analyzed_path"`
	TotalFiles       int       `json:"total_files"`
	TotalLines       int       `json:"total_lines"`
	AIStatus         string    `json:"ai_status"`
	WarningCount     int       `json:"warning_count"`
	SupportFileCount int       `json:"support_file_count"`
}

type workbenchBundle struct {
	Summary workbenchBundleSummary `json:"summary"`
	Data    viewer.BundleData      `json:"data"`
}

type workbenchBundleLocation struct {
	name    string
	path    string
	modTime time.Time
}

func listWorkbenchBundles(outputRoot string, limit int) ([]workbenchBundleSummary, error) {
	locations, err := listBundleLocations(outputRoot)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(locations) > limit {
		locations = locations[:limit]
	}

	summaries := make([]workbenchBundleSummary, 0, len(locations))
	for _, location := range locations {
		data, err := loadViewerBundleData(location.path)
		if err != nil {
			continue
		}
		summaries = append(summaries, summarizeWorkbenchBundle(location, data))
	}

	return summaries, nil
}

func loadWorkbenchBundle(outputRoot, bundleName string) (workbenchBundle, error) {
	bundleName = strings.TrimSpace(bundleName)
	if bundleName == "" || bundleName == "." || bundleName == ".." || filepath.Base(bundleName) != bundleName {
		return workbenchBundle{}, fmt.Errorf("bundle %q is not valid", bundleName)
	}

	bundlePath := filepath.Join(outputRoot, filepath.Clean(bundleName))
	absolutePath, err := filepath.Abs(bundlePath)
	if err != nil {
		return workbenchBundle{}, fmt.Errorf("resolve bundle path %q: %w", bundleName, err)
	}

	if err := validateBundlePath(absolutePath); err != nil {
		return workbenchBundle{}, err
	}

	data, err := loadViewerBundleData(absolutePath)
	if err != nil {
		return workbenchBundle{}, err
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return workbenchBundle{}, fmt.Errorf("inspect bundle path %q: %w", absolutePath, err)
	}

	location := workbenchBundleLocation{
		name:    filepath.Base(absolutePath),
		path:    absolutePath,
		modTime: info.ModTime().UTC(),
	}

	return workbenchBundle{
		Summary: summarizeWorkbenchBundle(location, data),
		Data:    data,
	}, nil
}

func listBundleLocations(outputRoot string) ([]workbenchBundleLocation, error) {
	entries, err := os.ReadDir(outputRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list output root %q: %w", outputRoot, err)
	}

	locations := make([]workbenchBundleLocation, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := filepath.Join(outputRoot, entry.Name())
		if !looksLikeBundle(candidatePath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect bundle %q: %w", candidatePath, err)
		}

		locations = append(locations, workbenchBundleLocation{
			name:    entry.Name(),
			path:    candidatePath,
			modTime: info.ModTime().UTC(),
		})
	}

	slices.SortFunc(locations, func(left, right workbenchBundleLocation) int {
		if !left.modTime.Equal(right.modTime) {
			if left.modTime.After(right.modTime) {
				return -1
			}
			return 1
		}
		return strings.Compare(right.name, left.name)
	})

	return locations, nil
}

func loadViewerBundleData(bundlePath string) (viewer.BundleData, error) {
	contents, err := os.ReadFile(filepath.Join(bundlePath, "ui", "viewer-data.js"))
	if err != nil {
		return viewer.BundleData{}, fmt.Errorf("read viewer payload for %q: %w", bundlePath, err)
	}

	payload := strings.TrimSpace(string(contents))
	payload = strings.TrimPrefix(payload, viewerDataPrefix)
	payload = strings.TrimSuffix(payload, ";")
	payload = strings.TrimSpace(payload)

	var data viewer.BundleData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return viewer.BundleData{}, fmt.Errorf("decode viewer payload for %q: %w", bundlePath, err)
	}

	return data, nil
}

func summarizeWorkbenchBundle(location workbenchBundleLocation, data viewer.BundleData) workbenchBundleSummary {
	generatedAt := data.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = location.modTime
	}

	return workbenchBundleSummary{
		Name:             location.name,
		Path:             location.path,
		GeneratedAt:      generatedAt,
		ProjectName:      data.Project.Name,
		ProjectType:      data.Project.Type,
		AnalyzedPath:     data.Project.AnalyzedPath,
		TotalFiles:       data.Metrics.TotalFiles,
		TotalLines:       data.Metrics.TotalLines,
		AIStatus:         data.AI.Status,
		WarningCount:     len(data.Warnings),
		SupportFileCount: data.Changes.SupportFileCount,
	}
}
