package analyzer

import "testing"

func TestDetectLanguage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		extension string
		want      string
	}{
		{
			name:      "go source",
			extension: ".go",
			want:      "Go",
		},
		{
			name:      "typescript source",
			extension: ".ts",
			want:      "TypeScript",
		},
		{
			name:      "case insensitive",
			extension: ".JS",
			want:      "JavaScript",
		},
		{
			name:      "unknown extension falls back to other",
			extension: ".custom",
			want:      "Other",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := detectLanguage(testCase.extension)
			if got != testCase.want {
				t.Fatalf("detectLanguage(%q) = %q, want %q", testCase.extension, got, testCase.want)
			}
		})
	}
}
