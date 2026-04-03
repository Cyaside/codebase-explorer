package repo

import "time"

type ScanOptions struct {
	RootPath            string
	ExtraIgnorePatterns []string
}

type FileInfo struct {
	Path            string    `json:"path"`
	Extension       string    `json:"extension"`
	SizeBytes       int64     `json:"size_bytes"`
	LineCount       int       `json:"line_count"`
	ImportCount     int       `json:"import_count"`
	TodoCount       int       `json:"todo_count"`
	FixmeCount      int       `json:"fixme_count"`
	HackCount       int       `json:"hack_count"`
	IsDocumentation bool      `json:"is_documentation"`
	IsEntryPoint    bool      `json:"is_entry_point"`
	ScannedAt       time.Time `json:"scanned_at"`
}

type ScanResult struct {
	RootPath       string     `json:"root_path"`
	ProjectName    string     `json:"project_name"`
	ScannedAt      time.Time  `json:"scanned_at"`
	Files          []FileInfo `json:"files"`
	Directories    []string   `json:"directories"`
	IgnorePatterns []string   `json:"ignore_patterns"`
	ManifestFiles  []string   `json:"manifest_files"`
	Documentation  []string   `json:"documentation_files"`
}
