package analyzer

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestAnalyzeDetectsLanguagesAndHotspots(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "sample-repo")
	scanResult, err := repo.NewScanner().Scan(t.Context(), repo.ScanOptions{RootPath: root})
	if err != nil {
		t.Fatalf("scan sample repo: %v", err)
	}

	result := NewService("test").Analyze(scanResult, true)

	if result.ProjectType == "" {
		t.Fatalf("expected project type to be detected")
	}
	if len(result.Languages) == 0 || result.Languages[0].Name != "Go" {
		t.Fatalf("expected Go to be detected as the primary language, got %#v", result.Languages)
	}
	if len(result.EntryPoints) == 0 {
		t.Fatalf("expected at least one entry point")
	}
	if len(result.Hotspots) == 0 {
		t.Fatalf("expected hotspot candidates")
	}
}

func TestBuildModulesGroupsRootFilesIntoRootBucket(t *testing.T) {
	t.Parallel()

	modules := buildModules([]repo.FileInfo{
		{Path: "README.md", LineCount: 10, Extension: ".md"},
		{Path: "go.mod", LineCount: 5, Extension: ".mod"},
		{Path: "internal/app/service.go", LineCount: 30, Extension: ".go"},
		{Path: "cmd/codearch/main.go", LineCount: 20, Extension: ".go", IsEntryPoint: true},
	})

	if len(modules) < 3 {
		t.Fatalf("expected root, internal, and cmd modules, got %#v", modules)
	}

	if modules[0].Path != "internal" {
		t.Fatalf("expected internal to stay as the dominant module, got %#v", modules[0])
	}

	var rootModule *ModuleInfo
	for index := range modules {
		if modules[index].Path == "." {
			rootModule = &modules[index]
			break
		}
	}
	if rootModule == nil {
		t.Fatalf("expected root files to be grouped into the root bucket, got %#v", modules)
	}
	if rootModule.FileCount != 2 {
		t.Fatalf("expected root bucket to contain two files, got %#v", rootModule)
	}
}

func TestBuildImportantDirectoriesSkipsRootBucket(t *testing.T) {
	t.Parallel()

	modules := []ModuleInfo{
		{Path: "internal", TotalLines: 200},
		{Path: ".", TotalLines: 100},
		{Path: "cmd", TotalLines: 50},
	}

	importantDirectories := buildImportantDirectories(modules)
	if len(importantDirectories) != 2 {
		t.Fatalf("expected only directory-like modules, got %#v", importantDirectories)
	}
	for _, directory := range importantDirectories {
		if directory == "." {
			t.Fatalf("did not expect the root bucket to appear in important directories: %#v", importantDirectories)
		}
	}
}

func TestAnalyzeKeepsFixturePathsOutOfOrientationSignals(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	scanResult, err := repo.NewScanner().Scan(t.Context(), repo.ScanOptions{RootPath: root})
	if err != nil {
		t.Fatalf("scan repository root: %v", err)
	}

	result := NewService("test").Analyze(scanResult, true)

	for _, entryPoint := range result.EntryPoints {
		if strings.Contains(entryPoint, "testdata/") {
			t.Fatalf("did not expect fixture entry point in orientation output: %#v", result.EntryPoints)
		}
	}
	for _, module := range result.CoreModules {
		if module == "testdata" {
			t.Fatalf("did not expect fixture module in core modules: %#v", result.CoreModules)
		}
	}
	for _, item := range result.ReadingPath {
		if strings.Contains(item.Path, "testdata/") || item.Path == "testdata" {
			t.Fatalf("did not expect fixture path in reading path: %#v", result.ReadingPath)
		}
	}
}

func TestComputeHotspotsRanksHighSignalFilesFirst(t *testing.T) {
	t.Parallel()

	hotspots := computeHotspots([]repo.FileInfo{
		{
			Path:         "internal/core/service.go",
			LineCount:    200,
			ImportCount:  5,
			TodoCount:    1,
			FixmeCount:   1,
			IsEntryPoint: true,
		},
		{
			Path:      "internal/core/helpers.go",
			LineCount: 40,
		},
	})

	if len(hotspots) != 2 {
		t.Fatalf("expected both files to be scored, got %#v", hotspots)
	}
	if hotspots[0].Path != "internal/core/service.go" {
		t.Fatalf("expected high-signal file to rank first, got %#v", hotspots)
	}
	if hotspots[0].Score <= hotspots[1].Score {
		t.Fatalf("expected first hotspot to outrank second, got %#v", hotspots)
	}
}
