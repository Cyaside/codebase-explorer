package app

import (
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestWorkbenchRejectsNonLoopbackBindWhenCredentialsAreAvailable(t *testing.T) {
	t.Parallel()
	service := New(config.Settings{DefaultOutputRoot: t.TempDir()})
	err := service.Serve(t.Context(), ServeRequest{Addr: "0.0.0.0:0", NoBrowser: true})
	if err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("expected non-loopback address to be rejected, got %v", err)
	}
}
