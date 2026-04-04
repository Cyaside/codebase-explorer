package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func nextBundleName(outputRoot string, projectName string, now time.Time) (string, error) {
	baseName := fmt.Sprintf("%s_%s", now.UTC().Format("2006-01-02_150405"), sanitizeName(projectName))
	candidate := filepath.Join(outputRoot, baseName)
	if _, err := os.Stat(candidate); err != nil {
		if os.IsNotExist(err) {
			return baseName, nil
		}
		return "", fmt.Errorf("inspect bundle path %q: %w", candidate, err)
	}

	for suffix := 1; suffix <= 999; suffix++ {
		name := fmt.Sprintf("%s_%03d", baseName, suffix)
		pathOnDisk := filepath.Join(outputRoot, name)
		if _, err := os.Stat(pathOnDisk); os.IsNotExist(err) {
			return name, nil
		} else if err != nil {
			return "", fmt.Errorf("inspect bundle path %q: %w", pathOnDisk, err)
		}
	}

	return "", fmt.Errorf("could not allocate unique bundle name for %q", baseName)
}
