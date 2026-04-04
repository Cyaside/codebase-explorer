package report

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

var mermaidUnsafeCharacters = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func ArchitectureMermaid(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("flowchart TD\n")

	projectID := mermaidNodeID("project", analysis.ProjectName)
	builder.WriteString(fmt.Sprintf("  %s[%q]\n", projectID, architectureProjectLabel(analysis)))

	for index, entryPoint := range limitedStrings(analysis.EntryPoints, 4) {
		entryID := mermaidNodeID("entry", fmt.Sprintf("%d_%s", index, entryPoint))
		builder.WriteString(fmt.Sprintf("  %s[%q]\n", entryID, "Entry: "+entryPoint))
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", projectID, entryID))
	}

	coreModules := analysis.CoreModules
	if len(coreModules) == 0 {
		coreModules = modulePaths(analysis.Modules, 4)
	}

	moduleIDs := map[string]string{}
	for index, modulePath := range limitedStrings(coreModules, 4) {
		moduleID := mermaidNodeID("module", fmt.Sprintf("%d_%s", index, modulePath))
		moduleIDs[modulePath] = moduleID
		builder.WriteString(fmt.Sprintf("  %s[%q]\n", moduleID, architectureModuleLabel(modulePath, analysis.Modules)))
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", projectID, moduleID))
	}

	for index, hotspot := range limitedHotspots(analysis.Hotspots, 5) {
		hotspotID := mermaidNodeID("hotspot", fmt.Sprintf("%d_%s", index, hotspot.Path))
		builder.WriteString(fmt.Sprintf("  %s[%q]\n", hotspotID, hotspotLabel(hotspot)))

		if moduleID, found := nearestModuleID(hotspot.Path, moduleIDs); found {
			builder.WriteString(fmt.Sprintf("  %s --> %s\n", moduleID, hotspotID))
			continue
		}
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", projectID, hotspotID))
	}

	return builder.String()
}

func DependenciesMermaid(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("flowchart LR\n")

	rootID := mermaidNodeID("dependencies", analysis.ProjectName)
	builder.WriteString(fmt.Sprintf("  %s[%q]\n", rootID, "Dependency concentration review"))

	moduleIDs := map[string]string{}
	for index, risk := range limitedDependencyRisks(analysis.DependencyRisks, 6) {
		modulePath := topLevelSegment(risk.Path)
		moduleID, found := moduleIDs[modulePath]
		if !found {
			moduleID = mermaidNodeID("module", fmt.Sprintf("%d_%s", index, modulePath))
			moduleIDs[modulePath] = moduleID
			builder.WriteString(fmt.Sprintf("  %s[%q]\n", moduleID, "Module: "+modulePath))
			builder.WriteString(fmt.Sprintf("  %s --> %s\n", rootID, moduleID))
		}

		riskID := mermaidNodeID("risk", fmt.Sprintf("%d_%s", index, risk.Path))
		builder.WriteString(fmt.Sprintf("  %s[%q]\n", riskID, dependencyRiskLabel(risk)))
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", moduleID, riskID))
	}

	if len(analysis.DependencyRisks) == 0 {
		emptyID := mermaidNodeID("dependencies", "none")
		builder.WriteString(fmt.Sprintf("  %s[%q]\n", emptyID, "No concentrated dependency hotspots detected"))
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", rootID, emptyID))
	}

	return builder.String()
}

func architectureProjectLabel(analysis analyzer.Result) string {
	name := strings.TrimSpace(analysis.ProjectName)
	projectType := strings.TrimSpace(analysis.ProjectType)
	if projectType == "" {
		return name
	}
	return name + "<br/>" + projectType
}

func architectureModuleLabel(modulePath string, modules []analyzer.ModuleInfo) string {
	for _, module := range modules {
		if module.Path != modulePath {
			continue
		}
		return fmt.Sprintf("Module: %s<br/>%d files / %d lines", modulePath, module.FileCount, module.TotalLines)
	}
	return "Module: " + modulePath
}

func hotspotLabel(hotspot analyzer.Hotspot) string {
	return fmt.Sprintf("Hotspot: %s<br/>score %.2f", hotspot.Path, hotspot.Score)
}

func dependencyRiskLabel(risk analyzer.DependencyRisk) string {
	return fmt.Sprintf("%s<br/>imports %d", risk.Path, risk.ImportCount)
}

func modulePaths(modules []analyzer.ModuleInfo, limit int) []string {
	paths := make([]string, 0, len(modules))
	for _, module := range modules {
		paths = append(paths, module.Path)
	}
	return limitedStrings(paths, limit)
}

func limitedStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitedHotspots(values []analyzer.Hotspot, limit int) []analyzer.Hotspot {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitedDependencyRisks(values []analyzer.DependencyRisk, limit int) []analyzer.DependencyRisk {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func nearestModuleID(pathValue string, moduleIDs map[string]string) (string, bool) {
	var bestModule string
	for modulePath := range moduleIDs {
		if modulePath == "." {
			continue
		}
		if strings.HasPrefix(pathValue, modulePath+"/") {
			if len(modulePath) > len(bestModule) {
				bestModule = modulePath
			}
		}
	}
	if bestModule == "" {
		return "", false
	}
	return moduleIDs[bestModule], true
}

func topLevelSegment(pathValue string) string {
	cleaned := path.Clean(strings.TrimSpace(pathValue))
	if cleaned == "." || cleaned == "" {
		return "."
	}
	if !strings.Contains(cleaned, "/") {
		return cleaned
	}
	return strings.SplitN(cleaned, "/", 2)[0]
}

func mermaidNodeID(prefix, value string) string {
	normalized := mermaidUnsafeCharacters.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "node"
	}
	return prefix + "_" + normalized
}
