package fullai

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

const maxEvidenceSnippetBytes = 6000

type Collector struct{}

func NewCollector() Collector {
	return Collector{}
}

func (Collector) Collect(input Input, plan Plan) Evidence {
	options := Options{
		Mode:        NormalizeMode(plan.Mode),
		ReadBudget:  plan.ReadBudget,
		TokenBudget: plan.TokenBudget,
	}.Normalize()

	if !options.Mode.Enabled() {
		return DisabledEvidence(input.Analysis.GeneratedAt, input.RootPath, options, "full-ai mode not requested")
	}

	items := make([]EvidenceItem, 0, len(plan.Targets))
	failedItems := 0
	for _, target := range plan.Targets {
		item := collectEvidenceItem(input, target)
		if item.ReadStatus != "read" {
			failedItems++
		}
		items = append(items, item)
	}

	note := fmt.Sprintf("collected %d evidence item(s) from planned full-ai targets", len(items))
	if failedItems > 0 {
		note = fmt.Sprintf("%s; %d item(s) could not be read directly", note, failedItems)
	}

	return Evidence{
		SchemaVersion:  EvidenceSchemaVersion,
		GeneratedAt:    input.Analysis.GeneratedAt,
		Mode:           string(options.Mode),
		RootPath:       strings.TrimSpace(input.RootPath),
		ReadBudget:     options.ReadBudget,
		CollectedItems: len(items),
		FailedItems:    failedItems,
		Items:          items,
		Note:           note,
	}
}

func collectEvidenceItem(input Input, target Target) EvidenceItem {
	item := EvidenceItem{
		TargetPath: target.Path,
		Reason:     target.Reason,
		Source:     target.Source,
		Priority:   target.Priority,
		ReadStatus: "missing",
	}

	resolvedPath, displayPath, resolution, lineCount, err := resolveEvidencePath(input, target)
	item.ResolvedPath = resolvedPath
	item.DisplayPath = displayPath
	item.Resolution = resolution
	item.LineCount = lineCount
	if err != nil {
		return item
	}

	contents, err := os.ReadFile(resolvedPath)
	if err != nil {
		return item
	}
	item.ByteCount = len(contents)

	if !utf8.Valid(contents) {
		item.ReadStatus = "binary-skipped"
		return item
	}

	if len(contents) > maxEvidenceSnippetBytes {
		item.Snippet = string(contents[:maxEvidenceSnippetBytes])
		item.Truncated = true
	} else {
		item.Snippet = string(contents)
	}
	item.ReadStatus = "read"
	if item.LineCount == 0 {
		item.LineCount = lineCountFromContent(item.Snippet)
	}
	return item
}

func resolveEvidencePath(input Input, target Target) (resolvedPath string, displayPath string, resolution string, lineCount int, err error) {
	cleanTarget := filepath.Clean(strings.TrimSpace(target.Path))
	if cleanTarget == "." || cleanTarget == "" {
		return "", "", "", 0, fmt.Errorf("empty target path")
	}

	if filepath.IsAbs(cleanTarget) {
		return cleanTarget, cleanTarget, "support-file", 0, nil
	}

	files := append([]repo.FileInfo(nil), input.ScanResult.Files...)
	normalizedTarget := filepath.ToSlash(cleanTarget)
	for _, file := range files {
		if filepath.ToSlash(file.Path) == normalizedTarget {
			return filepath.Join(input.RootPath, filepath.FromSlash(file.Path)), file.Path, "exact-file", file.LineCount, nil
		}
	}

	prefix := normalizedTarget + "/"
	candidates := make([]repo.FileInfo, 0, 4)
	for _, file := range files {
		if strings.HasPrefix(filepath.ToSlash(file.Path), prefix) {
			candidates = append(candidates, file)
		}
	}
	if len(candidates) == 0 {
		return "", "", "", 0, fmt.Errorf("no matching file")
	}

	slices.SortFunc(candidates, func(left repo.FileInfo, right repo.FileInfo) int {
		if left.IsEntryPoint != right.IsEntryPoint {
			if left.IsEntryPoint {
				return -1
			}
			return 1
		}
		if left.LineCount != right.LineCount {
			if left.LineCount > right.LineCount {
				return -1
			}
			return 1
		}
		return strings.Compare(left.Path, right.Path)
	})

	selected := candidates[0]
	return filepath.Join(input.RootPath, filepath.FromSlash(selected.Path)), selected.Path, "module-representative", selected.LineCount, nil
}

func lineCountFromContent(contents string) int {
	if strings.TrimSpace(contents) == "" {
		return 0
	}
	return strings.Count(contents, "\n") + 1
}
