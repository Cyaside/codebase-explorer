package repo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanExcludesCredentialFilesFromMetadata(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, content := range map[string]string{"README.md": "project", ".env": "API_KEY=secret", "credentials.json": `{"key":"secret"}`} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	result, err := NewScanner().Scan(context.Background(), ScanOptions{RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "README.md" {
		t.Fatalf("credential paths entered scan metadata: %#v", result.Files)
	}
}

func TestMarkerCountOnlyMatchesStandaloneMarkers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		line   string
		marker string
		want   int
	}{
		{
			name:   "counts comment marker",
			line:   "// TODO: wire provider layer later",
			marker: "TODO",
			want:   1,
		},
		{
			name:   "does not count identifier",
			line:   "TodoCount int",
			marker: "TODO",
			want:   0,
		},
		{
			name:   "counts mixed case marker",
			line:   "# fixme keep fallback path explicit",
			marker: "FIXME",
			want:   1,
		},
		{
			name:   "counts repeated standalone markers",
			line:   "HACK: temporary bridge // hack revisit after phase 1",
			marker: "HACK",
			want:   2,
		},
		{
			name:   "does not count marker embedded in identifier",
			line:   "hackScore := 10",
			marker: "HACK",
			want:   0,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := markerCount(testCase.line, testCase.marker)
			if got != testCase.want {
				t.Fatalf("markerCount(%q, %q) = %d, want %d", testCase.line, testCase.marker, got, testCase.want)
			}
		})
	}
}
