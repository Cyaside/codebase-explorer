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
			"allowed_evidence_paths": job.EvidencePaths,
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
	var envelope struct {
		Functions map[string]json.RawMessage `json:"functions"`
	}
	if err := json.Unmarshal([]byte(stripMarkdownFence(content)), &envelope); err != nil {
		return nil, fmt.Errorf("parse batch response: %w", err)
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
