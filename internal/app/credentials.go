package app

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/Cyaside/codebase-explorer/internal/provider"
)

var credentialIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

type savedCredential struct {
	Label   string `json:"label"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

type credentialStore struct {
	mu   sync.Mutex
	path string
}

func newCredentialStore(path string) *credentialStore { return &credentialStore{path: path} }

func (store *credentialStore) filePath() (string, error) {
	if store.path != "" {
		return store.path, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user configuration directory: %w", err)
	}
	return filepath.Join(configDir, "codebase-explorer", "credentials.json"), nil
}

func (store *credentialStore) readLocked() (map[string]savedCredential, error) {
	path, err := store.filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]savedCredential{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read saved connections: %w", err)
	}
	var saved map[string]savedCredential
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, fmt.Errorf("decode saved connections: %w", err)
	}
	if saved == nil {
		saved = map[string]savedCredential{}
	}
	return saved, nil
}

func (store *credentialStore) writeLocked(saved map[string]savedCredential) error {
	path, err := store.filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create connection directory: %w", err)
	}
	data, err := json.Marshal(saved)
	if err != nil {
		return fmt.Errorf("encode saved connections: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".credentials-*")
	if err != nil {
		return fmt.Errorf("stage saved connections: %w", err)
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return fmt.Errorf("secure saved connections: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write saved connections: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync saved connections: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close saved connections: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("replace saved connections: %w", err)
	}
	return nil
}

type savedCredentialInfo struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
}

func (store *credentialStore) list() ([]savedCredentialInfo, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	saved, err := store.readLocked()
	if err != nil {
		return nil, err
	}
	items := make([]savedCredentialInfo, 0, len(saved))
	for id, credential := range saved {
		items = append(items, savedCredentialInfo{ID: id, Label: credential.Label, Model: credential.Model, BaseURL: credential.BaseURL})
	}
	slices.SortFunc(items, func(a, b savedCredentialInfo) int { return strings.Compare(a.ID, b.ID) })
	return items, nil
}

func (store *credentialStore) save(id, label, model, baseURL, apiKey string) error {
	if !credentialIDPattern.MatchString(id) {
		return fmt.Errorf("connection ID must contain only letters, numbers, underscore, or hyphen")
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	label = strings.TrimSpace(label)
	model = strings.TrimSpace(model)
	apiKey = strings.TrimSpace(apiKey)
	if label == "" {
		return fmt.Errorf("connection label is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid connection base URL")
	}
	if parsed.Scheme == "http" {
		host := parsed.Hostname()
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return fmt.Errorf("saved API keys require HTTPS except for loopback endpoints")
		}
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	saved, err := store.readLocked()
	if err != nil {
		return err
	}
	if apiKey == "" {
		previous, exists := saved[id]
		if !exists || previous.BaseURL != baseURL {
			return fmt.Errorf("API key is required for a new or changed base URL")
		}
		apiKey = previous.APIKey
	}
	validationModel := model
	if validationModel == "" {
		validationModel = "model-list"
	}
	if err := provider.NewRegistry().Validate(provider.Config{Name: "compatible", Model: validationModel, APIKey: apiKey, BaseURL: baseURL}); err != nil {
		return err
	}
	saved[id] = savedCredential{Label: label, Model: model, APIKey: apiKey, BaseURL: baseURL}
	return store.writeLocked(saved)
}

func (store *credentialStore) remove(id string) error {
	if !credentialIDPattern.MatchString(id) {
		return fmt.Errorf("invalid connection ID")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	saved, err := store.readLocked()
	if err != nil {
		return err
	}
	delete(saved, id)
	return store.writeLocked(saved)
}

func (store *credentialStore) resolve(id string, config *provider.Config) (*provider.Config, error) {
	if id == "" {
		return config, nil
	}
	if !credentialIDPattern.MatchString(id) || config == nil {
		return nil, fmt.Errorf("invalid saved connection request")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	saved, err := store.readLocked()
	if err != nil {
		return nil, err
	}
	credential, ok := saved[id]
	if !ok {
		return nil, fmt.Errorf("saved API key for connection %q was not found", id)
	}
	if strings.TrimSpace(config.APIKey) != "" {
		return nil, fmt.Errorf("send a saved connection ID or an API key, not both")
	}
	if strings.TrimRight(strings.TrimSpace(config.BaseURL), "/") != credential.BaseURL {
		return nil, fmt.Errorf("saved API key is bound to a different base URL")
	}
	resolved := *config
	resolved.APIKey = credential.APIKey
	return &resolved, nil
}

func (store *credentialStore) connection(id string) (provider.Config, error) {
	if !credentialIDPattern.MatchString(id) {
		return provider.Config{}, fmt.Errorf("invalid saved connection ID")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	saved, err := store.readLocked()
	if err != nil {
		return provider.Config{}, err
	}
	credential, exists := saved[id]
	if !exists {
		return provider.Config{}, fmt.Errorf("saved connection %q was not found", id)
	}
	return provider.Config{Name: "compatible", Model: credential.Model, BaseURL: credential.BaseURL, APIKey: credential.APIKey}, nil
}

func localCredentialRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin == "http://"+r.Host || origin == "https://"+r.Host
	}
	return true
}
