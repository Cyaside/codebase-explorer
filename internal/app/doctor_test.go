package app

import (
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestDoctorFailsWhenProviderConfigIsInvalid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name: "openai",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	providerCheck := findDoctorCheck(result, "provider")
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "fail" {
		t.Fatalf("expected invalid provider config to fail doctor, got %#v", providerCheck)
	}
}

func TestDoctorPassesWhenProviderConfigIsValid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: "https://example.com/v1",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	providerCheck := findDoctorCheck(result, "provider")
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "pass" {
		t.Fatalf("expected valid provider config to pass doctor, got %#v", providerCheck)
	}
}

func findDoctorCheck(result DoctorResult, name string) *DoctorCheck {
	for index := range result.Checks {
		if result.Checks[index].Name == name {
			return &result.Checks[index]
		}
	}
	return nil
}
