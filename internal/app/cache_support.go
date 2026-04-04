package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type deterministicState struct {
	ScanResult repo.ScanResult
	Analysis   analyzer.Result
	Changes    changes.Result
	Status     string
}

func (s Service) loadDeterministicState(ctx context.Context, request AnalyzeRequest, repoPath string, supportFiles []string) (deterministicState, error) {
	if !s.cache.Enabled() {
		emitAnalyzeProgress(request, "cache", cache.StatusDisabled, "deterministic cache disabled")
		return s.computeDeterministicState(ctx, request, repoPath, supportFiles, cache.StatusDisabled)
	}

	cacheKey, err := cache.BuildDeterministicKey(repoPath, request.ExtraIgnorePatterns, supportFiles, s.settings.AppVersion, request.DeterministicOnly)
	if err != nil {
		emitAnalyzeProgress(request, "cache", "bypass", "could not build deterministic cache key; continuing without cache")
		return s.computeDeterministicState(ctx, request, repoPath, supportFiles, cache.StatusMiss)
	}

	payload, err := s.cache.LoadDeterministic(cacheKey)
	switch {
	case err == nil:
		emitAnalyzeProgress(request, "cache", cache.StatusHit, "reused cached scan and deterministic analysis")
		return deterministicState{
			ScanResult: payload.Scan,
			Analysis:   payload.Analysis,
			Changes:    payload.Changes,
			Status:     cache.StatusHit,
		}, nil
	case errors.Is(err, cache.ErrCacheMiss):
		emitAnalyzeProgress(request, "cache", cache.StatusMiss, "repository snapshot changed or has not been cached yet")
	default:
		emitAnalyzeProgress(request, "cache", "bypass", "cache lookup failed; continuing with fresh analysis")
		return s.computeDeterministicState(ctx, request, repoPath, supportFiles, cache.StatusMiss)
	}

	state, stateErr := s.computeDeterministicState(ctx, request, repoPath, supportFiles, cache.StatusMiss)
	if stateErr != nil {
		return deterministicState{}, stateErr
	}

	saveErr := s.cache.SaveDeterministic(cache.DeterministicPayload{
		CachedAt:   time.Now().UTC(),
		Key:        cacheKey,
		Scan:       state.ScanResult,
		Analysis:   state.Analysis,
		Changes:    state.Changes,
		AppVersion: s.settings.AppVersion,
	})
	if saveErr != nil {
		emitAnalyzeProgress(request, "cache", "store-skipped", "deterministic cache could not be written")
	}

	return state, nil
}

func (s Service) computeDeterministicState(ctx context.Context, request AnalyzeRequest, repoPath string, supportFiles []string, status string) (deterministicState, error) {
	emitAnalyzeProgress(request, "scan", "running", fmt.Sprintf("scanning %s", repoPath))
	scanResult, err := s.scanner.Scan(ctx, repo.ScanOptions{
		RootPath:            repoPath,
		ExtraIgnorePatterns: request.ExtraIgnorePatterns,
	})
	if err != nil {
		return deterministicState{}, err
	}
	emitAnalyzeProgress(request, "scan", "succeeded", fmt.Sprintf("scanned %d file(s)", len(scanResult.Files)))

	emitAnalyzeProgress(request, "deterministic-analysis", "running", "building repository summary and heuristics")
	analysis := s.analyzer.Analyze(scanResult, request.DeterministicOnly)
	emitAnalyzeProgress(request, "deterministic-analysis", "succeeded", fmt.Sprintf("identified %d file(s) and %d line(s)", analysis.Metrics.TotalFiles, analysis.Metrics.TotalLines))

	emitAnalyzeProgress(request, "change-awareness", "running", fmt.Sprintf("correlating %d support file(s)", len(supportFiles)))
	changeResult := changes.Analyze(analysis.GeneratedAt, analysis, supportFiles)
	emitAnalyzeProgress(request, "change-awareness", "succeeded", changeResult.Note)

	return deterministicState{
		ScanResult: scanResult,
		Analysis:   analysis,
		Changes:    changeResult,
		Status:     status,
	}, nil
}
