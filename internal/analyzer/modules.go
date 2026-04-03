package analyzer

import (
	"slices"
	"sort"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func buildModules(files []repo.FileInfo) []ModuleInfo {
	type aggregate struct {
		fileCount       int
		totalLines      int
		entryPointCount int
		markerCount     int
		languages       map[string]struct{}
	}

	modules := map[string]aggregate{}
	for _, file := range files {
		modulePath := topLevelPath(file.Path)
		current := modules[modulePath]
		if current.languages == nil {
			current.languages = map[string]struct{}{}
		}
		current.fileCount++
		current.totalLines += file.LineCount
		current.markerCount += markerTotal(file)
		if file.IsEntryPoint {
			current.entryPointCount++
		}
		current.languages[detectLanguage(file.Extension)] = struct{}{}
		modules[modulePath] = current
	}

	results := make([]ModuleInfo, 0, len(modules))
	for modulePath, aggregate := range modules {
		languages := make([]string, 0, len(aggregate.languages))
		for language := range aggregate.languages {
			languages = append(languages, language)
		}
		slices.Sort(languages)

		results = append(results, ModuleInfo{
			Path:            modulePath,
			FileCount:       aggregate.fileCount,
			TotalLines:      aggregate.totalLines,
			Languages:       languages,
			EntryPointCount: aggregate.entryPointCount,
			MarkerCount:     aggregate.markerCount,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalLines == results[j].TotalLines {
			return results[i].Path < results[j].Path
		}
		return results[i].TotalLines > results[j].TotalLines
	})

	return results
}

func buildImportantDirectories(modules []ModuleInfo) []string {
	var results []string
	for _, module := range modules {
		if module.Path == "." {
			continue
		}
		results = append(results, module.Path)
		if len(results) == 5 {
			break
		}
	}
	return results
}

func buildCoreModules(modules []ModuleInfo) []string {
	var results []string
	for _, module := range modules {
		if module.Path == "." {
			continue
		}
		results = append(results, module.Path)
		if len(results) == 4 {
			break
		}
	}
	return results
}
