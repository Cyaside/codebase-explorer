package config

import (
	"fmt"
	"strconv"
	"strings"
)

func loadOutputKeepLatest() (int, bool, error) {
	rawValue := strings.TrimSpace(getEnv("CODEARCH_OUTPUT_KEEP"))
	if rawValue == "" {
		return defaultOutputKeepLatest, false, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, true, fmt.Errorf("invalid CODEARCH_OUTPUT_KEEP %q: %w", rawValue, err)
	}
	if value < 0 {
		return 0, true, fmt.Errorf("invalid CODEARCH_OUTPUT_KEEP %q: must be zero or greater", rawValue)
	}

	return value, true, nil
}
