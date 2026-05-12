package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/provider"
)

type saveCredentialPayload struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

func (s Service) handleWorkbenchCredentials(w http.ResponseWriter, r *http.Request) {
	if !localCredentialRequest(r) {
		writeWorkbenchError(w, http.StatusForbidden, fmt.Errorf("saved connections are available only from this computer and same-origin pages"))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	switch r.Method {
	case http.MethodGet:
		connections, err := s.credentials.list()
		if err != nil {
			writeWorkbenchError(w, http.StatusInternalServerError, err)
			return
		}
		writeWorkbenchJSON(w, http.StatusOK, map[string]any{"connections": connections})
	case http.MethodPost:
		var payload saveCredentialPayload
		decoder := json.NewDecoder(io.LimitReader(r.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeWorkbenchError(w, http.StatusBadRequest, fmt.Errorf("decode connection: %w", err))
			return
		}
		if err := s.credentials.save(payload.ID, payload.Label, payload.Model, payload.BaseURL, payload.APIKey); err != nil {
			writeWorkbenchError(w, http.StatusBadRequest, err)
			return
		}
		writeWorkbenchJSON(w, http.StatusOK, savedCredentialInfo{ID: payload.ID, Label: strings.TrimSpace(payload.Label), Model: strings.TrimSpace(payload.Model), BaseURL: strings.TrimRight(strings.TrimSpace(payload.BaseURL), "/")})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s Service) handleWorkbenchCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !localCredentialRequest(r) {
		writeWorkbenchError(w, http.StatusForbidden, fmt.Errorf("saved connections are available only from this computer and same-origin pages"))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	id := strings.TrimPrefix(r.URL.Path, "/api/credentials/")
	if err := s.credentials.remove(id); err != nil {
		writeWorkbenchError(w, http.StatusBadRequest, err)
		return
	}
	writeWorkbenchJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (s Service) resolveWorkbenchAnalyzeConnection(r *http.Request, payload *workbenchAnalyzePayload) error {
	if payload.CredentialID != "" && !localCredentialRequest(r) {
		return fmt.Errorf("saved connections are available only from this computer and same-origin pages")
	}
	config, err := s.credentials.resolve(payload.CredentialID, payload.Provider)
	if err != nil {
		return fmt.Errorf("connection: %w", err)
	}
	payload.Provider = config
	return nil
}

func (s Service) resolveWorkbenchDiagnosticConnection(r *http.Request, config provider.Config, credentialID string) (provider.Config, error) {
	if credentialID != "" && !localCredentialRequest(r) {
		return provider.Config{}, fmt.Errorf("saved connections are available only from this computer and same-origin pages")
	}
	resolved, err := s.credentials.resolve(credentialID, &config)
	if err != nil {
		return provider.Config{}, fmt.Errorf("connection: %w", err)
	}
	return *resolved, nil
}
