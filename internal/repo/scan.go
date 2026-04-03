package repo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Scanner struct{}

func NewScanner() Scanner {
	return Scanner{}
}

func (Scanner) Scan(ctx context.Context, options ScanOptions) (ScanResult, error) {
	patterns, err := LoadPatterns(options.RootPath, options.ExtraIgnorePatterns)
	if err != nil {
		return ScanResult{}, err
	}

	matcher := NewMatcher(patterns)
	scannedAt := time.Now().UTC()

	directories := map[string]struct{}{".": {}}
	var files []FileInfo
	var manifests []string
	var documentation []string

	walkErr := filepath.WalkDir(options.RootPath, func(pathOnDisk string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		relativePath, err := filepath.Rel(options.RootPath, pathOnDisk)
		if err != nil {
			return fmt.Errorf("resolve relative path for %q: %w", pathOnDisk, err)
		}
		relativePath = filepath.ToSlash(relativePath)

		if matcher.ShouldIgnore(relativePath, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			directories[relativePath] = struct{}{}
			return nil
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat %q: %w", pathOnDisk, err)
		}

		analyzedFile, manifestMatch, documentationMatch, err := analyzeFile(pathOnDisk, relativePath, fileInfo.Size(), scannedAt)
		if err != nil {
			return err
		}

		files = append(files, analyzedFile)
		if manifestMatch {
			manifests = append(manifests, relativePath)
		}
		if documentationMatch {
			documentation = append(documentation, relativePath)
		}
		return nil
	})
	if walkErr != nil {
		return ScanResult{}, fmt.Errorf("walk repository: %w", walkErr)
	}

	directoryList := make([]string, 0, len(directories))
	for directory := range directories {
		directoryList = append(directoryList, directory)
	}
	slices.Sort(directoryList)
	slices.Sort(manifests)
	slices.Sort(documentation)
	slices.SortFunc(files, func(left, right FileInfo) int {
		return strings.Compare(left.Path, right.Path)
	})

	projectName := filepath.Base(options.RootPath)
	if projectName == "." || projectName == string(filepath.Separator) {
		projectName = "repository"
	}

	return ScanResult{
		RootPath:       options.RootPath,
		ProjectName:    projectName,
		ScannedAt:      scannedAt,
		Files:          files,
		Directories:    directoryList,
		IgnorePatterns: patterns,
		ManifestFiles:  manifests,
		Documentation:  documentation,
	}, nil
}

func analyzeFile(pathOnDisk, relativePath string, size int64, scannedAt time.Time) (FileInfo, bool, bool, error) {
	file, err := os.Open(pathOnDisk)
	if err != nil {
		return FileInfo{}, false, false, fmt.Errorf("open %q: %w", pathOnDisk, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	textLike, err := isProbablyText(reader)
	if err != nil {
		return FileInfo{}, false, false, fmt.Errorf("inspect %q: %w", pathOnDisk, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return FileInfo{}, false, false, fmt.Errorf("reset %q: %w", pathOnDisk, err)
	}

	extension := strings.ToLower(filepath.Ext(relativePath))
	baseName := strings.ToLower(filepath.Base(relativePath))

	info := FileInfo{
		Path:            relativePath,
		Extension:       extension,
		SizeBytes:       size,
		IsDocumentation: isDocumentationFile(baseName),
		IsEntryPoint:    isEntryPoint(relativePath),
		ScannedAt:       scannedAt,
	}

	manifestMatch := isManifestFile(baseName)
	documentationMatch := info.IsDocumentation

	if !textLike {
		return info, manifestMatch, documentationMatch, nil
	}

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		info.LineCount++
		info.ImportCount += importLikeCount(line)
		info.TodoCount += markerCount(line, "TODO")
		info.FixmeCount += markerCount(line, "FIXME")
		info.HackCount += markerCount(line, "HACK")
	}
	if err := scanner.Err(); err != nil {
		return FileInfo{}, false, false, fmt.Errorf("scan file %q: %w", pathOnDisk, err)
	}

	return info, manifestMatch, documentationMatch, nil
}

func isProbablyText(reader *bufio.Reader) (bool, error) {
	peek, err := reader.Peek(2048)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return false, err
	}
	for _, b := range peek {
		if b == 0 {
			return false, nil
		}
	}
	return true, nil
}

func markerCount(line, marker string) int {
	return strings.Count(strings.ToUpper(line), marker)
}

func importLikeCount(line string) int {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "import "):
		return 1
	case strings.HasPrefix(trimmed, "from "):
		return 1
	case strings.HasPrefix(trimmed, "require("):
		return 1
	case strings.HasPrefix(trimmed, "require \""):
		return 1
	case strings.HasPrefix(trimmed, "#include "):
		return 1
	case strings.HasPrefix(trimmed, "use "):
		return 1
	default:
		return 0
	}
}

func isDocumentationFile(baseName string) bool {
	switch baseName {
	case "readme", "readme.md", "readme.txt", "contributing.md", "architecture.md", "changelog", "changelog.md":
		return true
	default:
		return strings.HasSuffix(baseName, ".md")
	}
}

func isManifestFile(baseName string) bool {
	switch baseName {
	case "go.mod", "package.json", "pyproject.toml", "cargo.toml", "composer.json", "pom.xml", "build.gradle", "build.gradle.kts":
		return true
	default:
		return false
	}
}

func isEntryPoint(relativePath string) bool {
	normalized := strings.ToLower(filepath.ToSlash(relativePath))
	baseName := path.Base(normalized)

	switch {
	case baseName == "main.go":
		return true
	case strings.HasPrefix(normalized, "cmd/") && strings.HasSuffix(baseName, ".go"):
		return true
	case strings.Contains(normalized, "/cmd/") && strings.HasSuffix(baseName, ".go"):
		return true
	case baseName == "main.py" || baseName == "main.ts" || baseName == "main.js":
		return true
	case strings.Contains(baseName, "server") && strings.HasSuffix(baseName, ".go"):
		return true
	default:
		return false
	}
}
