package repo

import "testing"

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
