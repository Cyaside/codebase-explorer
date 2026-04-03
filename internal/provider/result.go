package provider

import (
	"context"
	"strings"
	"time"
)

const (
	ResultSchemaVersion   = "ai-result.v1"
	ResultStatusDisabled  = "disabled"
	ResultStatusSkipped   = "skipped"
	ResultStatusAvailable = "succeeded"
	ResultStatusFallback  = "fallback"
)

type Client interface {
	Synthesize(context.Context, Request) (Result, error)
}

type Request struct {
	Config  Config
	Context CondensedContext
}

type Result struct {
	SchemaVersion           string                   `json:"schema_version"`
	GeneratedAt             time.Time                `json:"generated_at"`
	Provider                string                   `json:"provider"`
	Model                   string                   `json:"model"`
	Status                  string                   `json:"status"`
	Used                    bool                     `json:"used"`
	FallbackReason          string                   `json:"fallback_reason,omitempty"`
	ProjectSummary          string                   `json:"project_summary,omitempty"`
	ArchitectureNarrative   string                   `json:"architecture_narrative,omitempty"`
	HotspotExplanations     []HotspotExplanation     `json:"hotspot_explanations,omitempty"`
	ReadingPathExplanations []ReadingPathExplanation `json:"reading_path_explanations,omitempty"`
}

type HotspotExplanation struct {
	Path        string `json:"path"`
	Explanation string `json:"explanation"`
}

type ReadingPathExplanation struct {
	Path      string `json:"path"`
	Rationale string `json:"rationale"`
}

func DisabledResult(generatedAt time.Time, reason string) Result {
	return Result{
		SchemaVersion:  ResultSchemaVersion,
		GeneratedAt:    generatedAt,
		Status:         ResultStatusDisabled,
		Used:           false,
		FallbackReason: strings.TrimSpace(reason),
	}
}

func SkippedResult(generatedAt time.Time, reason string) Result {
	return Result{
		SchemaVersion:  ResultSchemaVersion,
		GeneratedAt:    generatedAt,
		Status:         ResultStatusSkipped,
		Used:           false,
		FallbackReason: strings.TrimSpace(reason),
	}
}

func FallbackResult(generatedAt time.Time, config Config, reason string) Result {
	return Result{
		SchemaVersion:  ResultSchemaVersion,
		GeneratedAt:    generatedAt,
		Provider:       strings.TrimSpace(config.Name),
		Model:          strings.TrimSpace(config.Model),
		Status:         ResultStatusFallback,
		Used:           false,
		FallbackReason: strings.TrimSpace(reason),
	}
}

func (r Result) Successful() bool {
	return r.Status == ResultStatusAvailable && r.Used
}
