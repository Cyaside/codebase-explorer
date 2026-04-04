package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/bundle"
	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type Service struct {
	settings  config.Settings
	scanner   repo.Scanner
	analyzer  analyzer.Service
	providers provider.Registry
	writer    bundle.Writer
}

func New(settings config.Settings) Service {
	return Service{
		settings:  settings,
		scanner:   repo.NewScanner(),
		analyzer:  analyzer.NewService(settings.AppVersion),
		providers: provider.NewRegistry(),
		writer:    bundle.NewWriter(settings.AppVersion, settings.OutputKeepLatest),
	}
}

func (s Service) Analyze(ctx context.Context, request AnalyzeRequest) (AnalyzeResult, error) {
	repoPath, err := resolveRepoPath(request.RepoPath)
	if err != nil {
		return AnalyzeResult{}, err
	}

	outputRoot, err := s.resolveOutputRoot(request.OutputRoot)
	if err != nil {
		return AnalyzeResult{}, err
	}

	scanResult, err := s.scanner.Scan(ctx, repo.ScanOptions{
		RootPath:            repoPath,
		ExtraIgnorePatterns: request.ExtraIgnorePatterns,
	})
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("scan repository: %w", err)
	}

	analysis := s.analyzer.Analyze(scanResult, request.DeterministicOnly)
	aiContext := buildCondensedContext(analysis)
	emitAnalyzeProgress(request, "ai-context", "ready", summarizeAIContext(aiContext))
	aiResult := s.buildAIResult(ctx, request, analysis, request.DeterministicOnly, aiContext)

	writeResult, err := s.writer.Write(bundle.WriteRequest{
		OutputRoot:        outputRoot,
		DeterministicOnly: request.DeterministicOnly,
		ScanResult:        scanResult,
		Analysis:          analysis,
		AIContext:         aiContext,
		AIResult:          aiResult,
	})
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("write bundle: %w", err)
	}
	emitAnalyzeProgress(request, "output-cleanup", cleanupStatus(writeResult.PrunedBundles), cleanupDetail(writeResult.PrunedBundles, writeResult.RetentionLimit))

	primaryLanguage := ""
	if len(analysis.Languages) > 0 {
		primaryLanguage = analysis.Languages[0].Name
	}

	return AnalyzeResult{
		OutputPath:      writeResult.BundlePath,
		ProjectName:     analysis.ProjectName,
		ProjectType:     analysis.ProjectType,
		TotalFiles:      analysis.Metrics.TotalFiles,
		TotalLines:      analysis.Metrics.TotalLines,
		EntryPoints:     analysis.EntryPoints,
		PrimaryLanguage: primaryLanguage,
		AI:              buildAISummary(aiContext, aiResult),
		Output: AnalyzeOutputSummary{
			RetentionLimit: writeResult.RetentionLimit,
			PrunedBundles:  writeResult.PrunedBundles,
		},
	}, nil
}

func (s Service) Doctor(_ context.Context, _ DoctorRequest) (DoctorResult, error) {
	outputRoot, err := s.resolveOutputRoot("")
	if err != nil {
		return DoctorResult{}, err
	}

	checks := []DoctorCheck{
		{
			Name:   "config",
			Status: "pass",
			Detail: fmt.Sprintf("config loaded from %s", s.settings.ConfigSource),
		},
	}

	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		checks = append(checks, DoctorCheck{
			Name:   "output-root",
			Status: "fail",
			Detail: fmt.Sprintf("cannot prepare output root %q: %v", outputRoot, err),
		})
	} else {
		checks = append(checks, DoctorCheck{
			Name:   "output-root",
			Status: "pass",
			Detail: fmt.Sprintf("output root is ready at %s; keeping latest %d bundle(s)", outputRoot, s.settings.OutputKeepLatest),
		})
	}

	checks = append(checks, s.providerDoctorCheck())

	return DoctorResult{
		ConfigSource: s.settings.ConfigSource,
		OutputRoot:   outputRoot,
		Checks:       checks,
	}, nil
}

func (s Service) providerDoctorCheck() DoctorCheck {
	providerConfig := s.providerConfig()
	if !providerConfig.Enabled() {
		return DoctorCheck{
			Name:   "provider",
			Status: "pass",
			Detail: "no provider configured; deterministic analysis remains available",
		}
	}

	if err := s.providers.Validate(providerConfig); err != nil {
		return DoctorCheck{
			Name:   "provider",
			Status: "fail",
			Detail: err.Error(),
		}
	}

	detail := fmt.Sprintf("provider %s is configured with model %s", providerConfig.Name, providerConfig.Model)
	if strings.TrimSpace(providerConfig.BaseURL) != "" {
		detail += fmt.Sprintf(" via %s", providerConfig.BaseURL)
	}

	return DoctorCheck{
		Name:   "provider",
		Status: "pass",
		Detail: detail,
	}
}

func resolveRepoPath(repoPath string) (string, error) {
	if strings.TrimSpace(repoPath) == "" {
		return "", fmt.Errorf("repository path is required")
	}

	absolutePath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("resolve repository path %q: %w", repoPath, err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("repository path %q: %w", absolutePath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository path %q is not a directory", absolutePath)
	}

	return absolutePath, nil
}

func (s Service) resolveOutputRoot(outputRoot string) (string, error) {
	root := strings.TrimSpace(outputRoot)
	if root == "" {
		root = s.settings.DefaultOutputRoot
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve output root %q: %w", root, err)
	}

	return absoluteRoot, nil
}
