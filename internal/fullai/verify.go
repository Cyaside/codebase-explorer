package fullai

import "strings"

func VerifyFunctionOutput(job FunctionJob, output FunctionOutput) (FunctionOutput, bool) {
	allowed := map[string]struct{}{}
	for _, path := range job.EvidencePaths {
		trimmed := strings.TrimSpace(path)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return output, false
	}

	verified := true
	for index := range output.KeyFindings {
		output.KeyFindings[index].EvidencePaths, verified = filterEvidencePaths(output.KeyFindings[index].EvidencePaths, allowed, verified)
	}
	for index := range output.GraphEdges {
		output.GraphEdges[index].EvidencePaths, verified = filterEvidencePaths(output.GraphEdges[index].EvidencePaths, allowed, verified)
	}
	for index := range output.IssueSignals {
		output.IssueSignals[index].EvidencePaths, verified = filterEvidencePaths(output.IssueSignals[index].EvidencePaths, allowed, verified)
	}
	return output, verified
}

func filterEvidencePaths(paths []string, allowed map[string]struct{}, verified bool) ([]string, bool) {
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		if _, ok := allowed[trimmed]; !ok {
			verified = false
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered, verified
}
