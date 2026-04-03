package analyzer

import (
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

var supportDirectoryNames = map[string]struct{}{
	"__fixtures__":  {},
	"__snapshots__": {},
	"fixtures":      {},
	"snapshots":     {},
	"testdata":      {},
}

func buildOrientationScan(scanResult repo.ScanResult) repo.ScanResult {
	orientation := scanResult
	orientation.Files = filterOrientationFiles(scanResult.Files)
	if len(orientation.Files) == 0 {
		return scanResult
	}

	orientation.ManifestFiles = filterOrientationPaths(scanResult.ManifestFiles)
	orientation.Documentation = filterOrientationPaths(scanResult.Documentation)

	return orientation
}

func filterOrientationFiles(files []repo.FileInfo) []repo.FileInfo {
	filtered := make([]repo.FileInfo, 0, len(files))
	for _, file := range files {
		if isSupportPath(file.Path) {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func filterOrientationPaths(paths []string) []string {
	filtered := make([]string, 0, len(paths))
	for _, pathValue := range paths {
		if isSupportPath(pathValue) {
			continue
		}
		filtered = append(filtered, pathValue)
	}
	return filtered
}

func isSupportPath(pathValue string) bool {
	normalized := filepathToSlash(pathValue)
	for _, segment := range strings.Split(normalized, "/") {
		if _, found := supportDirectoryNames[strings.ToLower(segment)]; found {
			return true
		}
	}
	return false
}
