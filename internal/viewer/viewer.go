package viewer

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

//go:embed assets/*
var embeddedAssets embed.FS

type BundleData struct {
	BundleName           string                     `json:"bundle_name"`
	GeneratedAt          time.Time                  `json:"generated_at"`
	Project              ProjectData                `json:"project"`
	Metrics              analyzer.Metrics           `json:"metrics"`
	Warnings             []string                   `json:"warnings"`
	Languages            []analyzer.LanguageSummary `json:"languages"`
	ImportantDirectories []string                   `json:"important_directories"`
	EntryPoints          []string                   `json:"entry_points"`
	CoreModules          []string                   `json:"core_modules"`
	Modules              []analyzer.ModuleInfo      `json:"modules"`
	Hotspots             []analyzer.Hotspot         `json:"hotspots"`
	Dependencies         []analyzer.DependencyRisk  `json:"dependencies"`
	ReadingPath          []analyzer.ReadingPathItem `json:"reading_path"`
	Changes              changes.Result             `json:"changes"`
	AI                   provider.Result            `json:"ai"`
	FullAI               fullai.Summary             `json:"full_ai"`
	FullAIExecution      fullai.Execution           `json:"full_ai_execution"`
	Mermaid              MermaidData                `json:"mermaid"`
	Links                LinkData                   `json:"links"`
}

type ProjectData struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	AnalyzedPath string `json:"analyzed_path"`
	Summary      string `json:"summary"`
	ProviderMode string `json:"provider_mode"`
}

type MermaidData struct {
	Architecture string `json:"architecture"`
	Dependencies string `json:"dependencies"`
}

type LinkData struct {
	Root                string `json:"root"`
	Overview            string `json:"overview"`
	Architecture        string `json:"architecture"`
	Dependencies        string `json:"dependencies"`
	Hotspots            string `json:"hotspots"`
	ReadingPath         string `json:"reading_path"`
	Changes             string `json:"changes"`
	ArchitectureDiagram string `json:"architecture_diagram"`
	DependencyDiagram   string `json:"dependency_diagram"`
}

func Files(data BundleData) (map[string][]byte, error) {
	files := map[string][]byte{}

	assetFiles := []string{
		"assets/index.html",
		"assets/app.js",
		"assets/style.css",
	}
	for _, assetPath := range assetFiles {
		contents, err := fs.ReadFile(embeddedAssets, assetPath)
		if err != nil {
			return nil, fmt.Errorf("read viewer asset %q: %w", assetPath, err)
		}
		files[trimAssetPrefix(assetPath)] = contents
	}

	viewerData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal viewer data: %w", err)
	}
	files["viewer-data.js"] = append([]byte("window.CODEARCH_VIEWER_DATA = "), append(viewerData, []byte(";\n")...)...)

	return files, nil
}

func trimAssetPrefix(path string) string {
	return path[len("assets/"):]
}
