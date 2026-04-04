package app

import (
	"fmt"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

const (
	largeRepoFileThreshold    = 5000
	largeRepoLineThreshold    = 500000
	manySupportFilesThreshold = 10
)

func buildWarnings(analysis analyzer.Result, supportFileCount int) []string {
	warnings := []string{}

	if analysis.Metrics.TotalFiles >= largeRepoFileThreshold {
		warnings = append(warnings, fmt.Sprintf("Large repository detected: %d files may increase scan and bundle generation time.", analysis.Metrics.TotalFiles))
	}
	if analysis.Metrics.TotalLines >= largeRepoLineThreshold {
		warnings = append(warnings, fmt.Sprintf("Large text footprint detected: %d lines may make reports and viewer payloads heavier.", analysis.Metrics.TotalLines))
	}
	if supportFileCount >= manySupportFilesThreshold {
		warnings = append(warnings, fmt.Sprintf("%d support files were provided; change correlation may include more noise than usual.", supportFileCount))
	}

	return warnings
}
