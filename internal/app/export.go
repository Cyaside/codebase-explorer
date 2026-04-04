package app

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (s Service) Export(_ context.Context, request ExportRequest) (ExportResult, error) {
	outputRoot, err := s.resolveOutputRoot("")
	if err != nil {
		return ExportResult{}, err
	}

	bundlePath, _, err := resolveBundlePath(request.BundlePath, outputRoot)
	if err != nil {
		return ExportResult{}, err
	}

	archivePath, err := resolveArchivePath(bundlePath, request.OutputPath)
	if err != nil {
		return ExportResult{}, err
	}

	if err := zipBundle(bundlePath, archivePath); err != nil {
		return ExportResult{}, err
	}

	return ExportResult{
		BundlePath:  bundlePath,
		ArchivePath: archivePath,
	}, nil
}

func resolveArchivePath(bundlePath string, explicitOutput string) (string, error) {
	if strings.TrimSpace(explicitOutput) != "" {
		absolutePath, err := filepath.Abs(explicitOutput)
		if err != nil {
			return "", fmt.Errorf("resolve export output %q: %w", explicitOutput, err)
		}
		if !strings.EqualFold(filepath.Ext(absolutePath), ".zip") {
			absolutePath += ".zip"
		}
		return absolutePath, nil
	}

	return filepath.Abs(bundlePath + ".zip")
}

func zipBundle(bundlePath string, archivePath string) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return fmt.Errorf("prepare export directory %q: %w", filepath.Dir(archivePath), err)
	}

	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create export archive %q: %w", archivePath, err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	walkErr := filepath.Walk(bundlePath, func(pathOnDisk string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(bundlePath, pathOnDisk)
		if err != nil {
			return fmt.Errorf("resolve export path %q: %w", pathOnDisk, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create zip header for %q: %w", pathOnDisk, err)
		}
		header.Name = filepath.ToSlash(relativePath)
		header.Method = zip.Deflate

		zipWriter, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create zip entry %q: %w", relativePath, err)
		}

		sourceFile, err := os.Open(pathOnDisk)
		if err != nil {
			return fmt.Errorf("open bundle file %q: %w", pathOnDisk, err)
		}
		if _, err := io.Copy(zipWriter, sourceFile); err != nil {
			_ = sourceFile.Close()
			return fmt.Errorf("write zip entry %q: %w", relativePath, err)
		}
		if err := sourceFile.Close(); err != nil {
			return fmt.Errorf("close bundle file %q: %w", pathOnDisk, err)
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("export bundle %q: %w", bundlePath, walkErr)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("finalize export archive %q: %w", archivePath, err)
	}

	return nil
}
