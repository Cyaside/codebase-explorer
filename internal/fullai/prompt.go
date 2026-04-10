package fullai

import (
	"encoding/json"
	"fmt"
	"strings"
)

const functionSystemPrompt = `You are a Codebase Explorer function worker.
Use only the supplied repository evidence.
Return exactly one valid JSON object and nothing else.
Do not wrap the JSON in markdown fences.
Do not invent files, modules, flows, incidents, or paths.
If evidence is thin, say so in uncertainties.`

func BuildFunctionPrompt(job FunctionJob, evidence Evidence) (FunctionPrompt, error) {
	payload := map[string]any{
		"function":       job.Name,
		"objective":      job.Objective,
		"focus":          job.Focus,
		"evidence_items": evidenceForJob(job, evidence.Items),
		"output_shape": map[string]any{
			"summary":         "string",
			"key_findings":    []map[string]any{{"claim": "string", "evidence_paths": []string{"path"}, "confidence": "low|medium|high"}},
			"recommendations": []string{"string"},
			"graph_edges":     []map[string]any{{"from": "path-or-node", "to": "path-or-node", "label": "string", "evidence_paths": []string{"path"}}},
			"issue_signals":   []map[string]any{{"title": "string", "severity": "low|medium|high", "evidence_paths": []string{"path"}}},
			"uncertainties":   []string{"string"},
		},
	}
	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return FunctionPrompt{}, fmt.Errorf("marshal function prompt payload: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("Run this Codebase Explorer function using only the supplied evidence.\n")
	builder.WriteString("Keep the response concise, concrete, and implementation-focused.\n")
	builder.WriteString("Use paths exactly as provided in evidence_paths.\n\n")
	builder.Write(payloadJSON)

	return FunctionPrompt{
		FunctionName: job.Name,
		SystemPrompt: functionSystemPrompt,
		UserPrompt:   builder.String(),
	}, nil
}

func evidenceForJob(job FunctionJob, items []EvidenceItem) []EvidenceItem {
	if len(job.EvidencePaths) == 0 {
		return nil
	}
	allowed := map[string]struct{}{}
	for _, path := range job.EvidencePaths {
		allowed[strings.TrimSpace(path)] = struct{}{}
	}

	selected := make([]EvidenceItem, 0, len(job.EvidencePaths))
	for _, item := range items {
		displayPath := strings.TrimSpace(item.DisplayPath)
		if displayPath == "" {
			displayPath = strings.TrimSpace(item.ResolvedPath)
		}
		if _, ok := allowed[displayPath]; !ok {
			continue
		}
		selected = append(selected, item)
	}
	return selected
}

func ParseFunctionOutput(content string) (FunctionOutput, error) {
	payload := strings.TrimSpace(stripMarkdownFence(content))
	if payload == "" {
		return FunctionOutput{}, fmt.Errorf("function response did not contain JSON content")
	}

	var output FunctionOutput
	if err := json.Unmarshal([]byte(payload), &output); err != nil {
		return FunctionOutput{}, fmt.Errorf("parse function response: %w", err)
	}

	output = normalizeFunctionOutput(output)
	if strings.TrimSpace(output.Summary) == "" &&
		len(output.KeyFindings) == 0 &&
		len(output.Recommendations) == 0 &&
		len(output.GraphEdges) == 0 &&
		len(output.IssueSignals) == 0 &&
		len(output.Uncertainties) == 0 {
		return FunctionOutput{}, fmt.Errorf("function response did not include any usable sections")
	}

	return output, nil
}

func normalizeFunctionOutput(output FunctionOutput) FunctionOutput {
	output.Summary = strings.TrimSpace(output.Summary)
	output.Recommendations = trimStrings(output.Recommendations)
	output.Uncertainties = trimStrings(output.Uncertainties)
	output.KeyFindings = normalizeFindings(output.KeyFindings)
	output.GraphEdges = normalizeGraphEdges(output.GraphEdges)
	output.IssueSignals = normalizeIssueSignals(output.IssueSignals)
	return output
}

func normalizeFindings(items []Finding) []Finding {
	normalized := make([]Finding, 0, len(items))
	for _, item := range items {
		claim := strings.TrimSpace(item.Claim)
		if claim == "" {
			continue
		}
		normalized = append(normalized, Finding{
			Claim:         claim,
			EvidencePaths: trimStrings(item.EvidencePaths),
			Confidence:    strings.TrimSpace(item.Confidence),
		})
	}
	return normalized
}

func normalizeGraphEdges(items []GraphEdge) []GraphEdge {
	normalized := make([]GraphEdge, 0, len(items))
	for _, item := range items {
		from := strings.TrimSpace(item.From)
		to := strings.TrimSpace(item.To)
		if from == "" || to == "" {
			continue
		}
		normalized = append(normalized, GraphEdge{
			From:          from,
			To:            to,
			Label:         strings.TrimSpace(item.Label),
			EvidencePaths: trimStrings(item.EvidencePaths),
		})
	}
	return normalized
}

func normalizeIssueSignals(items []IssueSignal) []IssueSignal {
	normalized := make([]IssueSignal, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		normalized = append(normalized, IssueSignal{
			Title:         title,
			Severity:      strings.TrimSpace(item.Severity),
			EvidencePaths: trimStrings(item.EvidencePaths),
		})
	}
	return normalized
}

func trimStrings(items []string) []string {
	trimmed := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	return trimmed
}

func stripMarkdownFence(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 3 {
		return trimmed
	}
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		return trimmed
	}

	lastIndex := len(lines) - 1
	if strings.TrimSpace(lines[lastIndex]) != "```" {
		return trimmed
	}

	return strings.Join(lines[1:lastIndex], "\n")
}
