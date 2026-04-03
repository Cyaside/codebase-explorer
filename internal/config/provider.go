package config

import "strings"

type ProviderSettings struct {
	Name    string
	Model   string
	APIKey  string
	BaseURL string
}

func (s ProviderSettings) Enabled() bool {
	return strings.TrimSpace(s.Name) != ""
}

func loadProviderSettings() (ProviderSettings, bool) {
	settings := ProviderSettings{
		Name:    strings.TrimSpace(getEnv("CODEARCH_PROVIDER")),
		Model:   strings.TrimSpace(getEnv("CODEARCH_MODEL")),
		APIKey:  strings.TrimSpace(getEnv("CODEARCH_API_KEY")),
		BaseURL: strings.TrimSpace(getEnv("CODEARCH_BASE_URL")),
	}

	return settings, settings.Enabled() || settings.Model != "" || settings.APIKey != "" || settings.BaseURL != ""
}
