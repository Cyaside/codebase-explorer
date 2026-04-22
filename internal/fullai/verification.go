package fullai

import (
	"math"
	"strings"
	"time"
)

const VerificationSchemaVersion = "full-ai-verification.v1"

type Verification struct {
	SchemaVersion    string                 `json:"schema_version"`
	GeneratedAt      time.Time              `json:"generated_at"`
	Mode             string                 `json:"mode"`
	Status           string                 `json:"status"`
	VerifiedCount    int                    `json:"verified_count"`
	WarningCount     int                    `json:"warning_count"`
	FailedCheckCount int                    `json:"failed_check_count"`
	Functions        []FunctionVerification `json:"functions"`
	Note             string                 `json:"note,omitempty"`
}

type FunctionVerification struct {
	Name                  string              `json:"name"`
	Status                string              `json:"status"`
	Verified              bool                `json:"verified"`
	Confidence            float64             `json:"confidence"`
	AcceptedEvidencePaths []string            `json:"accepted_evidence_paths,omitempty"`
	RejectedEvidencePaths []string            `json:"rejected_evidence_paths,omitempty"`
	Warnings              []string            `json:"warnings,omitempty"`
	Checks                []VerificationCheck `json:"checks"`
}

type VerificationCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type functionVerificationBuilder struct {
	report FunctionVerification
}

func DisabledVerification(generatedAt time.Time, options Options, note string) Verification {
	normalized := options.Normalize()
	return Verification{
		SchemaVersion: VerificationSchemaVersion,
		GeneratedAt:   generatedAt,
		Mode:          string(normalized.Mode),
		Status:        "disabled",
		Note:          strings.TrimSpace(note),
	}
}

func NewVerification(generatedAt time.Time, mode string, results []FunctionResult) Verification {
	report := Verification{
		SchemaVersion: VerificationSchemaVersion,
		GeneratedAt:   generatedAt,
		Mode:          strings.TrimSpace(mode),
		Functions:     make([]FunctionVerification, 0, len(results)),
	}

	for _, result := range results {
		functionReport := result.Verification
		if strings.TrimSpace(functionReport.Name) == "" {
			functionReport = fallbackFunctionVerification(result)
		}
		report.Functions = append(report.Functions, functionReport)
		if functionReport.Verified {
			report.VerifiedCount++
		}
		report.WarningCount += len(functionReport.Warnings)
		for _, check := range functionReport.Checks {
			if check.Status == "fail" {
				report.FailedCheckCount++
			}
		}
	}

	report.Status = verificationStatus(report)
	if report.Status != "verified" {
		report.Note = "one or more function outputs need evidence or structure review"
	}
	return report
}

func VerifyFunctionOutputDetailed(job FunctionJob, output FunctionOutput) (FunctionOutput, FunctionVerification) {
	builder := newFunctionVerificationBuilder(job.Name)
	allowed := allowedEvidence(job.EvidencePaths)
	if len(allowed) == 0 {
		builder.fail("evidence.available", "function job does not include evidence paths")
		builder.finish()
		return output, builder.report
	}

	output, evidenceOK := verifyEvidencePaths(output, allowed, builder)
	verifyFunctionShape(job.Name, output, builder)
	if evidenceOK {
		builder.pass("evidence.paths", "all referenced evidence paths are allowed for this function")
	}

	builder.finish()
	return output, builder.report
}

func VerifyFunctionOutput(job FunctionJob, output FunctionOutput) (FunctionOutput, bool) {
	output, report := VerifyFunctionOutputDetailed(job, output)
	return output, report.Verified
}

func newFunctionVerificationBuilder(name string) *functionVerificationBuilder {
	return &functionVerificationBuilder{
		report: FunctionVerification{
			Name: strings.TrimSpace(name),
		},
	}
}

func (builder *functionVerificationBuilder) pass(name string, detail string) {
	builder.report.Checks = append(builder.report.Checks, VerificationCheck{Name: name, Status: "pass", Detail: strings.TrimSpace(detail)})
}

func (builder *functionVerificationBuilder) warn(name string, detail string) {
	builder.report.Warnings = append(builder.report.Warnings, strings.TrimSpace(detail))
	builder.report.Checks = append(builder.report.Checks, VerificationCheck{Name: name, Status: "warn", Detail: strings.TrimSpace(detail)})
}

func (builder *functionVerificationBuilder) fail(name string, detail string) {
	builder.report.Checks = append(builder.report.Checks, VerificationCheck{Name: name, Status: "fail", Detail: strings.TrimSpace(detail)})
}

func (builder *functionVerificationBuilder) finish() {
	failed := 0
	for _, check := range builder.report.Checks {
		if check.Status == "fail" {
			failed++
		}
	}

	builder.report.Confidence = verificationConfidence(len(builder.report.Checks), failed, len(builder.report.Warnings))
	builder.report.Verified = failed == 0
	builder.report.Status = "verified"
	if !builder.report.Verified {
		builder.report.Status = "unverified"
	}
}

func allowedEvidence(paths []string) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	return allowed
}

func verifyEvidencePaths(output FunctionOutput, allowed map[string]struct{}, builder *functionVerificationBuilder) (FunctionOutput, bool) {
	verified := true
	for index := range output.KeyFindings {
		output.KeyFindings[index].EvidencePaths, verified = filterEvidencePathsDetailed(output.KeyFindings[index].EvidencePaths, allowed, verified, builder)
	}
	for index := range output.GraphEdges {
		output.GraphEdges[index].EvidencePaths, verified = filterEvidencePathsDetailed(output.GraphEdges[index].EvidencePaths, allowed, verified, builder)
	}
	for index := range output.IssueSignals {
		output.IssueSignals[index].EvidencePaths, verified = filterEvidencePathsDetailed(output.IssueSignals[index].EvidencePaths, allowed, verified, builder)
	}
	if !verified {
		builder.fail("evidence.paths", "output referenced evidence paths outside this function job")
	}
	return output, verified
}

func filterEvidencePathsDetailed(paths []string, allowed map[string]struct{}, verified bool, builder *functionVerificationBuilder) ([]string, bool) {
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		if _, ok := allowed[trimmed]; !ok {
			verified = false
			builder.report.RejectedEvidencePaths = appendUnique(builder.report.RejectedEvidencePaths, trimmed)
			continue
		}
		builder.report.AcceptedEvidencePaths = appendUnique(builder.report.AcceptedEvidencePaths, trimmed)
		filtered = append(filtered, trimmed)
	}
	return filtered, verified
}

func verifyFunctionShape(name string, output FunctionOutput, builder *functionVerificationBuilder) {
	switch strings.TrimSpace(name) {
	case "summary":
		requireSummary(output, builder)
		requireEvidenceFinding(output, builder)
	case "architecture":
		requireSummary(output, builder)
		requireEvidenceFinding(output, builder)
	case "hotspots-and-dependencies":
		requireSummary(output, builder)
		requireEvidenceFinding(output, builder)
	case "flowchart":
		requireGraphEdges(output, builder)
	case "issues":
		requireIssueSignals(output, builder)
	case "recommendations":
		requireRecommendations(output, builder)
		requireGrounding(output, builder)
	case "dashboard":
		requireDashboardSurface(output, builder)
	default:
		requireAnyUsefulOutput(output, builder)
	}
}

func requireSummary(output FunctionOutput, builder *functionVerificationBuilder) {
	if strings.TrimSpace(output.Summary) == "" {
		builder.fail("summary.present", "summary text is required")
		return
	}
	builder.pass("summary.present", "summary text is present")
}

func requireEvidenceFinding(output FunctionOutput, builder *functionVerificationBuilder) {
	for _, finding := range output.KeyFindings {
		if strings.TrimSpace(finding.Claim) != "" && len(finding.EvidencePaths) > 0 {
			builder.pass("findings.evidence", "at least one finding has accepted evidence")
			return
		}
	}
	builder.fail("findings.evidence", "at least one evidence-backed finding is required")
}

func requireGraphEdges(output FunctionOutput, builder *functionVerificationBuilder) {
	if len(output.GraphEdges) == 0 {
		builder.fail("graph.edges", "flowchart output requires at least one graph edge")
		return
	}
	for _, edge := range output.GraphEdges {
		if strings.TrimSpace(edge.From) == "" || strings.TrimSpace(edge.To) == "" {
			builder.fail("graph.edge_shape", "graph edge source and target must be non-empty")
			return
		}
		if len(edge.EvidencePaths) == 0 {
			builder.fail("graph.edge_evidence", "graph edge must include accepted evidence")
			return
		}
	}
	builder.pass("graph.edges", "graph edges include source, target, and accepted evidence")
}

func requireIssueSignals(output FunctionOutput, builder *functionVerificationBuilder) {
	if len(output.IssueSignals) == 0 {
		builder.warn("issues.present", "no issue signals were produced")
		return
	}
	for _, signal := range output.IssueSignals {
		if strings.TrimSpace(signal.Title) == "" {
			builder.fail("issues.title", "issue signal title must be non-empty")
			return
		}
		if !validSeverity(signal.Severity) {
			builder.fail("issues.severity", "issue severity must be low, medium, high, or critical")
			return
		}
		if len(signal.EvidencePaths) == 0 {
			builder.fail("issues.evidence", "issue signal must include accepted evidence")
			return
		}
	}
	builder.pass("issues.signals", "issue signals include severity and accepted evidence")
}

func requireRecommendations(output FunctionOutput, builder *functionVerificationBuilder) {
	if len(output.Recommendations) == 0 {
		builder.fail("recommendations.present", "recommendations output requires at least one recommendation")
		return
	}
	builder.pass("recommendations.present", "recommendations are present")
}

func requireGrounding(output FunctionOutput, builder *functionVerificationBuilder) {
	if len(output.KeyFindings) > 0 || len(output.Uncertainties) > 0 {
		builder.pass("recommendations.grounded", "recommendations include findings or uncertainties for grounding")
		return
	}
	builder.warn("recommendations.grounded", "recommendations have no explicit findings or uncertainty notes")
}

func requireDashboardSurface(output FunctionOutput, builder *functionVerificationBuilder) {
	if strings.TrimSpace(output.Summary) != "" && (len(output.KeyFindings)+len(output.Recommendations)+len(output.GraphEdges)+len(output.IssueSignals)) > 0 {
		builder.pass("dashboard.surface", "dashboard output has a summary and at least one structured section")
		return
	}
	builder.fail("dashboard.surface", "dashboard output requires a summary plus one structured section")
}

func requireAnyUsefulOutput(output FunctionOutput, builder *functionVerificationBuilder) {
	if strings.TrimSpace(output.Summary) != "" || len(output.KeyFindings) > 0 || len(output.Recommendations) > 0 || len(output.GraphEdges) > 0 || len(output.IssueSignals) > 0 {
		builder.pass("output.present", "function produced usable output")
		return
	}
	builder.fail("output.present", "function produced no usable output")
}

func validSeverity(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

func verificationConfidence(totalChecks int, failedChecks int, warnings int) float64 {
	if totalChecks == 0 {
		return 0
	}
	score := 1 - (float64(failedChecks)*0.5+float64(warnings)*0.15)/float64(totalChecks)
	return math.Round(max(0, min(1, score))*100) / 100
}

func verificationStatus(report Verification) string {
	if len(report.Functions) == 0 {
		return "disabled"
	}
	if report.FailedCheckCount > 0 {
		return "needs-review"
	}
	if report.WarningCount > 0 {
		return "verified-with-warnings"
	}
	return "verified"
}

func fallbackFunctionVerification(result FunctionResult) FunctionVerification {
	builder := newFunctionVerificationBuilder(result.Name)
	if strings.TrimSpace(result.Error) != "" {
		builder.fail("function.error", result.Error)
	} else if result.Verified {
		builder.pass("function.result", "function result was already marked verified")
	} else {
		builder.warn("function.result", "function result has no detailed verification report")
	}
	builder.finish()
	return builder.report
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
