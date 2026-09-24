package provider

import "testing"

func TestRegistryValidateOpenAICompatibleConfig(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()

	if err := registry.Validate(Config{
		Name:    "openai-compatible",
		Model:   "gpt-4.1-mini",
		APIKey:  "test-key",
		BaseURL: "https://example.com/v1",
	}); err != nil {
		t.Fatalf("expected compatible provider config to validate: %v", err)
	}
}

func TestRegistryUsesOnePublicProviderAndMigratesOpenAI(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	if names := registry.Names(); len(names) != 1 || names[0] != "compatible" {
		t.Fatalf("expected one public provider, got %#v", names)
	}
	legacy := Config{Name: "openai", Model: "fixture-model", APIKey: "test-key"}.Canonical()
	if legacy.Name != "compatible" || legacy.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected old OpenAI connection to migrate, got %#v", legacy)
	}
	if err := registry.Validate(legacy); err != nil {
		t.Fatalf("expected migrated connection to validate: %v", err)
	}
}

func TestRegistryValidateRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()

	testCases := []struct {
		name   string
		config Config
	}{
		{
			name: "openai requires model",
			config: Config{
				Name:   "openai",
				APIKey: "test-key",
			},
		},
		{
			name: "openai-compatible requires base url",
			config: Config{
				Name:   "openai-compatible",
				Model:  "gpt-4.1-mini",
				APIKey: "test-key",
			},
		},
		{
			name: "unsupported provider rejected",
			config: Config{
				Name:   "custom",
				Model:  "model",
				APIKey: "test-key",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if err := registry.Validate(testCase.config); err == nil {
				t.Fatalf("expected validation failure for %#v", testCase.config)
			}
		})
	}
}
