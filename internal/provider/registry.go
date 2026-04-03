package provider

import (
	"fmt"
	"slices"
	"strings"
)

type Config struct {
	Name    string
	Model   string
	APIKey  string
	BaseURL string
}

func (c Config) Enabled() bool {
	return strings.TrimSpace(c.Name) != ""
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
		"openai": {
			Name:           "openai",
			RequiresAPIKey: true,
			RequiresModel:  true,
		},
		"openai-compatible": {
			Name:            "openai-compatible",
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
	if !config.Enabled() {
		return nil
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

	return nil
}

func (r Registry) Describe(name string) (Descriptor, bool) {
	descriptor, found := r.descriptors[normalizeName(name)]
	return descriptor, found
}

func (r Registry) ClientFor(config Config) (Client, error) {
	if err := r.Validate(config); err != nil {
		return nil, err
	}

	switch normalizeName(config.Name) {
	case "openai", "openai-compatible":
		client := NewOpenAICompatibleClient(nil)
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported provider %q; supported providers: %s", config.Name, strings.Join(r.Names(), ", "))
	}
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
