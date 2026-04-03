package repo

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var defaultSkippedDirectories = map[string]struct{}{
	".git":         {},
	".hg":          {},
	".svn":         {},
	".idea":        {},
	".vscode":      {},
	".cache":       {},
	".tmp":         {},
	"bin":          {},
	"build":        {},
	"coverage":     {},
	"dist":         {},
	"node_modules": {},
	"out":          {},
	"target":       {},
	"tmp":          {},
	"vendor":       {},
}

type Matcher struct {
	patterns []string
}

func LoadPatterns(root string, extra []string) ([]string, error) {
	patterns := make([]string, 0, len(extra)+8)
	files := []string{
		filepath.Join(root, ".gitignore"),
		filepath.Join(root, ".codearchaeologistignore"),
	}

	for _, filePath := range files {
		loaded, err := readIgnoreFile(filePath)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, loaded...)
	}

	for _, pattern := range extra {
		trimmed := strings.TrimSpace(pattern)
		if trimmed != "" {
			patterns = append(patterns, trimmed)
		}
	}

	slices.Sort(patterns)
	return slices.Compact(patterns), nil
}

func NewMatcher(patterns []string) Matcher {
	return Matcher{patterns: patterns}
}

func (m Matcher) ShouldIgnore(relativePath string, isDir bool) bool {
	if relativePath == "." || relativePath == "" {
		return false
	}

	normalized := filepath.ToSlash(relativePath)
	base := path.Base(normalized)

	if isDir {
		if _, skipped := defaultSkippedDirectories[base]; skipped {
			return true
		}
	}

	for _, pattern := range m.patterns {
		if matchesPattern(pattern, normalized, base, isDir) {
			return true
		}
	}

	return false
}

func readIgnoreFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ignore file %q: %w", filePath, err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan ignore file %q: %w", filePath, err)
	}

	return patterns, nil
}

func matchesPattern(pattern, normalizedPath, base string, isDir bool) bool {
	dirOnly := strings.HasSuffix(pattern, "/")
	if dirOnly && !isDir {
		return false
	}

	candidate := strings.TrimSuffix(pattern, "/")
	anchored := strings.HasPrefix(candidate, "/")
	candidate = strings.TrimPrefix(candidate, "/")
	if candidate == "" {
		return false
	}

	if anchored {
		return matchGlob(candidate, normalizedPath)
	}

	if !strings.Contains(candidate, "/") {
		return matchGlob(candidate, base)
	}

	if matchGlob(candidate, normalizedPath) {
		return true
	}

	segments := strings.Split(normalizedPath, "/")
	for index := range segments {
		tail := strings.Join(segments[index:], "/")
		if matchGlob(candidate, tail) {
			return true
		}
	}

	return false
}

func matchGlob(pattern, value string) bool {
	if matched, err := path.Match(pattern, value); err == nil && matched {
		return true
	}
	return pattern == value
}
