package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrCacheMiss = errors.New("cache miss")

type Store struct {
	root    string
	enabled bool
}

func NewStore(root string, enabled bool) Store {
	return Store{
		root:    root,
		enabled: enabled,
	}
}

func (s Store) Enabled() bool {
	return s.enabled
}

func (s Store) Root() string {
	return s.root
}

func (s Store) Ensure() error {
	if !s.enabled {
		return nil
	}
	return os.MkdirAll(s.root, 0o755)
}

func (s Store) LoadDeterministic(key string) (DeterministicPayload, error) {
	return loadJSON[DeterministicPayload](s.kindPath("deterministic", key))
}

func (s Store) SaveDeterministic(payload DeterministicPayload) error {
	return writeJSON(s.kindPath("deterministic", payload.Key), payload)
}

func (s Store) LoadProvider(key string) (ProviderPayload, error) {
	return loadJSON[ProviderPayload](s.kindPath("provider", key))
}

func (s Store) SaveProvider(payload ProviderPayload) error {
	return writeJSON(s.kindPath("provider", payload.Key), payload)
}

func (s Store) Clear() (ClearResult, error) {
	result := ClearResult{CacheRoot: s.root}

	if s.root == "" {
		return result, fmt.Errorf("cache root is not configured")
	}

	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(s.root, 0o755); err != nil {
				return result, fmt.Errorf("prepare cache root %q: %w", s.root, err)
			}
			return result, nil
		}
		return result, fmt.Errorf("read cache root %q: %w", s.root, err)
	}

	for _, entry := range entries {
		result.RemovedEntries++
		if err := os.RemoveAll(filepath.Join(s.root, entry.Name())); err != nil {
			return result, fmt.Errorf("remove cache entry %q: %w", entry.Name(), err)
		}
	}

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return result, fmt.Errorf("restore cache root %q: %w", s.root, err)
	}

	return result, nil
}

func (s Store) kindPath(kind string, key string) string {
	return filepath.Join(s.root, kind, key+".json")
}

func loadJSON[T any](pathOnDisk string) (T, error) {
	var value T

	contents, err := os.ReadFile(pathOnDisk)
	if err != nil {
		if os.IsNotExist(err) {
			return value, ErrCacheMiss
		}
		return value, fmt.Errorf("read cache file %q: %w", pathOnDisk, err)
	}

	if err := json.Unmarshal(contents, &value); err != nil {
		return value, fmt.Errorf("decode cache file %q: %w", pathOnDisk, err)
	}

	return value, nil
}

func writeJSON(pathOnDisk string, value any) error {
	if err := os.MkdirAll(filepath.Dir(pathOnDisk), 0o755); err != nil {
		return fmt.Errorf("prepare cache directory for %q: %w", pathOnDisk, err)
	}

	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache payload %q: %w", pathOnDisk, err)
	}
	contents = append(contents, '\n')

	if err := os.WriteFile(pathOnDisk, contents, 0o644); err != nil {
		return fmt.Errorf("write cache file %q: %w", pathOnDisk, err)
	}

	return nil
}
