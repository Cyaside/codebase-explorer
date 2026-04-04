package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type bundleLocation struct {
	path    string
	name    string
	modTime int64
}

func (s Service) Open(_ context.Context, request OpenRequest) (OpenResult, error) {
	outputRoot, err := s.resolveOutputRoot("")
	if err != nil {
		return OpenResult{}, err
	}

	bundlePath, resolvedLatest, err := resolveBundlePath(request.BundlePath, outputRoot)
	if err != nil {
		return OpenResult{}, err
	}

	viewerPath := filepath.Join(bundlePath, "ui", "index.html")
	info, err := os.Stat(viewerPath)
	if err != nil {
		return OpenResult{}, fmt.Errorf("viewer path %q: %w", viewerPath, err)
	}
	if info.IsDir() {
		return OpenResult{}, fmt.Errorf("viewer path %q is not a file", viewerPath)
	}

	return OpenResult{
		BundlePath:     bundlePath,
		ViewerPath:     viewerPath,
		ResolvedLatest: resolvedLatest,
	}, nil
}

func resolveBundlePath(bundlePath, outputRoot string) (string, bool, error) {
	if strings.TrimSpace(bundlePath) != "" {
		absolutePath, err := filepath.Abs(bundlePath)
		if err != nil {
			return "", false, fmt.Errorf("resolve bundle path %q: %w", bundlePath, err)
		}
		if err := validateBundlePath(absolutePath); err != nil {
			return "", false, err
		}
		return absolutePath, false, nil
	}

	entries, err := os.ReadDir(outputRoot)
	if err != nil {
		return "", false, fmt.Errorf("list output root %q: %w", outputRoot, err)
	}

	candidates := make([]bundleLocation, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := filepath.Join(outputRoot, entry.Name())
		if !looksLikeBundle(candidatePath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return "", false, fmt.Errorf("inspect bundle %q: %w", candidatePath, err)
		}

		candidates = append(candidates, bundleLocation{
			path:    candidatePath,
			name:    entry.Name(),
			modTime: info.ModTime().UTC().UnixNano(),
		})
	}

	if len(candidates) == 0 {
		return "", false, fmt.Errorf("no output bundle found in %q", outputRoot)
	}

	slices.SortFunc(candidates, func(left, right bundleLocation) int {
		if left.modTime != right.modTime {
			if left.modTime > right.modTime {
				return -1
			}
			return 1
		}
		return strings.Compare(right.name, left.name)
	})

	return candidates[0].path, true, nil
}

func validateBundlePath(bundlePath string) error {
	info, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("bundle path %q: %w", bundlePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("bundle path %q is not a directory", bundlePath)
	}
	if !looksLikeBundle(bundlePath) {
		return fmt.Errorf("bundle path %q does not look like a Codebase Explorer bundle", bundlePath)
	}
	return nil
}

func looksLikeBundle(bundlePath string) bool {
	contractPath := filepath.Join(bundlePath, "data", "contract.json")
	viewerPath := filepath.Join(bundlePath, "ui", "index.html")
	if info, err := os.Stat(contractPath); err != nil || info.IsDir() {
		return false
	}
	if info, err := os.Stat(viewerPath); err != nil || info.IsDir() {
		return false
	}
	return true
}
