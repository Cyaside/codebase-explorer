package fullai

import (
	"regexp"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

var secretAssignment = regexp.MustCompile(`(?im)\b(api[_-]?key|access[_-]?token|auth[_-]?token|client[_-]?secret|password|passwd|private[_-]?key|secret[_-]?key)\b(\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,}]+)`)
var privateKeyBlock = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)
var bearerToken = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{12,}`)
var tokenLiteral = regexp.MustCompile(`\b(?:sk-[A-Za-z0-9_-]{16,}|gh[pousr]_[A-Za-z0-9_]{20,}|AKIA[0-9A-Z]{16})\b`)

func SensitiveEvidencePath(path string) bool {
	return repo.SensitivePath(path)
}

func RedactSensitiveText(content string) string {
	content = privateKeyBlock.ReplaceAllString(content, "[REDACTED PRIVATE KEY]")
	content = secretAssignment.ReplaceAllString(content, "${1}${2}[REDACTED]")
	content = bearerToken.ReplaceAllString(content, "Bearer [REDACTED]")
	return tokenLiteral.ReplaceAllString(content, "[REDACTED]")
}

func RedactFunctionOutput(output FunctionOutput) FunctionOutput {
	output.Summary = RedactSensitiveText(output.Summary)
	for i := range output.KeyFindings {
		output.KeyFindings[i].Claim = RedactSensitiveText(output.KeyFindings[i].Claim)
	}
	for i := range output.Recommendations {
		output.Recommendations[i] = RedactSensitiveText(output.Recommendations[i])
	}
	for i := range output.GraphEdges {
		output.GraphEdges[i].From = RedactSensitiveText(output.GraphEdges[i].From)
		output.GraphEdges[i].To = RedactSensitiveText(output.GraphEdges[i].To)
		output.GraphEdges[i].Label = RedactSensitiveText(output.GraphEdges[i].Label)
	}
	for i := range output.IssueSignals {
		output.IssueSignals[i].Title = RedactSensitiveText(output.IssueSignals[i].Title)
	}
	for i := range output.Uncertainties {
		output.Uncertainties[i] = RedactSensitiveText(output.Uncertainties[i])
	}
	return output
}
