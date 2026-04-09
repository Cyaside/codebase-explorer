package fullai

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

const (
	PlanSchemaVersion     = "full-ai-plan.v1"
	EvidenceSchemaVersion = "full-ai-evidence.v1"
	SummarySchemaVersion  = "full-ai-summary.v1"

	DefaultReadBudget  = 24
	DefaultTokenBudget = 32000
)

type Mode string

const (
	ModeStandard Mode = "standard"
	ModeFull     Mode = "full-ai"
)

type Options struct {
	Mode        Mode `json:"mode"`
	ReadBudget  int  `json:"read_budget"`
	TokenBudget int  `json:"token_budget"`
}

type Input struct {
	RootPath     string
	ScanResult   repo.ScanResult
	Analysis     analyzer.Result
	Changes      changes.Result
	SupportFiles []string
}

type Plan struct {
	SchemaVersion string         `json:"schema_version"`
	GeneratedAt   time.Time      `json:"generated_at"`
	Mode          string         `json:"mode"`
	ReadBudget    int            `json:"read_budget"`
	TokenBudget   int            `json:"token_budget"`
	Goals         []string       `json:"goals"`
	Targets       []Target       `json:"targets"`
	Functions     []FunctionTask `json:"functions"`
	Truncated     bool           `json:"truncated"`
	Note          string         `json:"note,omitempty"`
}

type Summary struct {
	SchemaVersion    string `json:"schema_version"`
	Enabled          bool   `json:"enabled"`
	Mode             string `json:"mode"`
	Status           string `json:"status"`
	ReadBudget       int    `json:"read_budget"`
	TokenBudget      int    `json:"token_budget"`
	PlannedTargets   int    `json:"planned_targets"`
	PlannedFunctions int    `json:"planned_functions"`
	CollectedItems   int    `json:"collected_items"`
	FailedItems      int    `json:"failed_items"`
	Note             string `json:"note,omitempty"`
}

type Evidence struct {
	SchemaVersion  string         `json:"schema_version"`
	GeneratedAt    time.Time      `json:"generated_at"`
	Mode           string         `json:"mode"`
	RootPath       string         `json:"root_path"`
	ReadBudget     int            `json:"read_budget"`
	CollectedItems int            `json:"collected_items"`
	FailedItems    int            `json:"failed_items"`
	Items          []EvidenceItem `json:"items"`
	Note           string         `json:"note,omitempty"`
}

type EvidenceItem struct {
	TargetPath   string `json:"target_path"`
	ResolvedPath string `json:"resolved_path"`
	DisplayPath  string `json:"display_path"`
	Reason       string `json:"reason"`
	Source       string `json:"source"`
	Priority     int    `json:"priority"`
	Resolution   string `json:"resolution"`
	ReadStatus   string `json:"read_status"`
	ByteCount    int    `json:"byte_count"`
	LineCount    int    `json:"line_count"`
	Snippet      string `json:"snippet,omitempty"`
	Truncated    bool   `json:"truncated"`
}

type Target struct {
	Path     string `json:"path"`
	Reason   string `json:"reason"`
	Source   string `json:"source"`
	Priority int    `json:"priority"`
}

type FunctionTask struct {
	Name      string `json:"name"`
	Objective string `json:"objective"`
	Priority  int    `json:"priority"`
	Status    string `json:"status"`
}

func NormalizeMode(raw string) Mode {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", string(ModeStandard):
		return ModeStandard
	case string(ModeFull):
		return ModeFull
	default:
		return ModeStandard
	}
}

func (mode Mode) Enabled() bool {
	return mode == ModeFull
}

func (options Options) Normalize() Options {
	normalized := Options{
		Mode:        NormalizeMode(string(options.Mode)),
		ReadBudget:  options.ReadBudget,
		TokenBudget: options.TokenBudget,
	}

	if normalized.Mode.Enabled() {
		if normalized.ReadBudget <= 0 {
			normalized.ReadBudget = DefaultReadBudget
		}
		if normalized.TokenBudget <= 0 {
			normalized.TokenBudget = DefaultTokenBudget
		}
	}

	return normalized
}

func DisabledPlan(generatedAt time.Time, options Options, note string) Plan {
	normalized := options.Normalize()
	return Plan{
		SchemaVersion: PlanSchemaVersion,
		GeneratedAt:   generatedAt,
		Mode:          string(normalized.Mode),
		ReadBudget:    normalized.ReadBudget,
		TokenBudget:   normalized.TokenBudget,
		Goals:         fullAIGoals(),
		Functions:     defaultFunctionTasks(),
		Targets:       nil,
		Note:          strings.TrimSpace(note),
	}
}

func DisabledSummary(options Options, note string) Summary {
	normalized := options.Normalize()
	return Summary{
		SchemaVersion:    SummarySchemaVersion,
		Enabled:          false,
		Mode:             string(normalized.Mode),
		Status:           "disabled",
		ReadBudget:       normalized.ReadBudget,
		TokenBudget:      normalized.TokenBudget,
		PlannedTargets:   0,
		PlannedFunctions: len(defaultFunctionTasks()),
		Note:             strings.TrimSpace(note),
	}
}

func DisabledEvidence(generatedAt time.Time, rootPath string, options Options, note string) Evidence {
	normalized := options.Normalize()
	return Evidence{
		SchemaVersion: EvidenceSchemaVersion,
		GeneratedAt:   generatedAt,
		Mode:          string(normalized.Mode),
		RootPath:      strings.TrimSpace(rootPath),
		ReadBudget:    normalized.ReadBudget,
		Items:         nil,
		Note:          strings.TrimSpace(note),
	}
}

func supportFileTarget(path string) Target {
	baseName := filepath.Base(strings.TrimSpace(path))
	reason := "support file supplied for issue and change-aware exploration"
	if baseName != "" && baseName != "." && baseName != string(filepath.Separator) {
		reason = fmt.Sprintf("support file %s supplied for issue and change-aware exploration", baseName)
	}

	return Target{
		Path:     strings.TrimSpace(path),
		Reason:   reason,
		Source:   "support-file",
		Priority: 55,
	}
}

func fullAIGoals() []string {
	return []string{
		"inspect high-signal code and support files directly",
		"produce evidence-backed architecture, issue, and recommendation outputs",
		"record enough evidence targets to support future deeper AI execution",
	}
}

func defaultFunctionTasks() []FunctionTask {
	return []FunctionTask{
		{Name: "summary", Objective: "build a grounded project summary", Priority: 100, Status: "planned"},
		{Name: "architecture", Objective: "explain module boundaries and responsibilities", Priority: 95, Status: "planned"},
		{Name: "hotspots-and-dependencies", Objective: "review risky concentration areas and coupling", Priority: 90, Status: "planned"},
		{Name: "flowchart", Objective: "map the core runtime or request flows", Priority: 85, Status: "planned"},
		{Name: "issues", Objective: "correlate support-file issues with repository evidence", Priority: 80, Status: "planned"},
		{Name: "recommendations", Objective: "produce an onboarding and investigation path", Priority: 75, Status: "planned"},
		{Name: "dashboard", Objective: "assemble a dense reusable overview surface", Priority: 70, Status: "planned"},
	}
}
