package analyzer

import "strings"

func topLevelPath(filePath string) string {
	normalized := filepathToSlash(filePath)
	if normalized == "." || normalized == "" {
		return "."
	}
	if !strings.Contains(normalized, "/") {
		return "."
	}

	segments := strings.Split(normalized, "/")
	if len(segments) == 0 {
		return "."
	}

	return segments[0]
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}
