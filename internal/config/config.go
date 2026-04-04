package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultOutputRoot       = "out"
	defaultAppVersion       = "dev"
	defaultOutputKeepLatest = 10
)

type Settings struct {
	DefaultOutputRoot string
	CacheRoot         string
	CacheEnabled      bool
	AppVersion        string
	ConfigSource      string
	OutputKeepLatest  int
	Provider          ProviderSettings
}

func Load() (Settings, error) {
	outputRoot := strings.TrimSpace(getEnv("CODEARCH_OUTPUT_ROOT"))
	outputKeepLatest, outputKeepLatestFromEnv, err := loadOutputKeepLatest()
	if err != nil {
		return Settings{}, err
	}
	cacheRoot, cacheRootFromEnv := loadCacheRoot()
	cacheEnabled, cacheEnabledFromEnv, err := loadCacheEnabled()
	if err != nil {
		return Settings{}, err
	}
	source := "defaults"
	if outputRoot == "" {
		outputRoot = defaultOutputRoot
	} else {
		source = "environment"
	}
	if outputKeepLatestFromEnv {
		source = "environment"
	}
	if cacheRootFromEnv || cacheEnabledFromEnv {
		source = "environment"
	}

	providerSettings, providerFromEnv := loadProviderSettings()
	if providerFromEnv {
		source = "environment"
	}

	cleanRoot := filepath.Clean(outputRoot)
	if cleanRoot == "." {
		return Settings{}, fmt.Errorf("resolved output root %q is not valid", outputRoot)
	}

	return Settings{
		DefaultOutputRoot: cleanRoot,
		CacheRoot:         cacheRoot,
		CacheEnabled:      cacheEnabled,
		AppVersion:        defaultAppVersion,
		ConfigSource:      source,
		OutputKeepLatest:  outputKeepLatest,
		Provider:          providerSettings,
	}, nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}
