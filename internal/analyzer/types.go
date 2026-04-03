package analyzer

import (
	"time"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type Result struct {
	Version              string            `json:"version"`
	GeneratedAt          time.Time         `json:"generated_at"`
	ProjectName          string            `json:"project_name"`
	AnalyzedPath         string            `json:"analyzed_path"`
	ProjectType          string            `json:"project_type"`
	Summary              string            `json:"summary"`
	Provider             string            `json:"provider"`
	Languages            []LanguageSummary `json:"languages"`
	ImportantDirectories []string          `json:"important_directories"`
	EntryPoints          []string          `json:"entry_points"`
	CoreModules          []string          `json:"core_modules"`
	Hotspots             []Hotspot         `json:"hotspots"`
	DependencyRisks      []DependencyRisk  `json:"dependency_risks"`
	ReadingPath          []ReadingPathItem `json:"reading_path"`
	Modules              []ModuleInfo      `json:"modules"`
	Metrics              Metrics           `json:"metrics"`
	Files                []repo.FileInfo   `json:"files"`
}

type LanguageSummary struct {
	Name      string `json:"name"`
	FileCount int    `json:"file_count"`
	LineCount int    `json:"line_count"`
}

type ModuleInfo struct {
	Path            string   `json:"path"`
	FileCount       int      `json:"file_count"`
	TotalLines      int      `json:"total_lines"`
	Languages       []string `json:"languages"`
	EntryPointCount int      `json:"entry_point_count"`
	MarkerCount     int      `json:"marker_count"`
}

type Hotspot struct {
	Path        string   `json:"path"`
	Score       float64  `json:"score"`
	Reasons     []string `json:"reasons"`
	LineCount   int      `json:"line_count"`
	ImportCount int      `json:"import_count"`
	MarkerCount int      `json:"marker_count"`
}

type DependencyRisk struct {
	Path        string `json:"path"`
	ImportCount int    `json:"import_count"`
	Reason      string `json:"reason"`
}

type ReadingPathItem struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type Metrics struct {
	TotalFiles       int `json:"total_files"`
	TotalDirectories int `json:"total_directories"`
	TotalLines       int `json:"total_lines"`
	TodoCount        int `json:"todo_count"`
	FixmeCount       int `json:"fixme_count"`
	HackCount        int `json:"hack_count"`
}
