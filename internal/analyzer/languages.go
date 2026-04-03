package analyzer

import "strings"

var languageByExtension = map[string]string{
	".c":     "C",
	".cc":    "C++",
	".cpp":   "C++",
	".cs":    "C#",
	".css":   "CSS",
	".go":    "Go",
	".h":     "C/C++ Header",
	".hpp":   "C++ Header",
	".html":  "HTML",
	".java":  "Java",
	".js":    "JavaScript",
	".json":  "JSON",
	".jsx":   "JavaScript",
	".kt":    "Kotlin",
	".md":    "Markdown",
	".php":   "PHP",
	".py":    "Python",
	".rb":    "Ruby",
	".rs":    "Rust",
	".scss":  "SCSS",
	".sh":    "Shell",
	".sql":   "SQL",
	".swift": "Swift",
	".toml":  "TOML",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".txt":   "Text",
	".xml":   "XML",
	".yaml":  "YAML",
	".yml":   "YAML",
}

func detectLanguage(extension string) string {
	if language, found := languageByExtension[strings.ToLower(extension)]; found {
		return language
	}
	return "Other"
}
