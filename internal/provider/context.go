package provider

import "time"

type CondensedContext struct {
	SchemaVersion        string                `json:"schema_version"`
	GeneratedAt          time.Time             `json:"generated_at"`
	Project              ProjectFacts          `json:"project"`
	Modules              []ModuleSummary       `json:"modules"`
	EntryPoints          []string              `json:"entry_points"`
	Hotspots             []HotspotSummary      `json:"hotspots"`
	DependencyHighlights []DependencyHighlight `json:"dependency_highlights"`
	ReadingPath          []ReadingPathHint     `json:"reading_path"`
	Metadata             ContextMetadata       `json:"metadata"`
}

type ProjectFacts struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Summary         string `json:"summary"`
	PrimaryLanguage string `json:"primary_language"`
	TotalFiles      int    `json:"total_files"`
	TotalLines      int    `json:"total_lines"`
}

type ModuleSummary struct {
	Path            string   `json:"path"`
	FileCount       int      `json:"file_count"`
	TotalLines      int      `json:"total_lines"`
	Languages       []string `json:"languages"`
	EntryPointCount int      `json:"entry_point_count"`
	MarkerCount     int      `json:"marker_count"`
}

type HotspotSummary struct {
	Path    string   `json:"path"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

type DependencyHighlight struct {
	Path        string `json:"path"`
	ImportCount int    `json:"import_count"`
	Reason      string `json:"reason"`
}

type ReadingPathHint struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type ContextMetadata struct {
	ModulesIncluded      int  `json:"modules_included"`
	ModulesTrimmed       int  `json:"modules_trimmed"`
	HotspotsIncluded     int  `json:"hotspots_included"`
	HotspotsTrimmed      int  `json:"hotspots_trimmed"`
	DependenciesIncluded int  `json:"dependencies_included"`
	DependenciesTrimmed  int  `json:"dependencies_trimmed"`
	ReadingPathIncluded  int  `json:"reading_path_included"`
	ReadingPathTrimmed   int  `json:"reading_path_trimmed"`
	Truncated            bool `json:"truncated"`
}
