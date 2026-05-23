package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
)

const SchemaVersion = "graphs.v1"

type Set struct {
	SchemaVersion string `json:"schema_version"`
	Views         []View `json:"views"`
}

type View struct {
	ID    string `json:"id"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	Note  string `json:"note,omitempty"`
}

type Node struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	Type          string   `json:"type"`
	Path          string   `json:"path,omitempty"`
	EvidencePaths []string `json:"evidence_paths,omitempty"`
}

type Edge struct {
	ID            string   `json:"id"`
	Source        string   `json:"source"`
	Target        string   `json:"target"`
	Relation      string   `json:"relation"`
	EvidencePaths []string `json:"evidence_paths"`
	SourceKind    string   `json:"source_kind"`
}

func Build(root string, analysis analyzer.Result, execution fullai.Execution) Set {
	return Set{SchemaVersion: SchemaVersion, Views: []View{
		architecture(analysis, execution),
		flow(execution),
		dependencies(root, analysis),
	}}
}

func architecture(analysis analyzer.Result, execution fullai.Execution) View {
	view := View{ID: "architecture", Nodes: []Node{}, Edges: []Edge{}}
	known := map[string]Node{}
	for _, module := range analysis.Modules {
		modulePath := normalize(module.Path)
		if modulePath == "" {
			continue
		}
		node := Node{ID: nodeID("module", modulePath), Label: modulePath, Type: "module", Path: modulePath}
		for _, file := range analysis.Files {
			filePath := normalize(file.Path)
			if modulePath == "." || strings.HasPrefix(filePath, modulePath+"/") {
				node.EvidencePaths = []string{filePath}
				break
			}
		}
		view.Nodes = append(view.Nodes, node)
		known[modulePath] = node
	}
	for _, entry := range analysis.EntryPoints {
		entryPath := normalize(entry)
		if entryPath == "" {
			continue
		}
		node := Node{ID: nodeID("entry", entryPath), Label: path.Base(entryPath), Type: "entry-point", Path: entryPath, EvidencePaths: []string{entryPath}}
		view.Nodes = append(view.Nodes, node)
		bestPath := ""
		for modulePath := range known {
			if (modulePath == "." || entryPath == modulePath || strings.HasPrefix(entryPath, modulePath+"/")) && len(modulePath) > len(bestPath) {
				bestPath = modulePath
			}
		}
		if bestPath != "" {
			view.Edges = append(view.Edges, newEdge(node.ID, known[bestPath].ID, "belongs-to", []string{entryPath}, "scan"))
		}
		known[entryPath] = node
	}
	for _, result := range execution.Results {
		if result.Name != "architecture" || !result.Verified {
			continue
		}
		for _, edge := range result.Output.GraphEdges {
			if len(edge.EvidencePaths) == 0 {
				continue
			}
			from := architectureNode(edge.From, edge.EvidencePaths, known, &view)
			to := architectureNode(edge.To, edge.EvidencePaths, known, &view)
			if from.ID != to.ID {
				view.Edges = append(view.Edges, newEdge(from.ID, to.ID, edge.Label, edge.EvidencePaths, "ai-evidence"))
			}
		}
	}
	view.Edges = uniqueEdges(view.Edges)
	return view
}

func architectureNode(label string, evidence []string, known map[string]Node, view *View) Node {
	name := normalize(label)
	if node, ok := known[name]; ok {
		for _, evidencePath := range evidence {
			if !slices.Contains(node.EvidencePaths, evidencePath) {
				node.EvidencePaths = append(node.EvidencePaths, evidencePath)
			}
		}
		known[name] = node
		for index := range view.Nodes {
			if view.Nodes[index].ID == node.ID {
				view.Nodes[index] = node
				break
			}
		}
		return node
	}
	typeName := "component"
	pathName := ""
	if strings.Contains(name, "/") || path.Ext(name) != "" {
		typeName = "file"
		pathName = name
		if path.Ext(name) == "" {
			typeName = "module"
		}
	}
	node := Node{ID: nodeID("architecture", name), Label: name, Type: typeName, Path: pathName, EvidencePaths: append([]string(nil), evidence...)}
	known[name] = node
	view.Nodes = append(view.Nodes, node)
	return node
}

func flow(execution fullai.Execution) View {
	view := View{ID: "flow", Nodes: []Node{}, Edges: []Edge{}}
	seen := map[string]int{}
	for _, result := range execution.Results {
		if result.Name != "flowchart" || !result.Verified {
			continue
		}
		for _, edge := range result.Output.GraphEdges {
			if len(edge.EvidencePaths) == 0 {
				continue
			}
			for _, label := range []string{edge.From, edge.To} {
				id := nodeID("flow", strings.TrimSpace(label))
				if index, exists := seen[id]; !exists {
					view.Nodes = append(view.Nodes, Node{ID: id, Label: strings.TrimSpace(label), Type: "flow-step", EvidencePaths: append([]string(nil), edge.EvidencePaths...)})
					seen[id] = len(view.Nodes) - 1
				} else {
					for _, evidence := range edge.EvidencePaths {
						if !slices.Contains(view.Nodes[index].EvidencePaths, evidence) {
							view.Nodes[index].EvidencePaths = append(view.Nodes[index].EvidencePaths, evidence)
						}
					}
				}
			}
			view.Edges = append(view.Edges, newEdge(nodeID("flow", edge.From), nodeID("flow", edge.To), edge.Label, edge.EvidencePaths, "ai-evidence"))
		}
	}
	view.Edges = uniqueEdges(view.Edges)
	if len(view.Edges) == 0 {
		view.Note = "No evidence-backed runtime flow was returned."
	}
	return view
}

var relativeImport = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*|\brequire\s*\()\s*['"]((?:\.{1,2}/|@/)[^'"]+)['"]`)

func dependencies(root string, analysis analyzer.Result) View {
	view := View{ID: "dependencies", Nodes: []Node{}, Edges: []Edge{}}
	knownFiles := map[string]struct{}{}
	for _, file := range analysis.Files {
		knownFiles[normalize(file.Path)] = struct{}{}
	}
	moduleName := readGoModule(root)
	nodes := map[string]Node{}
	type moduleImport struct {
		from     string
		to       string
		evidence map[string]struct{}
	}
	imports := map[string]*moduleImport{}
	for _, file := range analysis.Files {
		filePath := normalize(file.Path)
		absolute := filepath.Join(root, filepath.FromSlash(filePath))
		var targets []string
		switch strings.ToLower(path.Ext(filePath)) {
		case ".go":
			targets = goImportTargets(absolute, moduleName, knownFiles)
		case ".ts", ".tsx", ".js", ".jsx":
			targets = jsImportTargets(absolute, filePath, knownFiles)
		}
		for _, target := range targets {
			from := normalize(path.Dir(filePath))
			to := normalize(path.Dir(target))
			if from == to {
				continue
			}
			fromID, toID := nodeID("dependency", from), nodeID("dependency", to)
			nodes[fromID] = Node{ID: fromID, Label: from, Type: "module", Path: from}
			nodes[toID] = Node{ID: toID, Label: to, Type: "module", Path: to}
			key := fromID + "\x00" + toID
			importGroup := imports[key]
			if importGroup == nil {
				importGroup = &moduleImport{from: fromID, to: toID, evidence: map[string]struct{}{}}
				imports[key] = importGroup
			}
			importGroup.evidence[filePath] = struct{}{}
			importGroup.evidence[target] = struct{}{}
		}
	}
	for _, node := range nodes {
		view.Nodes = append(view.Nodes, node)
	}
	slices.SortFunc(view.Nodes, func(a, b Node) int { return strings.Compare(a.ID, b.ID) })
	for _, importGroup := range imports {
		evidence := make([]string, 0, len(importGroup.evidence))
		for filePath := range importGroup.evidence {
			evidence = append(evidence, filePath)
		}
		slices.Sort(evidence)
		view.Edges = append(view.Edges, newEdge(importGroup.from, importGroup.to, "imports", evidence, "parsed-import"))
	}
	view.Edges = uniqueEdges(view.Edges)
	nodeEvidence := map[string]map[string]struct{}{}
	for _, edge := range view.Edges {
		for _, id := range []string{edge.Source, edge.Target} {
			if nodeEvidence[id] == nil {
				nodeEvidence[id] = map[string]struct{}{}
			}
			for _, filePath := range edge.EvidencePaths {
				nodeEvidence[id][filePath] = struct{}{}
			}
		}
	}
	for index := range view.Nodes {
		for filePath := range nodeEvidence[view.Nodes[index].ID] {
			view.Nodes[index].EvidencePaths = append(view.Nodes[index].EvidencePaths, filePath)
		}
		slices.Sort(view.Nodes[index].EvidencePaths)
	}
	if len(view.Edges) == 0 {
		view.Note = "No resolved local imports were found for supported Go and JavaScript/TypeScript files."
	}
	return view
}

func readGoModule(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}

func goImportTargets(absolute, moduleName string, known map[string]struct{}) []string {
	if moduleName == "" {
		return nil
	}
	file, err := parser.ParseFile(token.NewFileSet(), absolute, nil, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	var targets []string
	for _, imported := range file.Imports {
		value, err := strconv.Unquote(imported.Path.Value)
		if err != nil || !strings.HasPrefix(value, moduleName+"/") {
			continue
		}
		directory := normalize(strings.TrimPrefix(value, moduleName+"/"))
		matches := make([]string, 0)
		for candidate := range known {
			if path.Dir(candidate) == directory && path.Ext(candidate) == ".go" {
				matches = append(matches, candidate)
			}
		}
		if len(matches) > 0 {
			slices.Sort(matches)
			targets = append(targets, matches[0])
		}
	}
	return targets
}

func jsImportTargets(absolute, source string, known map[string]struct{}) []string {
	data, err := os.ReadFile(absolute)
	if err != nil || len(data) > 1<<20 {
		return nil
	}
	var targets []string
	for _, match := range relativeImport.FindAllStringSubmatch(string(data), -1) {
		base := normalize(path.Join(path.Dir(source), match[1]))
		if strings.HasPrefix(match[1], "@/") {
			base = ""
			for directory := path.Dir(source); ; directory = path.Dir(directory) {
				candidate := path.Join(directory, "src", strings.TrimPrefix(match[1], "@/"))
				if hasJSImportTarget(candidate, known) {
					base = candidate
					break
				}
				if directory == "." || directory == "/" {
					break
				}
			}
			if base == "" {
				continue
			}
		}
		for _, candidate := range []string{base, base + ".ts", base + ".tsx", base + ".js", base + ".jsx", path.Join(base, "index.ts"), path.Join(base, "index.tsx"), path.Join(base, "index.js")} {
			if _, ok := known[candidate]; ok {
				targets = append(targets, candidate)
				break
			}
		}
	}
	return targets
}

func hasJSImportTarget(base string, known map[string]struct{}) bool {
	for _, candidate := range []string{base, base + ".ts", base + ".tsx", base + ".js", base + ".jsx", path.Join(base, "index.ts"), path.Join(base, "index.tsx"), path.Join(base, "index.js")} {
		if _, ok := known[candidate]; ok {
			return true
		}
	}
	return false
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return path.Clean(strings.ReplaceAll(value, "\\", "/"))
}
func nodeID(kind, label string) string { return kind + ":" + strings.ToLower(strings.TrimSpace(label)) }
func newEdge(from, to, relation string, evidence []string, sourceKind string) Edge {
	key := from + "\x00" + to + "\x00" + relation + "\x00" + strings.Join(evidence, "\x00")
	sum := sha256.Sum256([]byte(key))
	return Edge{ID: hex.EncodeToString(sum[:8]), Source: from, Target: to, Relation: strings.TrimSpace(relation), EvidencePaths: append([]string(nil), evidence...), SourceKind: sourceKind}
}
func uniqueEdges(edges []Edge) []Edge {
	seen := map[string]struct{}{}
	unique := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		if _, ok := seen[edge.ID]; ok {
			continue
		}
		seen[edge.ID] = struct{}{}
		unique = append(unique, edge)
	}
	slices.SortFunc(unique, func(a, b Edge) int { return strings.Compare(a.ID, b.ID) })
	return unique
}
