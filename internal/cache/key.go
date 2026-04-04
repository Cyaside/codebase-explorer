package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type fileSnapshot struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	IsDir    bool   `json:"is_dir"`
}

type deterministicKeyPayload struct {
	Version           string         `json:"version"`
	RootPath          string         `json:"root_path"`
	DeterministicOnly bool           `json:"deterministic_only"`
	IgnorePatterns    []string       `json:"ignore_patterns"`
	Repository        []fileSnapshot `json:"repository"`
	SupportFiles      []fileSnapshot `json:"support_files"`
}

type providerKeyPayload struct {
	Version string                    `json:"version"`
	Name    string                    `json:"name"`
	Model   string                    `json:"model"`
	BaseURL string                    `json:"base_url"`
	APIKey  string                    `json:"api_key_hash"`
	Context provider.CondensedContext `json:"context"`
}

func BuildDeterministicKey(rootPath string, extraIgnorePatterns []string, supportFiles []string, version string, deterministicOnly bool) (string, error) {
	patterns, err := repo.LoadPatterns(rootPath, extraIgnorePatterns)
	if err != nil {
		return "", err
	}

	repositorySnapshot, err := snapshotRepository(rootPath, patterns)
	if err != nil {
		return "", err
	}
	supportSnapshot, err := snapshotSupportFiles(supportFiles)
	if err != nil {
		return "", err
	}

	payload := deterministicKeyPayload{
		Version:           strings.TrimSpace(version),
		RootPath:          filepath.Clean(rootPath),
		DeterministicOnly: deterministicOnly,
		IgnorePatterns:    patterns,
		Repository:        repositorySnapshot,
		SupportFiles:      supportSnapshot,
	}

	return hashPayload(payload)
}

func BuildProviderKey(config provider.Config, context provider.CondensedContext, version string) (string, error) {
	payload := providerKeyPayload{
		Version: strings.TrimSpace(version),
		Name:    strings.TrimSpace(config.Name),
		Model:   strings.TrimSpace(config.Model),
		BaseURL: strings.TrimSpace(config.BaseURL),
		APIKey:  hashText(strings.TrimSpace(config.APIKey)),
		Context: context,
	}
	return hashPayload(payload)
}

func snapshotRepository(rootPath string, patterns []string) ([]fileSnapshot, error) {
	matcher := repo.NewMatcher(patterns)
	snapshots := []fileSnapshot{}

	walkErr := filepath.WalkDir(rootPath, func(pathOnDisk string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(rootPath, pathOnDisk)
		if err != nil {
			return fmt.Errorf("resolve relative path for %q: %w", pathOnDisk, err)
		}
		relativePath = filepath.ToSlash(relativePath)

		if matcher.ShouldIgnore(relativePath, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relativePath == "." || relativePath == "" {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat %q: %w", pathOnDisk, err)
		}

		snapshots = append(snapshots, fileSnapshot{
			Path:     relativePath,
			Size:     info.Size(),
			Modified: info.ModTime().UTC().UnixNano(),
			IsDir:    entry.IsDir(),
		})
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("snapshot repository: %w", walkErr)
	}

	sort.SliceStable(snapshots, func(i int, j int) bool {
		return snapshots[i].Path < snapshots[j].Path
	})
	return snapshots, nil
}

func snapshotSupportFiles(paths []string) ([]fileSnapshot, error) {
	sortedPaths := append([]string(nil), paths...)
	sort.Strings(sortedPaths)

	snapshots := make([]fileSnapshot, 0, len(sortedPaths))
	for _, pathOnDisk := range sortedPaths {
		info, err := os.Stat(pathOnDisk)
		if err != nil {
			if os.IsNotExist(err) {
				snapshots = append(snapshots, fileSnapshot{Path: pathOnDisk, Modified: -1})
				continue
			}
			return nil, fmt.Errorf("stat support file %q: %w", pathOnDisk, err)
		}

		snapshots = append(snapshots, fileSnapshot{
			Path:     pathOnDisk,
			Size:     info.Size(),
			Modified: info.ModTime().UTC().UnixNano(),
			IsDir:    info.IsDir(),
		})
	}

	return snapshots, nil
}

func hashPayload(value any) (string, error) {
	contents, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal cache key payload: %w", err)
	}

	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:]), nil
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
