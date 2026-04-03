package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultOutputRoot = "out"
	defaultAppVersion = "phase1"
)

type Settings struct {
	DefaultOutputRoot string
	AppVersion        string
	ConfigSource      string
	Provider          ProviderSettings
}

func Load() (Settings, error) {
	outputRoot := strings.TrimSpace(getEnv("CODEARCH_OUTPUT_ROOT"))
	source := "defaults"
	if outputRoot == "" {
		outputRoot = defaultOutputRoot
	} else {
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
		AppVersion:        defaultAppVersion,
		ConfigSource:      source,
		Provider:          providerSettings,
	}, nil
}

func getEnv(key string) string {
	return os.Getenv(key)
}
