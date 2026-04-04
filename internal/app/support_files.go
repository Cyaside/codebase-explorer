package app

import (
	"os"
	"path/filepath"
	"strings"
)

func resolveSupportFiles(repoPath string, values []string) []string {
	seen := map[string]struct{}{}
	files := make([]string, 0, len(values))

	for _, value := range values {
		resolved := resolveSupportFile(repoPath, value)
		if resolved == "" {
			continue
		}
		if _, found := seen[resolved]; found {
			continue
		}
		seen[resolved] = struct{}{}
		files = append(files, resolved)
	}

	return files
}

func resolveSupportFile(repoPath string, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}

	workingPath, err := filepath.Abs(trimmed)
	if err == nil && pathExists(workingPath) {
		return workingPath
	}

	repoRelativePath, err := filepath.Abs(filepath.Join(repoPath, trimmed))
	if err == nil && pathExists(repoRelativePath) {
		return repoRelativePath
	}

	if err == nil {
		return workingPath
	}
	return filepath.Clean(filepath.Join(repoPath, trimmed))
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
