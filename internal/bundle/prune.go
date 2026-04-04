package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type bundleCandidate struct {
	path    string
	name    string
	modTime int64
}

func pruneBundles(outputRoot, currentBundlePath string, keepLatest int) (int, error) {
	if keepLatest <= 0 {
		return 0, nil
	}

	entries, err := os.ReadDir(outputRoot)
	if err != nil {
		return 0, fmt.Errorf("list output root %q: %w", outputRoot, err)
	}

	candidates := make([]bundleCandidate, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		bundlePath := filepath.Join(outputRoot, entry.Name())
		if !isManagedBundle(bundlePath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return 0, fmt.Errorf("inspect bundle %q: %w", bundlePath, err)
		}

		candidates = append(candidates, bundleCandidate{
			path:    bundlePath,
			name:    entry.Name(),
			modTime: info.ModTime().UTC().UnixNano(),
		})
	}

	if len(candidates) <= keepLatest {
		return 0, nil
	}

	slices.SortFunc(candidates, func(left, right bundleCandidate) int {
		if left.path == currentBundlePath && right.path != currentBundlePath {
			return -1
		}
		if right.path == currentBundlePath && left.path != currentBundlePath {
			return 1
		}
		if left.modTime != right.modTime {
			if left.modTime > right.modTime {
				return -1
			}
			return 1
		}
		return strings.Compare(right.name, left.name)
	})

	prunedBundles := 0
	for index := keepLatest; index < len(candidates); index++ {
		if candidates[index].path == currentBundlePath {
			continue
		}
		if err := os.RemoveAll(candidates[index].path); err != nil {
			return prunedBundles, fmt.Errorf("remove old bundle %q: %w", candidates[index].path, err)
		}
		prunedBundles++
	}

	return prunedBundles, nil
}

func isManagedBundle(bundlePath string) bool {
	info, err := os.Stat(filepath.Join(bundlePath, "data", "contract.json"))
	if err != nil {
		return false
	}
	return !info.IsDir()
}
