package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

const synthesisSystemPrompt = `You are a repository orientation assistant.
Use only the supplied repository facts.
Return exactly one valid JSON object and nothing else.
Do not wrap the JSON in markdown fences.
Keep the prose concise, concrete, and implementation-focused.`

func buildSynthesisPrompt(context CondensedContext) (string, string, error) {
	contextJSON, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal condensed context: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("Summarize this repository context into a JSON object with this exact shape:\n")
	builder.WriteString("{\n")
	builder.WriteString(`  "project_summary": "string",` + "\n")
	builder.WriteString(`  "architecture_narrative": "string",` + "\n")
	builder.WriteString(`  "hotspot_explanations": [{"path": "string", "explanation": "string"}],` + "\n")
	builder.WriteString(`  "reading_path_explanations": [{"path": "string", "rationale": "string"}]` + "\n")
	builder.WriteString("}\n\n")
	builder.WriteString("Rules:\n")
	builder.WriteString("- Use only paths present in the provided context.\n")
	builder.WriteString("- Keep hotspot_explanations aligned with the hotspot list.\n")
	builder.WriteString("- Keep reading_path_explanations aligned with the reading_path list.\n")
	builder.WriteString("- If a section is uncertain, use an empty string or an empty array instead of inventing facts.\n\n")
	builder.WriteString("Repository context:\n")
	builder.Write(contextJSON)

	return synthesisSystemPrompt, builder.String(), nil
}

func parseSynthesisResult(content string) (Result, error) {
	payload := strings.TrimSpace(stripMarkdownFence(content))
	if payload == "" {
		return Result{}, fmt.Errorf("synthesis response did not contain JSON content")
	}

	var result Result
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return Result{}, fmt.Errorf("parse synthesis response: %w", err)
	}

	if strings.TrimSpace(result.ProjectSummary) == "" &&
		strings.TrimSpace(result.ArchitectureNarrative) == "" &&
		len(result.HotspotExplanations) == 0 &&
		len(result.ReadingPathExplanations) == 0 {
		return Result{}, fmt.Errorf("synthesis response did not include any usable sections")
	}

	return result, nil
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
