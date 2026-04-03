package provider

import (
	"testing"
	"time"
)

func TestDisabledResultMarksAIAsUnused(t *testing.T) {
	t.Parallel()

	result := DisabledResult(time.Date(2026, time.April, 3, 15, 0, 0, 0, time.UTC), "no provider configured")
	if result.SchemaVersion != ResultSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", ResultSchemaVersion, result.SchemaVersion)
	}
	if result.Status != ResultStatusDisabled {
		t.Fatalf("expected disabled status, got %q", result.Status)
	}
	if result.Used {
		t.Fatalf("expected disabled result to mark AI as unused")
	}
}

func TestFallbackResultCarriesProviderIdentity(t *testing.T) {
	t.Parallel()

	result := FallbackResult(time.Now().UTC(), Config{
		Name:  "openai-compatible",
		Model: "gpt-4.1-mini",
	}, "adapter not wired yet")

	if result.Provider != "openai-compatible" {
		t.Fatalf("expected provider identity to be preserved, got %#v", result)
	}
	if result.Model != "gpt-4.1-mini" {
		t.Fatalf("expected model identity to be preserved, got %#v", result)
	}
	if result.Status != ResultStatusFallback {
		t.Fatalf("expected fallback status, got %#v", result)
	}
	if result.Successful() {
		t.Fatalf("did not expect fallback result to be successful")
	}
}
