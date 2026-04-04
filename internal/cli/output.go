package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/app"
)

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  codearch analyze <repo-path> [support-file ...] [--support <file>] [--issues <file>] [--changelog <file>] [--output <dir>] [--deterministic-only] [--ignore <pattern>]")
	fmt.Fprintln(output, "  codearch open [bundle-path] [--no-browser]")
	fmt.Fprintln(output, "  codearch export [bundle-path] [--output <zip-path>]")
	fmt.Fprintln(output, "  codearch doctor")
	fmt.Fprintln(output, "  codearch cache clear")
}

func printAnalyzeResult(output io.Writer, result app.AnalyzeResult) {
	fmt.Fprintf(output, "Analysis complete.\n")
	fmt.Fprintf(output, "Project: %s\n", result.ProjectName)
	fmt.Fprintf(output, "Type: %s\n", result.ProjectType)
	if result.PrimaryLanguage != "" {
		fmt.Fprintf(output, "Primary language: %s\n", result.PrimaryLanguage)
	}
	fmt.Fprintf(output, "Files analyzed: %d\n", result.TotalFiles)
	fmt.Fprintf(output, "Total lines: %d\n", result.TotalLines)
	if len(result.EntryPoints) > 0 {
		fmt.Fprintf(output, "Entry points: %s\n", strings.Join(result.EntryPoints, ", "))
	}
	if result.AI.Status != "" {
		fmt.Fprintf(output, "AI synthesis: %s\n", result.AI.Status)
	}
	if result.AI.Provider != "" {
		if result.AI.Model != "" {
			fmt.Fprintf(output, "AI provider: %s (%s)\n", result.AI.Provider, result.AI.Model)
		} else {
			fmt.Fprintf(output, "AI provider: %s\n", result.AI.Provider)
		}
	}
	if result.AI.ContextSummary != "" {
		fmt.Fprintf(output, "AI context: %s\n", result.AI.ContextSummary)
	}
	if result.AI.Note != "" {
		fmt.Fprintf(output, "AI note: %s\n", result.AI.Note)
	}
	if result.Changes.SupportFileCount > 0 {
		fmt.Fprintf(output, "Support files: %d input(s), %d parsed item(s)\n", result.Changes.SupportFileCount, result.Changes.ParsedItemCount)
		if result.Changes.MentionedAreas > 0 {
			fmt.Fprintf(output, "Change awareness: %d area(s) correlated\n", result.Changes.MentionedAreas)
		}
		if result.Changes.Note != "" {
			fmt.Fprintf(output, "Change note: %s\n", result.Changes.Note)
		}
	}
	if result.Cache.Enabled {
		fmt.Fprintf(output, "Cache: deterministic %s, provider %s\n", result.Cache.DeterministicStatus, result.Cache.ProviderStatus)
	}
	if result.Output.RetentionLimit > 0 {
		fmt.Fprintf(output, "Output retention: keep latest %d bundle(s)\n", result.Output.RetentionLimit)
	}
	if result.Output.PrunedBundles > 0 {
		fmt.Fprintf(output, "Output cleanup: removed %d older bundle(s)\n", result.Output.PrunedBundles)
	}
	fmt.Fprintf(output, "Bundle: %s\n", result.OutputPath)
}

func composeAnalyzeProgress(existing app.AnalyzeProgressReporter, output io.Writer) app.AnalyzeProgressReporter {
	return func(event app.AnalyzeProgressEvent) {
		if existing != nil {
			existing(event)
		}
		printAnalyzeProgress(output, event)
	}
}

func printAnalyzeProgress(output io.Writer, event app.AnalyzeProgressEvent) {
	if output == nil {
		return
	}

	stage := strings.TrimSpace(event.Stage)
	status := strings.TrimSpace(event.Status)
	detail := strings.TrimSpace(event.Detail)
	if stage == "" && status == "" && detail == "" {
		return
	}

	if detail == "" {
		fmt.Fprintf(output, "AI progress [%s/%s]\n", stage, status)
		return
	}
	fmt.Fprintf(output, "AI progress [%s/%s]: %s\n", stage, status, detail)
}

func printDoctorResult(output io.Writer, result app.DoctorResult) {
	fmt.Fprintf(output, "Config source: %s\n", result.ConfigSource)
	fmt.Fprintf(output, "Output root: %s\n", result.OutputRoot)
	for _, check := range result.Checks {
		fmt.Fprintf(output, "- [%s] %s: %s\n", strings.ToUpper(check.Status), check.Name, check.Detail)
	}
}

func printCacheClearResult(output io.Writer, result app.CacheClearResult) {
	fmt.Fprintf(output, "Cache cleared.\n")
	fmt.Fprintf(output, "Cache root: %s\n", result.CacheRoot)
	fmt.Fprintf(output, "Removed entries: %d\n", result.RemovedEntries)
}

func printOpenResult(output io.Writer, result app.OpenResult, openErr error, noBrowser bool) {
	fmt.Fprintf(output, "Viewer ready.\n")
	fmt.Fprintf(output, "Bundle: %s\n", result.BundlePath)
	fmt.Fprintf(output, "Viewer: %s\n", result.ViewerPath)
	if result.ResolvedLatest {
		fmt.Fprintf(output, "Bundle source: latest bundle in output root\n")
	}
	if noBrowser {
		fmt.Fprintf(output, "Browser launch: skipped by flag\n")
		return
	}
	if openErr != nil {
		fmt.Fprintf(output, "Browser launch: failed (%v)\n", openErr)
		return
	}
	fmt.Fprintf(output, "Browser launch: requested\n")
}

func printExportResult(output io.Writer, result app.ExportResult) {
	fmt.Fprintf(output, "Bundle exported.\n")
	fmt.Fprintf(output, "Bundle: %s\n", result.BundlePath)
	fmt.Fprintf(output, "Archive: %s\n", result.ArchivePath)
}
