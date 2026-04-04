package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultCacheFolderName = "codearch"

func loadCacheRoot() (string, bool) {
	rawValue := strings.TrimSpace(getEnv("CODEARCH_CACHE_ROOT"))
	if rawValue != "" {
		return filepath.Clean(rawValue), true
	}

	if userCacheDir, err := os.UserCacheDir(); err == nil && strings.TrimSpace(userCacheDir) != "" {
		return filepath.Join(userCacheDir, defaultCacheFolderName), false
	}

	return filepath.Clean(".codearch-cache"), false
}

func loadCacheEnabled() (bool, bool, error) {
	rawValue := strings.TrimSpace(strings.ToLower(getEnv("CODEARCH_CACHE")))
	if rawValue == "" {
		return true, false, nil
	}

	switch rawValue {
	case "1", "true", "yes", "on":
		return true, true, nil
	case "0", "false", "no", "off":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("invalid CODEARCH_CACHE %q: expected true/false", rawValue)
	}
}
