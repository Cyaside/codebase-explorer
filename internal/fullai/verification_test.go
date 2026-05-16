package fullai

import (
	"testing"
	"time"
)

func TestVerifyFunctionOutputDetailedAcceptsFlowchartEdges(t *testing.T) {
	t.Parallel()

	output, report := VerifyFunctionOutputDetailed(FunctionJob{
		Name:          "flowchart",
		EvidencePaths: []string{"cmd/main.go"},
	}, FunctionOutput{
		GraphEdges: []GraphEdge{
			{From: "CLI", To: "Analyzer", Label: "runs", EvidencePaths: []string{"cmd/main.go"}},
		},
	})

	if !report.Verified {
		t.Fatalf("expected verified flowchart report, got %#v", report)
	}
	if output.GraphEdges[0].EvidencePaths[0] != "cmd/main.go" {
		t.Fatalf("expected accepted evidence to remain, got %#v", output)
	}
	if len(report.AcceptedEvidencePaths) != 1 || report.AcceptedEvidencePaths[0] != "cmd/main.go" {
		t.Fatalf("expected accepted evidence to be recorded, got %#v", report.AcceptedEvidencePaths)
	}
}

func TestVerifyFunctionOutputDetailedRejectsInventedGraphFile(t *testing.T) {
	t.Parallel()
	output, report := VerifyFunctionOutputDetailed(FunctionJob{Name: "flowchart", EvidencePaths: []string{"cmd/main.go"}}, FunctionOutput{
		GraphEdges: []GraphEdge{{From: "missing/file.go", To: "CLI", EvidencePaths: []string{"cmd/main.go"}}},
	})
	if report.Verified || len(output.GraphEdges) != 0 || !hasVerificationCheck(report, "graph.endpoint", "warn") || !hasVerificationCheck(report, "graph.edges", "fail") {
		t.Fatalf("invented graph endpoint should fail path checks: %#v", report)
	}
}

func TestVerifyFunctionOutputDetailedKeepsSupportedFlowchartEdges(t *testing.T) {
	t.Parallel()
	output, report := VerifyFunctionOutputDetailed(FunctionJob{Name: "flowchart", EvidencePaths: []string{"cmd/main.go"}}, FunctionOutput{
		GraphEdges: []GraphEdge{
			{From: "CLI", To: "Analyzer", EvidencePaths: []string{"cmd/main.go"}},
			{From: "missing/file.go", To: "CLI", EvidencePaths: []string{"cmd/main.go"}},
		},
	})
	if !report.Verified || len(output.GraphEdges) != 1 || !hasVerificationCheck(report, "graph.endpoint", "warn") {
		t.Fatalf("supported edge should remain while invented edge is removed: output=%#v report=%#v", output, report)
	}
}

func TestVerifyFunctionOutputDetailedRejectsInvalidIssueSeverity(t *testing.T) {
	t.Parallel()

	_, report := VerifyFunctionOutputDetailed(FunctionJob{
		Name:          "issues",
		EvidencePaths: []string{"internal/app/app.go"},
	}, FunctionOutput{
		IssueSignals: []IssueSignal{
			{Title: "missing timeout", Severity: "urgent", EvidencePaths: []string{"internal/app/app.go"}},
		},
	})

	if report.Verified {
		t.Fatalf("expected invalid severity to fail verification")
	}
	if !hasVerificationCheck(report, "issues.severity", "fail") {
		t.Fatalf("expected failed severity check, got %#v", report.Checks)
	}
}

func TestVerifyFunctionOutputDetailedRecordsRejectedEvidence(t *testing.T) {
	t.Parallel()

	output, report := VerifyFunctionOutputDetailed(FunctionJob{
		Name:          "summary",
		EvidencePaths: []string{"README.md"},
	}, FunctionOutput{
		Summary: "project summary",
		KeyFindings: []Finding{
			{Claim: "claim", EvidencePaths: []string{"README.md", "missing.go"}},
		},
	})

	if report.Verified {
		t.Fatalf("expected rejected evidence to fail verification")
	}
	if len(output.KeyFindings[0].EvidencePaths) != 1 {
		t.Fatalf("expected rejected evidence to be filtered, got %#v", output.KeyFindings[0].EvidencePaths)
	}
	if len(report.RejectedEvidencePaths) != 1 || report.RejectedEvidencePaths[0] != "missing.go" {
		t.Fatalf("expected rejected evidence to be recorded, got %#v", report.RejectedEvidencePaths)
	}
}

func TestRecommendationsWithoutEvidenceAreNotAccepted(t *testing.T) {
	t.Parallel()
	_, report := VerifyFunctionOutputDetailed(FunctionJob{Name: "recommendations", EvidencePaths: []string{"README.md"}}, FunctionOutput{
		Recommendations: []string{"Rewrite the service boundary."},
		Uncertainties:   []string{"The dependency graph is incomplete."},
	})
	if report.Verified || !hasVerificationCheck(report, "recommendations.grounded", "fail") {
		t.Fatalf("ungrounded recommendations must not enter the main result: %#v", report)
	}
}

func TestNewVerificationAggregatesFunctionReports(t *testing.T) {
	t.Parallel()

	report := NewVerification(time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC), string(ModeFull), []FunctionResult{
		{
			Name:     "summary",
			Verified: true,
			Verification: FunctionVerification{
				Name:     "summary",
				Status:   "verified",
				Verified: true,
				Checks:   []VerificationCheck{{Name: "summary.present", Status: "pass"}},
			},
		},
		{
			Name: "flowchart",
			Verification: FunctionVerification{
				Name:     "flowchart",
				Status:   "unverified",
				Verified: false,
				Checks:   []VerificationCheck{{Name: "graph.edges", Status: "fail"}},
			},
		},
	})

	if report.Status != "needs-review" {
		t.Fatalf("expected needs-review status, got %q", report.Status)
	}
	if report.VerifiedCount != 1 || report.FailedCheckCount != 1 {
		t.Fatalf("expected aggregate counts, got %#v", report)
	}
}

func hasVerificationCheck(report FunctionVerification, name string, status string) bool {
	for _, check := range report.Checks {
		if check.Name == name && check.Status == status {
			return true
		}
	}
	return false
}
