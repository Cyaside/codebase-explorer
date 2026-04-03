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
}

func Load() (Settings, error) {
	outputRoot := strings.TrimSpace(os.Getenv("CODEARCH_OUTPUT_ROOT"))
	source := "defaults"
	if outputRoot == "" {
		outputRoot = defaultOutputRoot
	} else {
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
	}, nil
}
