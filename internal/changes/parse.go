package changes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type normalizedItem struct {
	SourcePath string
	SourceKind string
	Title      string
	Body       string
	Labels     []string
	RawText    string
	Terms      map[string]struct{}
}

var (
	headingPattern = regexp.MustCompile(`^#{1,6}\s+(.+)$`)
	bulletPattern  = regexp.MustCompile(`^[-*+]\s+(.+)$`)
	tokenPattern   = regexp.MustCompile(`[A-Za-z][A-Za-z0-9._/-]{2,}`)
)

func loadSource(path string) (Source, []normalizedItem) {
	source := Source{
		Path:   path,
		Kind:   detectKind(path),
		Format: detectFormat(path),
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		source.Status = SourceStatusFailed
		source.Message = err.Error()
		return source, nil
	}

	var items []normalizedItem
	switch source.Format {
	case "json":
		items, err = parseJSONItems(path, source.Kind, contents)
	default:
		items, err = parseTextItems(path, source.Kind, contents)
	}
	if err != nil {
		source.Status = SourceStatusFailed
		source.Message = err.Error()
		return source, nil
	}

	source.ItemCount = len(items)
	if len(items) == 0 {
		source.Status = SourceStatusEmpty
		source.Message = "no structured change entries found"
		return source, nil
	}

	source.Status = SourceStatusParsed
	return source, items
}

func parseJSONItems(path string, kind string, contents []byte) ([]normalizedItem, error) {
	var payload any
	if err := json.Unmarshal(contents, &payload); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	objects := extractIssueObjects(payload)
	if len(objects) == 0 {
		return nil, fmt.Errorf("unsupported issue export shape")
	}

	items := make([]normalizedItem, 0, len(objects))
	for _, object := range objects {
		title := firstString(object, "title", "name", "summary")
		body := firstString(object, "body", "description", "content")
		labels := extractLabels(object["labels"])
		if len(labels) == 0 {
			labels = extractLabels(object["tags"])
		}
		item := newNormalizedItem(path, kind, title, body, labels)
		if item.RawText == "" {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func parseTextItems(path string, kind string, contents []byte) ([]normalizedItem, error) {
	lines := strings.Split(string(contents), "\n")
	items := []normalizedItem{}
	currentTitle := ""
	currentBody := []string{}

	flush := func() {
		item := newNormalizedItem(path, kind, currentTitle, strings.Join(currentBody, "\n"), nil)
		if item.RawText != "" {
			items = append(items, item)
		}
		currentTitle = ""
		currentBody = nil
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			if currentTitle != "" && len(currentBody) > 0 {
				flush()
			}
			continue
		}

		if match := headingPattern.FindStringSubmatch(line); len(match) == 2 {
			if currentTitle != "" || len(currentBody) > 0 {
				flush()
			}
			currentTitle = strings.TrimSpace(match[1])
			continue
		}

		if match := bulletPattern.FindStringSubmatch(line); len(match) == 2 {
			bullet := strings.TrimSpace(match[1])
			if currentTitle == "" && len(currentBody) == 0 {
				currentTitle = bullet
				flush()
				continue
			}
			currentBody = append(currentBody, bullet)
			continue
		}

		if currentTitle == "" {
			currentTitle = line
			continue
		}
		currentBody = append(currentBody, line)
	}

	if currentTitle != "" || len(currentBody) > 0 {
		flush()
	}

	return items, nil
}

func extractIssueObjects(payload any) []map[string]any {
	switch typed := payload.(type) {
	case []any:
		return sliceObjects(typed)
	case map[string]any:
		for _, key := range []string{"issues", "items", "data", "results", "values"} {
			if nested, ok := typed[key]; ok {
				if objects := extractIssueObjects(nested); len(objects) > 0 {
					return objects
				}
			}
		}
		if looksLikeIssueObject(typed) {
			return []map[string]any{typed}
		}
	}

	return nil
}

func sliceObjects(values []any) []map[string]any {
	objects := make([]map[string]any, 0, len(values))
	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok {
			continue
		}
		objects = append(objects, object)
	}
	return objects
}

func looksLikeIssueObject(object map[string]any) bool {
	return firstString(object, "title", "name", "summary", "body", "description", "content") != ""
}

func firstString(object map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := object[key]
		if !ok {
			continue
		}
		if stringValue, ok := value.(string); ok {
			return strings.TrimSpace(stringValue)
		}
	}
	return ""
}

func extractLabels(value any) []string {
	switch typed := value.(type) {
	case []any:
		labels := make([]string, 0, len(typed))
		for _, item := range typed {
			switch label := item.(type) {
			case string:
				labels = append(labels, strings.TrimSpace(label))
			case map[string]any:
				if name := firstString(label, "name", "label", "value"); name != "" {
					labels = append(labels, name)
				}
			}
		}
		return compactStrings(labels)
	case []string:
		return compactStrings(typed)
	default:
		return nil
	}
}

func detectFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	case ".txt":
		return "text"
	default:
		return "text"
	}
}

func detectKind(path string) string {
	lowerName := strings.ToLower(filepath.Base(path))
	switch {
	case strings.Contains(lowerName, "changelog") || strings.Contains(lowerName, "release"):
		return "changelog"
	case strings.HasSuffix(lowerName, ".json") || strings.Contains(lowerName, "issue") || strings.Contains(lowerName, "ticket"):
		return "issue-export"
	default:
		return "notes"
	}
}

func newNormalizedItem(sourcePath string, sourceKind string, title string, body string, labels []string) normalizedItem {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	labels = compactStrings(labels)

	parts := []string{title, body}
	if len(labels) > 0 {
		parts = append(parts, strings.Join(labels, " "))
	}
	rawText := strings.ToLower(strings.Join(compactStrings(parts), "\n"))

	return normalizedItem{
		SourcePath: sourcePath,
		SourceKind: sourceKind,
		Title:      title,
		Body:       body,
		Labels:     labels,
		RawText:    rawText,
		Terms:      extractTerms(parts...),
	}
}

func extractTerms(parts ...string) map[string]struct{} {
	terms := make(map[string]struct{})
	for _, part := range parts {
		for _, token := range tokenPattern.FindAllString(strings.ToLower(part), -1) {
			for _, piece := range splitToken(token) {
				if ignoreToken(piece) {
					continue
				}
				terms[piece] = struct{}{}
			}
		}
	}
	return terms
}

func splitToken(token string) []string {
	pieces := strings.FieldsFunc(token, func(r rune) bool {
		return r == '/' || r == '\\' || r == '-' || r == '_' || r == '.'
	})
	return compactStrings(pieces)
}

func compactStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

var stopWords = map[string]struct{}{
	"about": {}, "after": {}, "again": {}, "analysis": {}, "around": {}, "before": {},
	"being": {}, "bundle": {}, "change": {}, "changes": {}, "code": {}, "could": {},
	"deterministic": {}, "does": {}, "during": {}, "error": {}, "errors": {}, "feature": {},
	"file": {}, "files": {}, "from": {}, "have": {}, "into": {}, "issue": {}, "issues": {},
	"latest": {}, "local": {}, "mode": {}, "module": {}, "modules": {}, "note": {},
	"problem": {}, "problems": {}, "project": {}, "release": {}, "repo": {}, "report": {},
	"runtime": {}, "should": {}, "support": {}, "system": {}, "task": {}, "that": {},
	"them": {}, "there": {}, "these": {}, "this": {}, "ticket": {}, "tickets": {},
	"update": {}, "updated": {}, "using": {}, "with": {}, "without": {},
}

var shortAllowedTokens = map[string]struct{}{
	"api": {}, "app": {}, "cli": {}, "sql": {}, "ui": {}, "web": {},
}

func ignoreToken(token string) bool {
	if token == "" {
		return true
	}
	if _, found := stopWords[token]; found {
		return true
	}
	if len(token) >= 4 {
		return false
	}
	_, allowed := shortAllowedTokens[token]
	return !allowed
}
