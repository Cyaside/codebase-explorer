package fullai

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	agentpack "github.com/Cyaside/codebase-explorer"
)

// BuildBatchPrompt puts the shared worker contract in the system message once.
// The evidence budget applies to the entire call, not to each function.
func BuildBatchPrompt(jobs []FunctionJob, evidence Evidence, maxEvidenceBytes int) (FunctionPrompt, error) {
	return buildBatchPromptFrom(agentpack.Files, jobs, evidence, maxEvidenceBytes)
}

func buildBatchPromptFrom(files fs.FS, jobs []FunctionJob, evidence Evidence, maxEvidenceBytes int) (FunctionPrompt, error) {
	if len(jobs) == 0 {
		return FunctionPrompt{}, fmt.Errorf("batch has no functions")
	}
	if maxEvidenceBytes <= 0 {
		maxEvidenceBytes = 40000
	}
	var system strings.Builder
	system.WriteString(functionSystemPrompt)
	system.WriteString("\nReturn a JSON object with a functions property mapping each requested function name to its output object. Include every requested function.\n")
	system.WriteString("Keep each summary under 80 words and each function to at most 3 findings, 3 recommendations, 6 graph edges, and 3 issue signals. Use exact evidence file paths for path-like graph endpoints; use short names without slashes for conceptual steps.\n")
	for _, filePath := range packPrelude {
		data, err := fs.ReadFile(files, filePath)
		if err != nil {
			return FunctionPrompt{}, fmt.Errorf("read shared instruction %q: %w", filePath, err)
		}
		if strings.TrimSpace(string(data)) == "" {
			return FunctionPrompt{}, fmt.Errorf("shared instruction %q is empty", filePath)
		}
		system.WriteString("\n\n# Instruction: ")
		system.WriteString(filePath)
		system.WriteByte('\n')
		system.Write(data)
	}
	requested := make([]map[string]any, 0, len(jobs))
	selected := make([]EvidenceItem, 0)
	seen := map[string]struct{}{}
	for _, job := range jobs {
		instructions, err := loadInstructionsFrom(files, job)
		if err != nil {
			return FunctionPrompt{}, err
		}
		// The function guide is the final document in the validated pack.
		guidePath := instructions.Paths[len(instructions.Paths)-1]
		guide, err := fs.ReadFile(files, guidePath)
		if err != nil {
			return FunctionPrompt{}, err
		}
		system.WriteString("\n\n# Instruction: ")
		system.WriteString(guidePath)
		system.WriteByte('\n')
		system.Write(guide)
		requested = append(requested, map[string]any{
			"name": job.Name, "objective": job.Objective, "focus": job.Focus,
			"priority_evidence_paths": job.EvidencePaths,
		})
		for _, item := range evidenceForJob(job, evidence.Items) {
			if _, ok := seen[item.DisplayPath]; ok {
				continue
			}
			seen[item.DisplayPath] = struct{}{}
			selected = append(selected, item)
		}
	}
	if len(selected) > 0 {
		perItem := maxEvidenceBytes / len(selected)
		for i := range selected {
			if len(selected[i].Snippet) > perItem {
				selected[i].Snippet = validUTF8Prefix(selected[i].Snippet, perItem)
				selected[i].Truncated = true
			}
		}
	}
	allowedPaths := make([]string, 0, len(selected))
	for _, item := range selected {
		if item.ReadStatus == "read" && item.Snippet != "" {
			allowedPaths = append(allowedPaths, item.DisplayPath)
		}
	}
	for _, function := range requested {
		function["allowed_evidence_paths"] = allowedPaths
	}
	payload, err := json.Marshal(map[string]any{
		"functions": requested, "evidence_items": selected,
		"output_shape": map[string]any{"functions": map[string]any{
			"<function-name>": map[string]any{
				"summary": "string", "key_findings": []any{map[string]any{"claim": "string", "evidence_paths": []string{"path"}, "confidence": "low|medium|high"}},
				"recommendations": []string{"string"}, "graph_edges": []any{map[string]any{"from": "node", "to": "node", "label": "string", "evidence_paths": []string{"path"}}},
				"issue_signals": []any{map[string]any{"title": "string", "severity": "low|medium|high", "evidence_paths": []string{"path"}}},
				"uncertainties": []string{"string"},
			},
		}},
	})
	if err != nil {
		return FunctionPrompt{}, err
	}
	return FunctionPrompt{FunctionName: "batch", SystemPrompt: system.String(), UserPrompt: string(payload)}, nil
}

func validUTF8Prefix(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len(value) <= limit {
		return value
	}
	for limit > 0 && limit < len(value) && value[limit]&0xc0 == 0x80 {
		limit--
	}
	return value[:limit]
}

func ParseBatchOutput(content string) (map[string]FunctionOutput, error) {
	content = stripMarkdownFence(content)
	var envelope struct {
		Functions map[string]json.RawMessage `json:"functions"`
	}
	if err := json.Unmarshal([]byte(content), &envelope); err != nil {
		content = removeTrailingJSONCommas(content)
		if cleanErr := json.Unmarshal([]byte(content), &envelope); cleanErr != nil {
			// A response may end or break after complete function objects. Keep
			// those objects so only the missing sections need a repair call.
			if completed := parseCompletedBatchFunctions(content); len(completed) > 0 {
				return completed, nil
			}
			return nil, fmt.Errorf("parse batch response: %w", cleanErr)
		}
	}
	if len(envelope.Functions) == 0 {
		return nil, fmt.Errorf("batch response has no functions")
	}
	outputs := make(map[string]FunctionOutput, len(envelope.Functions))
	for name, raw := range envelope.Functions {
		output, err := ParseFunctionOutput(string(raw))
		if err != nil {
			continue
		}
		outputs[name] = output
	}
	return outputs, nil
}

func removeTrailingJSONCommas(content string) string {
	var clean strings.Builder
	clean.Grow(len(content))
	inString, escaped := false, false
	for index := 0; index < len(content); index++ {
		character := content[index]
		if inString {
			clean.WriteByte(character)
			if escaped {
				escaped = false
			} else if character == '\\' {
				escaped = true
			} else if character == '"' {
				inString = false
			}
			continue
		}
		if character == '"' {
			inString = true
		}
		if character == ',' {
			lookahead := index + 1
			for lookahead < len(content) && (content[lookahead] == ' ' || content[lookahead] == '\n' || content[lookahead] == '\r' || content[lookahead] == '\t') {
				lookahead++
			}
			if lookahead < len(content) && (content[lookahead] == '}' || content[lookahead] == ']') {
				continue
			}
		}
		clean.WriteByte(character)
	}
	return clean.String()
}

func parseCompletedBatchFunctions(content string) map[string]FunctionOutput {
	decoder := json.NewDecoder(strings.NewReader(content))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil
	}
	outputs := map[string]FunctionOutput{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return outputs
		}
		if key != "functions" {
			var ignored json.RawMessage
			if decoder.Decode(&ignored) != nil {
				return outputs
			}
			continue
		}
		start, err := decoder.Token()
		if err != nil || start != json.Delim('{') {
			return outputs
		}
		for decoder.More() {
			nameToken, err := decoder.Token()
			name, ok := nameToken.(string)
			if err != nil || !ok {
				return outputs
			}
			var raw json.RawMessage
			if decoder.Decode(&raw) != nil {
				return outputs
			}
			output, err := ParseFunctionOutput(string(raw))
			if err == nil {
				outputs[name] = output
			}
		}
		return outputs
	}
	return outputs
}
