import dagre from "@dagrejs/dagre";
import {
  Background,
  Controls,
  Handle,
  MarkerType,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
} from "@xyflow/react";

import type { InspectorState, WorkbenchBundle } from "@/lib/types";

const graphNodeWidth = 238;
const graphNodeHeight = 112;

interface FlowchartCanvasProps {
  bundle: WorkbenchBundle | null;
  onInspect: (value: InspectorState | null) => void;
}

interface GraphNodeData extends Record<string, unknown> {
  eyebrow: string;
  title: string;
  description: string;
  tone: "project" | "module" | "entry" | "issue" | "reading";
  inspector: InspectorState;
}

type FlowNode = Node<GraphNodeData, "card">;

const nodeTypes = {
  card: GraphCardNode,
};

export function FlowchartCanvas({ bundle, onInspect }: FlowchartCanvasProps) {
  if (!bundle) {
    return (
      <div className="graph-shell">
        <div className="grid h-full place-items-center px-8 text-center">
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-zinc-500">Flowchart</p>
            <h3 className="mt-3 text-lg font-semibold text-white">No project graph yet.</h3>
            <p className="mt-3 max-w-md text-sm leading-6 text-zinc-500">
              Run an analysis for the active workspace and the workbench will render a real project graph here.
            </p>
          </div>
        </div>
      </div>
    );
  }

  const graph = buildFlowGraph(bundle);

  return (
    <div className="graph-shell">
      <ReactFlow<FlowNode, Edge>
        className="codearch-flow"
        defaultEdgeOptions={{
          animated: false,
          markerEnd: {
            color: "#3f3f46",
            height: 18,
            type: MarkerType.ArrowClosed,
            width: 18,
          },
          style: {
            stroke: "#3f3f46",
            strokeWidth: 1.2,
          },
          type: "smoothstep",
        }}
        edges={graph.edges}
        fitView
        fitViewOptions={{ maxZoom: 1.1, padding: 0.22 }}
        minZoom={0.35}
        nodes={graph.nodes}
        nodesDraggable={false}
        nodesFocusable
        nodeTypes={nodeTypes}
        onNodeClick={(_, node) => onInspect(node.data.inspector)}
        panOnDrag
        proOptions={{ hideAttribution: true }}
      >
        <Background color="#171717" gap={20} size={1} />
        <Controls fitViewOptions={{ maxZoom: 1.1, padding: 0.22 }} showInteractive={false} />
      </ReactFlow>
    </div>
  );
}

function GraphCardNode({ data, selected }: NodeProps<FlowNode>) {
  return (
    <div className={`graph-node graph-node-${data.tone}${selected ? " graph-node-selected" : ""}`}>
      <Handle className="graph-node-handle" position={Position.Top} type="target" />
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-zinc-500">{data.eyebrow}</p>
          <h4 className="mt-2 text-sm font-semibold text-zinc-100">{data.title}</h4>
        </div>
        <span className={`graph-node-dot graph-node-dot-${data.tone}`} />
      </div>
      <p className="mt-3 text-xs leading-5 text-zinc-400">{data.description}</p>
      <Handle className="graph-node-handle" position={Position.Bottom} type="source" />
    </div>
  );
}

function buildFlowGraph(bundle: WorkbenchBundle) {
  const nodes: FlowNode[] = [];
  const edges: Edge[] = [];
  const modules = bundle.data.modules.slice(0, 6);
  const rootNodeID = "project-root";

  nodes.push({
    data: {
      eyebrow: "Workspace",
      title: bundle.summary.project_name || bundle.summary.name,
      description: `${bundle.summary.project_type || "Repository"} · ${bundle.summary.total_files} files · ${bundle.summary.total_lines} lines`,
      inspector: {
        eyebrow: "Workspace",
        title: bundle.summary.project_name || bundle.summary.name,
        description: bundle.data.project.summary || bundle.data.ai.project_summary || "No project summary available.",
        notes: [
          bundle.summary.analyzed_path || "Local checkout path unavailable.",
          bundle.data.ai.note || "Deterministic-first bundle with optional AI synthesis.",
        ],
        properties: [
          { label: "Project type", value: bundle.summary.project_type || "Repository" },
          { label: "Files", value: String(bundle.summary.total_files) },
          { label: "Lines", value: String(bundle.summary.total_lines) },
          { label: "AI status", value: bundle.summary.ai_status || "disabled" },
        ],
      },
      tone: "project",
    },
    id: rootNodeID,
    position: { x: 0, y: 0 },
    sourcePosition: Position.Bottom,
    targetPosition: Position.Top,
    type: "card",
  });

  for (const entryPoint of bundle.data.entry_points.slice(0, 4)) {
    const nodeID = `entry:${entryPoint}`;
    nodes.push({
      data: {
        eyebrow: "Entry point",
        title: entryPoint,
        description: "Starting surface the reader should inspect early.",
        inspector: {
          eyebrow: "Entry point",
          title: entryPoint,
          description: "This path was identified as a meaningful project entry point.",
          notes: [`Recommended reading path hits: ${bundle.data.reading_path.filter((item) => item.path === entryPoint).length || 0}`],
          properties: [
            { label: "Path", value: entryPoint },
            { label: "Kind", value: "Entry point" },
          ],
        },
        tone: "entry",
      },
      id: nodeID,
      position: { x: 0, y: 0 },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
      type: "card",
    });
    edges.push({ id: `${rootNodeID}-${nodeID}`, source: rootNodeID, target: nodeID });
  }

  for (const module of modules) {
    const nodeID = `module:${module.path}`;
    nodes.push({
      data: {
        eyebrow: "Module",
        title: module.path,
        description: `${module.file_count} files · ${module.total_lines} lines · ${module.entry_point_count} entry point(s)`,
        inspector: {
          eyebrow: "Module",
          title: module.path,
          description: bundle.data.ai.architecture_narrative || "Structured module inventory for the active project.",
          notes: collectModuleNotes(bundle, module.path),
          properties: [
            { label: "Files", value: String(module.file_count) },
            { label: "Lines", value: String(module.total_lines) },
            { label: "Entry points", value: String(module.entry_point_count) },
          ],
        },
        tone: "module",
      },
      id: nodeID,
      position: { x: 0, y: 0 },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
      type: "card",
    });
    edges.push({ id: `${rootNodeID}-${nodeID}`, source: rootNodeID, target: nodeID });
  }

  for (const area of bundle.data.changes.frequently_mentioned_areas.slice(0, 4)) {
    const nodeID = `issue:${area.path}`;
    const parent = findBestParent(area.path, modules);
    nodes.push({
      data: {
        eyebrow: "Issue focus",
        title: area.path,
        description: `${area.mention_count} mention(s) · ${area.confidence}`,
        inspector: {
          eyebrow: "Issue focus",
          title: area.path,
          description: "This repository area keeps showing up across linked support files.",
          notes: [
            `Mention count: ${area.mention_count}`,
            `Confidence: ${area.confidence}`,
            bundle.data.changes.note || "No additional change note recorded.",
          ],
          properties: [
            { label: "Path", value: area.path },
            { label: "Confidence", value: area.confidence },
          ],
        },
        tone: "issue",
      },
      id: nodeID,
      position: { x: 0, y: 0 },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
      type: "card",
    });
    edges.push({ id: `${parent}-${nodeID}`, source: parent, target: nodeID });
  }

  for (const step of bundle.data.reading_path.slice(0, 4)) {
    const nodeID = `reading:${step.path}`;
    const parent = findBestParent(step.path, modules);
    nodes.push({
      data: {
        eyebrow: "Reading path",
        title: step.path,
        description: step.reason,
        inspector: {
          eyebrow: "Reading path",
          title: step.path,
          description: step.reason,
          notes: [findReadingRationale(bundle, step.path) || "No AI rationale recorded for this path."],
          properties: [
            { label: "Path", value: step.path },
            { label: "Reason", value: step.reason },
          ],
        },
        tone: "reading",
      },
      id: nodeID,
      position: { x: 0, y: 0 },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
      type: "card",
    });
    edges.push({ id: `${parent}-${nodeID}`, source: parent, target: nodeID });
  }

  return layoutGraph(nodes, edges);
}

function layoutGraph(nodes: FlowNode[], edges: Edge[]) {
  const graph = new dagre.graphlib.Graph();
  graph.setDefaultEdgeLabel(() => ({}));
  graph.setGraph({
    nodesep: 34,
    rankdir: "TB",
    ranksep: 92,
  });

  for (const node of nodes) {
    graph.setNode(node.id, { height: graphNodeHeight, width: graphNodeWidth });
  }

  for (const edge of edges) {
    graph.setEdge(edge.source, edge.target);
  }

  dagre.layout(graph);

  return {
    nodes: nodes.map((node) => {
      const position = graph.node(node.id);
      return {
        ...node,
        position: {
          x: position.x - graphNodeWidth / 2,
          y: position.y - graphNodeHeight / 2,
        },
      };
    }),
    edges,
  };
}

function findBestParent(path: string, modules: WorkbenchBundle["data"]["modules"]) {
  for (const module of modules) {
    if (path.startsWith(module.path)) {
      return `module:${module.path}`;
    }
  }
  return "project-root";
}

function collectModuleNotes(bundle: WorkbenchBundle, modulePath: string) {
  const notes: string[] = [];

  const hotspot = bundle.data.ai.hotspot_explanations.find((item) => item.path.startsWith(modulePath));
  if (hotspot) {
    notes.push(hotspot.explanation);
  }

  const area = bundle.data.changes.frequently_mentioned_areas.find((item) => item.path.startsWith(modulePath));
  if (area) {
    notes.push(`Frequently mentioned area with ${area.mention_count} mention(s).`);
  }

  if (!notes.length) {
    notes.push("No special hotspot note recorded for this module.");
  }

  return notes;
}

function findReadingRationale(bundle: WorkbenchBundle, path: string) {
  return bundle.data.ai.reading_path_explanations.find((item) => item.path === path)?.rationale || "";
}
