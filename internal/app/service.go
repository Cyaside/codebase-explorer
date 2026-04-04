package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/bundle"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type Service struct {
	settings  config.Settings
	scanner   repo.Scanner
	analyzer  analyzer.Service
	cache     cache.Store
	providers provider.Registry
	writer    bundle.Writer
}

func New(settings config.Settings) Service {
	cacheRoot := strings.TrimSpace(settings.CacheRoot)
	cacheEnabled := settings.CacheEnabled
	if cacheRoot == "" {
		if strings.TrimSpace(settings.DefaultOutputRoot) != "" {
			cacheRoot = filepath.Join(settings.DefaultOutputRoot, ".codearch-cache")
		} else {
			cacheRoot = ".codearch-cache"
		}
		if !settings.CacheEnabled {
			cacheEnabled = true
		}
	}

	return Service{
		settings:  settings,
		scanner:   repo.NewScanner(),
		analyzer:  analyzer.NewService(settings.AppVersion),
		cache:     cache.NewStore(cacheRoot, cacheEnabled),
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

	supportFiles := resolveSupportFiles(repoPath, request.OptionalSupportFiles)
	state, err := s.loadDeterministicState(ctx, request, repoPath, supportFiles)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("scan repository: %w", err)
	}

	scanResult := state.ScanResult
	analysis := state.Analysis
	changeResult := state.Changes
	aiContext := buildCondensedContext(analysis)
	emitAnalyzeProgress(request, "ai-context", "ready", summarizeAIContext(aiContext))
	aiResult, providerCacheStatus := s.buildAIResult(ctx, request, analysis, request.DeterministicOnly, aiContext)

	writeResult, err := s.writer.Write(bundle.WriteRequest{
		OutputRoot:        outputRoot,
		DeterministicOnly: request.DeterministicOnly,
		ScanResult:        scanResult,
		Analysis:          analysis,
		Changes:           changeResult,
		Cache: bundle.CacheMeta{
			Enabled:             s.cache.Enabled(),
			Root:                s.cache.Root(),
			DeterministicStatus: state.Status,
			ProviderStatus:      providerCacheStatus,
		},
		AIContext: aiContext,
		AIResult:  aiResult,
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
		Cache: AnalyzeCacheSummary{
			Enabled:             s.cache.Enabled(),
			Root:                s.cache.Root(),
			DeterministicStatus: state.Status,
			ProviderStatus:      providerCacheStatus,
		},
		Changes: AnalyzeChangesSummary{
			Available:        changeResult.Available,
			SupportFileCount: changeResult.SupportFileCount,
			ParsedItemCount:  changeResult.ParsedItemCount,
			MentionedAreas:   len(changeResult.FrequentlyMentionedAreas),
			Note:             changeResult.Note,
		},
		AI: buildAISummary(aiContext, aiResult),
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

	if err := s.cache.Ensure(); err != nil {
		checks = append(checks, DoctorCheck{
			Name:   "cache",
			Status: "fail",
			Detail: fmt.Sprintf("cannot prepare cache root %q: %v", s.cache.Root(), err),
		})
	} else {
		status := "enabled"
		if !s.cache.Enabled() {
			status = "disabled"
		}
		checks = append(checks, DoctorCheck{
			Name:   "cache",
			Status: "pass",
			Detail: fmt.Sprintf("cache %s at %s", status, s.cache.Root()),
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
