package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/app"
)

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  codearch analyze <repo-path> [--output <dir>] [--deterministic-only] [--ignore <pattern>]")
	fmt.Fprintln(output, "  codearch doctor")
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
	fmt.Fprintf(output, "Bundle: %s\n", result.OutputPath)
}

func printDoctorResult(output io.Writer, result app.DoctorResult) {
	fmt.Fprintf(output, "Config source: %s\n", result.ConfigSource)
	fmt.Fprintf(output, "Output root: %s\n", result.OutputRoot)
	for _, check := range result.Checks {
		fmt.Fprintf(output, "- [%s] %s: %s\n", strings.ToUpper(check.Status), check.Name, check.Detail)
	}
}
