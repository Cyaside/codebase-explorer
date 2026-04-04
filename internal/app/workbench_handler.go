package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/ui"
)

type workbenchStatusResponse struct {
	AppVersion         string                    `json:"app_version"`
	OutputRoot         string                    `json:"output_root"`
	CacheRoot          string                    `json:"cache_root"`
	CacheEnabled       bool                      `json:"cache_enabled"`
	DefaultProvider    workbenchProviderConfig   `json:"default_provider"`
	SupportedProviders []workbenchProviderOption `json:"supported_providers"`
	RecentBundles      []workbenchBundleSummary  `json:"recent_bundles"`
	BundleWarnings     []string                  `json:"bundle_warnings"`
}

type workbenchProviderOption struct {
	Name            string `json:"name"`
	RequiresAPIKey  bool   `json:"requires_api_key"`
	RequiresModel   bool   `json:"requires_model"`
	RequiresBaseURL bool   `json:"requires_base_url"`
}

type workbenchProviderConfig struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
}

type workbenchAnalyzePayload struct {
	RepoPath            string           `json:"repo_path"`
	DeterministicOnly   bool             `json:"deterministic_only"`
	ExtraIgnorePatterns []string         `json:"extra_ignore_patterns"`
	SupportFiles        []string         `json:"support_files"`
	Provider            *provider.Config `json:"provider"`
}

type workbenchAnalyzeResponse struct {
	Result appAnalyzeResult       `json:"result"`
	Bundle workbenchBundleSummary `json:"bundle"`
	Data   any                    `json:"data"`
}

type appAnalyzeResult = AnalyzeResult

func (s Service) workbenchHandler(outputRoot string) (http.Handler, error) {
	assets, err := ui.Assets()
	if err != nil {
		return nil, fmt.Errorf("load workbench assets: %w", err)
	}

	mux := http.NewServeMux()
	assetHandler := http.FileServer(http.FS(assets))
	bundleHandler := http.StripPrefix("/bundles/", http.FileServer(http.Dir(outputRoot)))

	mux.Handle("/assets/", http.StripPrefix("/assets/", assetHandler))
	mux.Handle("/bundles/", bundleHandler)
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		s.handleWorkbenchStatus(w, r, outputRoot)
	})
	mux.HandleFunc("/api/doctor", s.handleWorkbenchDoctor)
	mux.HandleFunc("/api/analyze", func(w http.ResponseWriter, r *http.Request) {
		s.handleWorkbenchAnalyze(w, r)
	})
	mux.HandleFunc("/api/bundles", func(w http.ResponseWriter, r *http.Request) {
		s.handleWorkbenchBundles(w, r, outputRoot)
	})
	mux.HandleFunc("/api/bundles/", func(w http.ResponseWriter, r *http.Request) {
		s.handleWorkbenchBundle(w, r, outputRoot)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		contents, err := fs.ReadFile(assets, "index.html")
		if err != nil {
			writeWorkbenchError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(contents)
	})

	return mux, nil
}

func (s Service) handleWorkbenchStatus(w http.ResponseWriter, r *http.Request, outputRoot string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bundles, warnings, err := listWorkbenchBundles(outputRoot, s.settings.OutputKeepLatest)
	if err != nil {
		writeWorkbenchError(w, http.StatusInternalServerError, err)
		return
	}

	options := make([]workbenchProviderOption, 0, len(s.providers.Names()))
	for _, name := range s.providers.Names() {
		descriptor, _ := s.providers.Describe(name)
		options = append(options, workbenchProviderOption{
			Name:            descriptor.Name,
			RequiresAPIKey:  descriptor.RequiresAPIKey,
			RequiresModel:   descriptor.RequiresModel,
			RequiresBaseURL: descriptor.RequiresBaseURL,
		})
	}

	response := workbenchStatusResponse{
		AppVersion:   s.settings.AppVersion,
		OutputRoot:   outputRoot,
		CacheRoot:    s.cache.Root(),
		CacheEnabled: s.cache.Enabled(),
		DefaultProvider: workbenchProviderConfig{
			Name:    s.settings.Provider.Name,
			Model:   s.settings.Provider.Model,
			BaseURL: s.settings.Provider.BaseURL,
		},
		SupportedProviders: options,
		RecentBundles:      bundles,
		BundleWarnings:     warnings,
	}

	writeWorkbenchJSON(w, http.StatusOK, response)
}

func (s Service) handleWorkbenchDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := s.Doctor(r.Context(), DoctorRequest{})
	if err != nil {
		writeWorkbenchError(w, http.StatusInternalServerError, err)
		return
	}

	writeWorkbenchJSON(w, http.StatusOK, result)
}

func (s Service) handleWorkbenchAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}

	var payload workbenchAnalyzePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}

	if isRemoteRepository(payload.RepoPath) {
		writeWorkbenchError(w, http.StatusBadRequest, fmt.Errorf("remote repository URLs are not supported yet; analyze a local checkout path"))
		return
	}

	result, err := s.Analyze(r.Context(), AnalyzeRequest{
		RepoPath:             payload.RepoPath,
		DeterministicOnly:    payload.DeterministicOnly,
		ExtraIgnorePatterns:  payload.ExtraIgnorePatterns,
		OptionalSupportFiles: payload.SupportFiles,
		ProviderOverride:     payload.Provider,
	})
	if err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}

	data, err := loadViewerBundleData(result.OutputPath)
	if err != nil {
		writeWorkbenchError(w, http.StatusInternalServerError, err)
		return
	}

	info, err := os.Stat(result.OutputPath)
	if err != nil {
		writeWorkbenchError(w, http.StatusInternalServerError, err)
		return
	}

	response := workbenchAnalyzeResponse{
		Result: result,
		Bundle: summarizeWorkbenchBundle(workbenchBundleLocation{
			name:    filepath.Base(result.OutputPath),
			path:    result.OutputPath,
			modTime: info.ModTime().UTC(),
		}, data),
		Data: data,
	}

	writeWorkbenchJSON(w, http.StatusOK, response)
}

func (s Service) handleWorkbenchBundles(w http.ResponseWriter, r *http.Request, outputRoot string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/api/bundles" {
		http.NotFound(w, r)
		return
	}

	bundles, _, err := listWorkbenchBundles(outputRoot, s.settings.OutputKeepLatest)
	if err != nil {
		writeWorkbenchError(w, http.StatusInternalServerError, err)
		return
	}

	writeWorkbenchJSON(w, http.StatusOK, bundles)
}

func (s Service) handleWorkbenchBundle(w http.ResponseWriter, r *http.Request, outputRoot string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bundleName := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/bundles/"), "/")
	if bundleName == "" || strings.Contains(bundleName, "/") || strings.Contains(bundleName, "\\") {
		http.NotFound(w, r)
		return
	}

	bundle, err := loadWorkbenchBundle(outputRoot, bundleName)
	if err != nil {
		writeWorkbenchError(w, http.StatusNotFound, err)
		return
	}

	writeWorkbenchJSON(w, http.StatusOK, bundle)
}

func writeWorkbenchJSON(w http.ResponseWriter, statusCode int, value any) {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(append(contents, '\n'))
}

func writeWorkbenchError(w http.ResponseWriter, statusCode int, err error) {
	writeWorkbenchJSON(w, statusCode, map[string]string{
		"error": strings.TrimSpace(err.Error()),
	})
}

func isRemoteRepository(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "git@")
}
