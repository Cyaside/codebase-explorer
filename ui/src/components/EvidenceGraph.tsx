import { useEffect, useMemo, useState } from "react";
import type { ELK as ElkAPI, ElkNode } from "elkjs/lib/elk.bundled.js";
import {
  Background,
  BaseEdge,
  Controls,
  EdgeText,
  Handle,
  MarkerType,
  MiniMap,
  Position,
  ReactFlow,
  getSmoothStepPath,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps,
  type ReactFlowInstance,
} from "@xyflow/react";

import type { GraphEdge, GraphNode, GraphView, InspectorState, WorkbenchBundle } from "@/lib/types";

let elkInstance: Promise<ElkAPI> | null = null;
function getElk() {
  elkInstance ??= import("elkjs/lib/elk.bundled.js").then(({ default: ELK }) => new ELK());
  return elkInstance;
}
const nodeWidth = 226;
const nodeHeight = 102;
const viewLabels: Record<string, string> = {
  architecture: "Architecture",
  flow: "Execution flow",
  dependencies: "Dependencies & impact",
};

interface CanvasNodeData extends Record<string, unknown> {
  label: string;
  type: string;
  path: string;
  evidencePaths: string[];
  faded: boolean;
}

interface CanvasEdgeData extends Record<string, unknown> {
  evidencePaths: string[];
  relation: string;
  sourceKind: string;
  route?: Array<{ x: number; y: number }>;
  faded: boolean;
  flowing: boolean;
}

type CanvasNode = Node<CanvasNodeData, "evidence">;
type CanvasEdge = Edge<CanvasEdgeData, "routed">;

const nodeTypes = { evidence: EvidenceNode };
const edgeTypes = { routed: RoutedEdge };

function roundedRoute(points: Array<{ x: number; y: number }>) {
  let path = `M ${points[0].x} ${points[0].y}`;
  for (let index = 1; index < points.length - 1; index += 1) {
    const previous = points[index - 1];
    const current = points[index];
    const next = points[index + 1];
    const before = Math.hypot(current.x - previous.x, current.y - previous.y);
    const after = Math.hypot(next.x - current.x, next.y - current.y);
    if (!before || !after) continue;
    const radius = Math.min(12, before / 2, after / 2);
    const start = { x: current.x - (current.x - previous.x) * radius / before, y: current.y - (current.y - previous.y) * radius / before };
    const end = { x: current.x + (next.x - current.x) * radius / after, y: current.y + (next.y - current.y) * radius / after };
    path += ` L ${start.x} ${start.y} Q ${current.x} ${current.y} ${end.x} ${end.y}`;
  }
  const last = points[points.length - 1];
  return `${path} L ${last.x} ${last.y}`;
}

function EvidenceNode({ data, selected }: NodeProps<CanvasNode>) {
  return (
    <div className={`evidence-node${selected ? " evidence-node-selected" : ""}${data.faded ? " evidence-node-faded" : ""}`} data-kind={data.type}>
      <Handle className="graph-node-handle" position={Position.Left} type="target" />
      <span className="evidence-node-type"><span className="evidence-node-dot" aria-hidden="true" />{data.type.replaceAll("-", " ")}</span>
      <strong title={data.label}>{data.label}</strong>
      <small title={data.path || data.evidencePaths.join(", ")}>{data.path || `${data.evidencePaths.length} evidence path(s)`}</small>
      <Handle className="graph-node-handle" position={Position.Right} type="source" />
    </div>
  );
}

function RoutedEdge({ id, sourceX, sourceY, targetX, targetY, data, markerEnd }: EdgeProps<CanvasEdge>) {
  const [fallbackPath] = getSmoothStepPath({ sourceX, sourceY, targetX, targetY });
  const route = data?.route;
  const path = route?.length && route.length >= 2
    ? roundedRoute(route)
    : fallbackPath;
  const center = route?.length ? route[Math.floor(route.length / 2)] : { x: (sourceX + targetX) / 2, y: (sourceY + targetY) / 2 };

  return (
    <>
      <BaseEdge id={id} markerEnd={markerEnd} path={path} style={{ opacity: data?.faded ? 0.12 : 0.9, stroke: "#698b96", strokeWidth: 1.5 }} />
      {!data?.faded && data?.flowing ? <path className="graph-edge-flow" d={path} fill="none" pointerEvents="none" /> : null}
      {!data?.faded && data?.relation ? (
        <EdgeText
          x={center.x}
          y={center.y}
          label={data.relation}
          labelBgPadding={[5, 3]}
          labelBgBorderRadius={4}
          labelBgStyle={{ fill: "#172025", fillOpacity: 0.98, stroke: "#35464d", strokeWidth: 1 }}
          labelStyle={{ fill: "#abc4c8", fontSize: 10 }}
        />
      ) : null}
    </>
  );
}

function neighborhood(view: GraphView, focusID: string) {
  const ids = new Set([focusID]);
  for (const edge of view.edges) {
    if (edge.source === focusID) ids.add(edge.target);
    if (edge.target === focusID) ids.add(edge.source);
  }
  return ids;
}

function defaultFocus(view: GraphView) {
  const degree = new Map<string, number>();
  for (const edge of view.edges) {
    degree.set(edge.source, (degree.get(edge.source) || 0) + 1);
    degree.set(edge.target, (degree.get(edge.target) || 0) + 1);
  }
  return [...view.nodes].sort((left, right) => (degree.get(right.id) || 0) - (degree.get(left.id) || 0))[0]?.id || "";
}

async function layout(view: GraphView, shown: Set<string>) {
  const nodes = view.nodes.filter((node) => shown.has(node.id));
  const edges = view.edges.filter((edge) => shown.has(edge.source) && shown.has(edge.target));
  const input: ElkNode = {
    id: "root",
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": "RIGHT",
      "elk.edgeRouting": "ORTHOGONAL",
      "elk.spacing.nodeNode": "34",
      "elk.layered.spacing.nodeNodeBetweenLayers": "130",
      "elk.layered.crossingMinimization.strategy": "LAYER_SWEEP",
    },
    children: nodes.map((node) => ({ id: node.id, width: nodeWidth, height: nodeHeight })),
    edges: edges.map((edge) => ({ id: edge.id, sources: [edge.source], targets: [edge.target] })),
  };
  const result = await (await getElk()).layout(input);
  const locations = new Map(result.children?.map((node) => [node.id, node]) || []);
  const routes = new Map(result.edges?.map((edge) => {
    const section = edge.sections?.[0];
    return [edge.id, section ? [section.startPoint, ...(section.bendPoints || []), section.endPoint] : undefined] as const;
  }) || []);

  const canvasNodes: CanvasNode[] = nodes.map((node) => ({
    id: node.id,
    type: "evidence",
    position: { x: locations.get(node.id)?.x || 0, y: locations.get(node.id)?.y || 0 },
    initialWidth: nodeWidth,
    initialHeight: nodeHeight,
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
    data: {
      label: node.label,
      type: node.type,
      path: node.path,
      evidencePaths: node.evidence_paths,
      faded: false,
    },
  }));
  const canvasEdges: CanvasEdge[] = edges.map((edge, index) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    type: "routed",
    markerEnd: { type: MarkerType.ArrowClosed, color: "#698b96", width: 14, height: 14 },
    data: {
      relation: edge.relation,
      evidencePaths: edge.evidence_paths,
      sourceKind: edge.source_kind,
      route: routes.get(edge.id),
      faded: false,
      flowing: index < 64,
    },
  }));
  return { nodes: canvasNodes, edges: canvasEdges };
}

function nodeInspector(node: GraphNode, view: GraphView): InspectorState {
  const adjacent = view.edges.filter((edge) => edge.source === node.id || edge.target === node.id);
  return {
    eyebrow: viewLabels[view.id] || "Graph",
    title: node.label,
    description: `${node.type.replaceAll("-", " ")} with ${adjacent.length} visible relationship(s).`,
    notes: node.evidence_paths.length ? node.evidence_paths : ["This node comes from the repository scan."],
    properties: [
      { label: "Path", value: node.path || "—" },
      { label: "Relationships", value: String(adjacent.length) },
    ],
  };
}

function edgeInspector(edge: GraphEdge, view: GraphView): InspectorState {
  const source = view.nodes.find((node) => node.id === edge.source)?.label || edge.source;
  const target = view.nodes.find((node) => node.id === edge.target)?.label || edge.target;
  return {
    eyebrow: "Relationship",
    title: `${source} → ${target}`,
    description: edge.relation,
    notes: edge.evidence_paths.length ? edge.evidence_paths : ["No evidence path attached."],
    properties: [{ label: "Source", value: edge.source_kind }, { label: "View", value: viewLabels[view.id] || view.id }],
  };
}

export function EvidenceGraph({ bundle }: { bundle: WorkbenchBundle | null }) {
  const views = bundle?.data.graphs.views || [];
  const [viewID, setViewID] = useState("architecture");
  const [focusID, setFocusID] = useState("");
  const [query, setQuery] = useState("");
  const [canvas, setCanvas] = useState<{ nodes: CanvasNode[]; edges: CanvasEdge[] }>({ nodes: [], edges: [] });
  const [layoutError, setLayoutError] = useState("");
  const [layoutBusy, setLayoutBusy] = useState(false);
  const [selectedInspector, setSelectedInspector] = useState<InspectorState | null>(null);
  const [flowInstance, setFlowInstance] = useState<ReactFlowInstance<CanvasNode, CanvasEdge> | null>(null);
  const view = views.find((item) => item.id === viewID) || views[0];
  const effectiveFocus = view?.nodes.some((node) => node.id === focusID) ? focusID : view ? defaultFocus(view) : "";
  const limited = !!view && view.nodes.length > 60;
  const layoutFocus = limited ? effectiveFocus : "";
  const shown = useMemo(() => {
    if (!view) return new Set<string>();
    if (!limited) return new Set(view.nodes.map((node) => node.id));
    const neighbors = neighborhood(view, layoutFocus);
    const ordered = view.nodes.filter((node) => neighbors.has(node.id) && node.id !== layoutFocus).sort((a, b) => a.label.localeCompare(b.label));
    return new Set([layoutFocus, ...ordered.slice(0, 59).map((node) => node.id)]);
  }, [view, limited, layoutFocus]);
  const highlighted = useMemo(() => {
    if (!view || !focusID) return new Set<string>();
    return neighborhood(view, focusID);
  }, [view, focusID]);
  const displayCanvas = useMemo(() => ({
    nodes: canvas.nodes.map((node) => ({ ...node, data: { ...node.data, faded: highlighted.size > 0 && !highlighted.has(node.id) } })),
    edges: canvas.edges.map((edge) => ({ ...edge, data: { evidencePaths: edge.data?.evidencePaths || [], relation: edge.data?.relation || "", sourceKind: edge.data?.sourceKind || "", route: edge.data?.route, flowing: edge.data?.flowing || false, faded: highlighted.size > 0 && !(highlighted.has(edge.source) && highlighted.has(edge.target)) } })),
  }), [canvas, highlighted]);

  useEffect(() => {
    let current = true;
    if (!view || view.nodes.length === 0) {
      setCanvas({ nodes: [], edges: [] });
      setLayoutError("");
      setLayoutBusy(false);
      return;
    }
    setCanvas({ nodes: [], edges: [] });
    setLayoutBusy(true);
    void layout(view, shown).then((result) => {
      if (current) {
        setCanvas(result);
        setLayoutError("");
        setLayoutBusy(false);
      }
    }).catch(() => {
      if (current) {
        setLayoutError("Graph layout failed. Select another view or reopen this bundle.");
        setLayoutBusy(false);
      }
    });
    return () => { current = false; };
  }, [view, shown]);

  useEffect(() => {
    if (!focusID || !flowInstance) return;
    const focusedNode = canvas.nodes.find((node) => node.id === focusID);
    if (!focusedNode) return;
    void flowInstance.setCenter(focusedNode.position.x + nodeWidth / 2, focusedNode.position.y + nodeHeight / 2, { zoom: 0.9, duration: 260 });
  }, [canvas.nodes, flowInstance, focusID]);

  if (!bundle) return <div className="graph-shell grid place-items-center text-sm text-zinc-500">Run analysis to create evidence-backed graph views.</div>;
  if (!views.length) return <div className="graph-shell grid place-items-center px-6 text-center text-sm text-zinc-500">This older bundle has no structured graph. Create a new analysis to explore architecture, flow, and dependencies.</div>;

  const matches = view?.nodes.filter((node) => `${node.label} ${node.path}`.toLowerCase().includes(query.toLowerCase())).slice(0, 12) || [];

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <div className="graph-view-switch" role="group" aria-label="Graph view">
          {views.map((item) => (
          <button
            aria-pressed={item.id === view?.id}
            key={item.id}
            onClick={() => { setFlowInstance(null); setViewID(item.id); setFocusID(""); setQuery(""); setSelectedInspector(null); }}
            type="button"
          >{viewLabels[item.id] || item.id}</button>
          ))}
        </div>
        <span className="graph-view-count ml-auto">{shown.size}/{view.nodes.length} nodes · {canvas.edges.length}/{view.edges.length} relationships</span>
      </div>
      <div className="flex flex-wrap items-start gap-2">
        <div className="min-w-56 flex-1">
          <input
            aria-label="Find graph node"
            className="compact-input !py-2"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Find a module, file, or flow step"
            type="search"
            value={query}
          />
          {query ? (
            <div className="mt-1 flex flex-wrap gap-1">
              {matches.map((node) => <button className="rounded border border-zinc-800 px-2 py-1 text-xs text-zinc-300 hover:border-zinc-500" key={node.id} onClick={() => { setFocusID(node.id); setQuery(""); setSelectedInspector(nodeInspector(node, view)); }} type="button">{node.label}</button>)}
              {!matches.length ? <span className="text-xs text-zinc-500">No matching node.</span> : null}
            </div>
          ) : null}
        </div>
        {focusID ? <button className="secondary-control !py-2" onClick={() => setFocusID("")} type="button">Clear focus</button> : null}
      </div>
      {limited ? <p className="text-xs text-zinc-500">Large graph: showing up to 60 nodes around the selected node. Search for any other node to explore its neighbors.</p> : null}
      {view.note ? <p className="text-xs text-zinc-500">{view.note}</p> : null}
      <div className="graph-workspace">
      <div className="graph-shell">
        {layoutError ? <div className="grid h-full place-items-center text-sm text-red-300">{layoutError}</div> : layoutBusy ? (
          <div className="grid h-full place-items-center text-sm text-zinc-500">Arranging relationships…</div>
        ) : !view.nodes.length ? (
          <div className="grid h-full place-items-center px-6 text-center text-sm text-zinc-500">No evidence-backed relationships are available in this view.</div>
        ) : (
          <ReactFlow<CanvasNode, CanvasEdge>
            key={`${bundle.summary.name}:${view.id}:${limited ? effectiveFocus : "all"}:${canvas.nodes.length}`}
            className="codearch-flow"
            colorMode="dark"
            defaultViewport={{ x: 38, y: 180, zoom: 0.88 }}
            edgeTypes={edgeTypes}
            edges={displayCanvas.edges}
            maxZoom={1.7}
            minZoom={0.1}
            nodeTypes={nodeTypes}
            nodes={displayCanvas.nodes}
            nodesDraggable={false}
            nodesFocusable
            onInit={setFlowInstance}
            onEdgeClick={(_, edge) => { const original = view.edges.find((item) => item.id === edge.id); if (original) setSelectedInspector(edgeInspector(original, view)); }}
            onNodeClick={(_, node) => { setFocusID(node.id); const original = view.nodes.find((item) => item.id === node.id); if (original) setSelectedInspector(nodeInspector(original, view)); }}
            panOnDrag
            proOptions={{ hideAttribution: true }}
          >
            <Background color="#27343a" gap={24} size={1} />
            <Controls fitViewOptions={{ maxZoom: 1, padding: 0.15 }} showInteractive={false} />
            {canvas.nodes.length > 1 ? (
              <MiniMap
                ariaLabel="Graph overview; drag to navigate"
                bgColor="#172125"
                className="graph-minimap"
                maskColor="#0d1417b8"
                maskStrokeColor="#9cc7cf"
                maskStrokeWidth={2}
                nodeBorderRadius={3}
                nodeColor="#547783"
                nodeStrokeColor="#94bac2"
                pannable
              />
            ) : null}
          </ReactFlow>
        )}
      </div>
      {selectedInspector ? (
        <aside aria-label="Graph details" className="graph-inspector">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-xs text-zinc-500">{selectedInspector.eyebrow}</p>
              <h3 className="mt-1 break-all text-sm font-semibold text-zinc-100">{selectedInspector.title}</h3>
            </div>
            <button className="text-xs text-zinc-400 hover:text-white" onClick={() => setSelectedInspector(null)} type="button">Close</button>
          </div>
          <p className="mt-2 text-sm text-zinc-300">{selectedInspector.description}</p>
          <dl className="mt-3 flex flex-wrap gap-x-6 gap-y-2 text-xs">{selectedInspector.properties.map((item) => <div key={item.label}><dt className="text-zinc-500">{item.label}</dt><dd className="mt-1 text-zinc-200">{item.value}</dd></div>)}</dl>
          <p className="mt-3 text-xs font-medium text-zinc-400">Evidence</p>
          <ul className="mt-1 space-y-1 font-mono text-xs text-zinc-300">{selectedInspector.notes.map((note) => <li className="break-all" key={note}>{note}</li>)}</ul>
        </aside>
      ) : (
        <aside aria-label="Graph details" className="graph-inspector">
          <p className="panel-kicker">Inspector</p>
          <h3 className="mt-3">Explore the repository</h3>
          <p className="mt-3 text-xs leading-6">Select a node or connection to see its evidence and relationship details. Drag the canvas to follow the flow.</p>
          <div className="graph-inspector-divider mt-7 border-t pt-5">
            <p className="panel-kicker">Current view</p>
            <p className="mt-2 text-sm font-semibold">{viewLabels[view.id] || view.id}</p>
            <p className="mt-1 text-xs leading-5">{view.note || "Relationships are drawn from the accepted analysis and repository scan."}</p>
          </div>
        </aside>
      )}
      </div>
    </div>
  );
}
