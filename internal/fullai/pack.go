package fullai

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"strings"

	agentpack "github.com/Cyaside/codebase-explorer"
)

const PackVersion = "agents.v1"

var packPrelude = []string{
	".agents/README.md",
	".agents/agents.md",
	".agents/ai/README.md",
}

type Instructions struct {
	Text  string
	Hash  string
	Paths []string
}

func loadInstructions(job FunctionJob) (Instructions, error) {
	return loadInstructionsFrom(agentpack.Files, job)
}

func InstructionHash(job FunctionJob) (string, error) {
	instructions, err := loadInstructions(job)
	if err != nil {
		return "", err
	}
	return instructions.Hash, nil
}

// PackHash validates every required worker guide and identifies the exact pack
// used by a run. The hash changes when any embedded guide changes.
func PackHash() (string, error) {
	return packHashFrom(agentpack.Files)
}

func packHashFrom(files fs.FS) (string, error) {
	hash := sha256.New()
	for _, task := range defaultFunctionTasks() {
		instructions, err := loadInstructionsFrom(files, FunctionJob{Name: task.Name})
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(task.Name + "\x00" + instructions.Hash + "\x00"))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func loadInstructionsFrom(files fs.FS, job FunctionJob) (Instructions, error) {
	name := strings.TrimSpace(job.Name)
	if name == "" || strings.ContainsAny(name, `/\.`) {
		return Instructions{}, fmt.Errorf("invalid function name %q", job.Name)
	}
	functionPath := path.Join(".agents/ai/functions", name+".md")
	if job.InstructionPath != "" && strings.TrimSpace(job.InstructionPath) != functionPath {
		return Instructions{}, fmt.Errorf("function %q instruction path %q does not match %q", name, job.InstructionPath, functionPath)
	}
	paths := append(append([]string(nil), packPrelude...), functionPath)
	hash := sha256.New()
	var builder strings.Builder
	for _, filePath := range paths {
		data, err := fs.ReadFile(files, filePath)
		if err != nil {
			return Instructions{}, fmt.Errorf("read instruction %q: %w", filePath, err)
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			return Instructions{}, fmt.Errorf("instruction %q is empty", filePath)
		}
		_, _ = hash.Write([]byte(filePath + "\x00" + content + "\x00"))
		builder.WriteString("\n\n# Instruction: ")
		builder.WriteString(filePath)
		builder.WriteString("\n")
		builder.WriteString(content)
	}
	return Instructions{Text: builder.String(), Hash: hex.EncodeToString(hash.Sum(nil)), Paths: paths}, nil
}
