package changes

import "time"

const (
	ResultSchemaVersion = "changes.v1"

	SourceStatusParsed = "parsed"
	SourceStatusEmpty  = "empty"
	SourceStatusFailed = "failed"

	ConfidenceStrong    = "strong"
	ConfidenceModerate  = "moderate"
	ConfidenceAmbiguous = "ambiguous"
)

type Result struct {
	SchemaVersion            string               `json:"schema_version"`
	GeneratedAt              time.Time            `json:"generated_at"`
	Available                bool                 `json:"available"`
	SupportFileCount         int                  `json:"support_file_count"`
	ParsedItemCount          int                  `json:"parsed_item_count"`
	Sources                  []Source             `json:"sources"`
	RepeatedThemes           []Theme              `json:"repeated_themes,omitempty"`
	FrequentlyMentionedAreas []AreaMention        `json:"frequently_mentioned_areas,omitempty"`
	LikelyUnstableModules    []AreaMention        `json:"likely_unstable_modules,omitempty"`
	HotspotCorrelations      []HotspotCorrelation `json:"hotspot_correlations,omitempty"`
	Note                     string               `json:"note,omitempty"`
}

type Source struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Format    string `json:"format"`
	Status    string `json:"status"`
	ItemCount int    `json:"item_count"`
	Message   string `json:"message,omitempty"`
}

type Theme struct {
	Name         string   `json:"name"`
	MentionCount int      `json:"mention_count"`
	SourceCount  int      `json:"source_count"`
	RelatedAreas []string `json:"related_areas,omitempty"`
}

type AreaMention struct {
	Path           string   `json:"path"`
	MentionCount   int      `json:"mention_count"`
	SourceCount    int      `json:"source_count"`
	Confidence     string   `json:"confidence"`
	RelatedHotspot bool     `json:"related_hotspot"`
	Reasons        []string `json:"reasons,omitempty"`
	Examples       []string `json:"examples,omitempty"`
}

type HotspotCorrelation struct {
	Path         string `json:"path"`
	MentionCount int    `json:"mention_count"`
	Confidence   string `json:"confidence"`
}
