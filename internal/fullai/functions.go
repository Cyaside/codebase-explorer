package fullai

import "strings"

type FunctionPreparer struct{}

func NewFunctionPreparer() FunctionPreparer {
	return FunctionPreparer{}
}

func (FunctionPreparer) Prepare(input Input, plan Plan, evidence Evidence) Functions {
	options := Options{
		Mode:        NormalizeMode(plan.Mode),
		ReadBudget:  plan.ReadBudget,
		TokenBudget: plan.TokenBudget,
	}.Normalize()
	if !options.Mode.Enabled() {
		return DisabledFunctions(input.Analysis.GeneratedAt, options, "full-ai mode not requested")
	}

	jobs := make([]FunctionJob, 0, len(defaultFunctionTasks()))
	for _, task := range defaultFunctionTasks() {
		evidencePaths := selectFunctionEvidence(task.Name, evidence.Items)
		jobs = append(jobs, FunctionJob{
			Name:            task.Name,
			Objective:       task.Objective,
			Status:          "prepared",
			InstructionPath: instructionPath(task.Name),
			EvidencePaths:   evidencePaths,
			EvidenceCount:   len(evidencePaths),
			Focus:           focusAreasForFunction(task.Name),
			Note:            buildFunctionNote(len(evidencePaths)),
		})
	}

	return Functions{
		SchemaVersion: FunctionsSchemaVersion,
		GeneratedAt:   input.Analysis.GeneratedAt,
		Mode:          string(options.Mode),
		Jobs:          jobs,
		Note:          "prepared function-scoped work items from collected evidence",
	}
}

func selectFunctionEvidence(name string, items []EvidenceItem) []string {
	paths := make([]string, 0, 6)
	seen := map[string]struct{}{}
	add := func(item EvidenceItem) {
		path := strings.TrimSpace(item.DisplayPath)
		if path == "" {
			path = strings.TrimSpace(item.ResolvedPath)
		}
		if path == "" {
			return
		}
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	for _, item := range items {
		if matchesFunctionEvidence(name, item) {
			add(item)
		}
	}
	if len(paths) == 0 {
		for _, item := range items {
			add(item)
			if len(paths) >= 3 {
				break
			}
		}
	}
	return paths
}

func matchesFunctionEvidence(name string, item EvidenceItem) bool {
	switch name {
	case "summary":
		return item.Source == "reading-path" || item.Source == "module"
	case "architecture":
		return item.Source == "module" || item.Source == "dependency-risk"
	case "hotspots-and-dependencies":
		return item.Source == "hotspot" || item.Source == "dependency-risk" || item.Source == "issue-correlation"
	case "flowchart":
		return item.Source == "reading-path" || item.Resolution == "exact-file"
	case "issues":
		return item.Source == "support-file" || item.Source == "issue-correlation"
	case "recommendations":
		return item.Source == "reading-path" || item.Source == "hotspot" || item.Source == "support-file"
	case "dashboard":
		return true
	default:
		return false
	}
}

func focusAreasForFunction(name string) []string {
	switch name {
	case "summary":
		return []string{"project identity", "primary code surface", "entry points"}
	case "architecture":
		return []string{"module boundaries", "responsibility split", "integration edges"}
	case "hotspots-and-dependencies":
		return []string{"risky concentration", "import pressure", "likely unstable areas"}
	case "flowchart":
		return []string{"request flow", "entry path", "handoff sequence"}
	case "issues":
		return []string{"support-file signals", "code correlation", "uncertainty"}
	case "recommendations":
		return []string{"onboarding order", "investigation path", "next questions"}
	case "dashboard":
		return []string{"headline metrics", "status summary", "actionable overview"}
	default:
		return nil
	}
}

func buildFunctionNote(evidenceCount int) string {
	if evidenceCount == 0 {
		return "no direct evidence was mapped to this function yet"
	}
	return "prepared from collected full-ai evidence"
}
