package provider

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

type Config struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func (c Config) Enabled() bool {
	return strings.TrimSpace(c.Name) != ""
}

// Canonical preserves old connection names while exposing one provider contract.
func (c Config) Canonical() Config {
	c.Name = normalizeName(c.Name)
	c.Model = strings.TrimSpace(c.Model)
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if c.Name == "openai" {
		c.Name = "compatible"
		if c.BaseURL == "" {
			c.BaseURL = "https://api.openai.com/v1"
		}
	} else if c.Name == "openai-compatible" {
		c.Name = "compatible"
	}
	return c
}

type Descriptor struct {
	Name            string
	RequiresAPIKey  bool
	RequiresModel   bool
	RequiresBaseURL bool
}

type Registry struct {
	descriptors map[string]Descriptor
}

func NewRegistry() Registry {
	descriptors := map[string]Descriptor{
		"compatible": {
			Name:            "compatible",
			RequiresAPIKey:  true,
			RequiresModel:   true,
			RequiresBaseURL: true,
		},
	}

	return Registry{descriptors: descriptors}
}

func (r Registry) Names() []string {
	names := make([]string, 0, len(r.descriptors))
	for name := range r.descriptors {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (r Registry) Validate(config Config) error {
	config = config.Canonical()
	if !config.Enabled() {
		return fmt.Errorf("compatible connection is required: configure a base URL, model, and API key")
	}

	name := normalizeName(config.Name)
	descriptor, found := r.descriptors[name]
	if !found {
		return fmt.Errorf("unsupported provider %q; supported providers: %s", config.Name, strings.Join(r.Names(), ", "))
	}
	if descriptor.RequiresModel && strings.TrimSpace(config.Model) == "" {
		return fmt.Errorf("provider %q requires a model", descriptor.Name)
	}
	if descriptor.RequiresAPIKey && strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("provider %q requires an API key", descriptor.Name)
	}
	if descriptor.RequiresBaseURL && strings.TrimSpace(config.BaseURL) == "" {
		return fmt.Errorf("provider %q requires a base URL", descriptor.Name)
	}
	parsed, err := url.Parse(config.BaseURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("provider %q requires an HTTP(S) base URL without credentials or query", descriptor.Name)
	}

	return nil
}

func (r Registry) Describe(name string) (Descriptor, bool) {
	descriptor, found := r.descriptors[(Config{Name: name}).Canonical().Name]
	return descriptor, found
}

func (r Registry) ClientFor(config Config) (Client, error) {
	config = config.Canonical()
	if err := r.Validate(config); err != nil {
		return nil, err
	}

	switch normalizeName(config.Name) {
	case "compatible":
		client := NewOpenAICompatibleClient(nil)
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported provider %q; supported providers: %s", config.Name, strings.Join(r.Names(), ", "))
	}
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
