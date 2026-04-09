package fullai

import (
	"fmt"
	"slices"
	"strings"
)

type Planner struct{}

func NewPlanner() Planner {
	return Planner{}
}

func (Planner) Plan(input Input, options Options) (Plan, Summary) {
	normalized := options.Normalize()
	if !normalized.Mode.Enabled() {
		disabledPlan := DisabledPlan(input.Analysis.GeneratedAt, normalized, "full-ai mode not requested")
		return disabledPlan, DisabledSummary(normalized, "full-ai mode not requested")
	}

	targets, truncated := buildTargets(input, normalized.ReadBudget)
	functions := defaultFunctionTasks()
	note := "planner scaffold recorded evidence targets for future deep execution"
	if truncated {
		note = fmt.Sprintf("%s; target list was truncated to the read budget of %d", note, normalized.ReadBudget)
	}

	plan := Plan{
		SchemaVersion: PlanSchemaVersion,
		GeneratedAt:   input.Analysis.GeneratedAt,
		Mode:          string(normalized.Mode),
		ReadBudget:    normalized.ReadBudget,
		TokenBudget:   normalized.TokenBudget,
		Goals:         fullAIGoals(),
		Targets:       targets,
		Functions:     functions,
		Truncated:     truncated,
		Note:          note,
	}

	summary := Summary{
		SchemaVersion:    SummarySchemaVersion,
		Enabled:          true,
		Mode:             string(normalized.Mode),
		Status:           "planned",
		ReadBudget:       normalized.ReadBudget,
		TokenBudget:      normalized.TokenBudget,
		PlannedTargets:   len(targets),
		PlannedFunctions: len(functions),
		Note:             note,
	}

	return plan, summary
}

func buildTargets(input Input, readBudget int) ([]Target, bool) {
	var targets []Target
	seen := map[string]struct{}{}
	addTarget := func(path string, reason string, source string, priority int) {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath == "" {
			return
		}
		if _, exists := seen[trimmedPath]; exists {
			return
		}
		seen[trimmedPath] = struct{}{}
		targets = append(targets, Target{
			Path:     trimmedPath,
			Reason:   strings.TrimSpace(reason),
			Source:   source,
			Priority: priority,
		})
	}

	for index, item := range input.Analysis.ReadingPath {
		addTarget(item.Path, item.Reason, "reading-path", 100-index)
	}
	for index, hotspot := range input.Analysis.Hotspots {
		addTarget(hotspot.Path, strings.Join(hotspot.Reasons, "; "), "hotspot", 90-index)
	}
	for index, risk := range input.Analysis.DependencyRisks {
		addTarget(risk.Path, risk.Reason, "dependency-risk", 80-index)
	}
	for index, area := range input.Changes.FrequentlyMentionedAreas {
		addTarget(area.Path, strings.Join(area.Reasons, "; "), "issue-correlation", 85-index)
	}
	for _, path := range input.SupportFiles {
		target := supportFileTarget(path)
		addTarget(target.Path, target.Reason, target.Source, target.Priority)
	}
	for index, module := range input.Analysis.Modules {
		reason := fmt.Sprintf("module spans %d file(s) and %d total line(s)", module.FileCount, module.TotalLines)
		addTarget(module.Path, reason, "module", 70-index)
	}

	slices.SortFunc(targets, func(left Target, right Target) int {
		if left.Priority == right.Priority {
			return strings.Compare(left.Path, right.Path)
		}
		if left.Priority > right.Priority {
			return -1
		}
		return 1
	})

	if readBudget <= 0 || len(targets) <= readBudget {
		return targets, false
	}

	return targets[:readBudget], true
}
