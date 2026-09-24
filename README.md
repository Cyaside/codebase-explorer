<div align="center">
  <img src="ui/public/brand-mark.svg" alt="Codebase Explorer logo" width="64" height="64" />
  <h1>Codebase Explorer</h1>
  <p><strong>Find your way through an unfamiliar repository.</strong></p>
  <p>A local workbench that turns a source checkout into an evidence-linked overview, reading path, and interactive graphs.</p>
  <p>
    <a href="go.mod"><img alt="Go 1.26" src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" /></a>
    <a href="ui/package.json"><img alt="React 19" src="https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white" /></a>
    <a href="LICENSE"><img alt="Apache 2.0 license" src="https://img.shields.io/badge/License-Apache%202.0-64748B" /></a>
    <a href="https://github.com/Cyaside/codebase-explorer/actions/workflows/ci.yml"><img alt="CI status" src="https://github.com/Cyaside/codebase-explorer/actions/workflows/ci.yml/badge.svg?branch=renovasi" /></a>
  </p>
  <p>
    <a href="#why-codebase-explorer">Why</a> ·
    <a href="#quick-start">Quick start</a> ·
    <a href="#how-it-works">How it works</a> ·
    <a href="#workbench-and-bundles">Workbench and bundles</a> ·
    <a href="#development">Development</a>
  </p>
</div>

https://github.com/user-attachments/assets/9fdedceb-da4f-42b4-801b-be9c74a5b5d0

## Why Codebase Explorer?

Getting oriented in a new codebase means finding its entry points, understanding how modules relate, and checking whether an explanation actually points to source. Codebase Explorer combines a local repository scan with an OpenAI-compatible model, then keeps the source paths behind its findings visible in the workbench.

It runs against a **local checkout**. You choose the endpoint and model; an API key is required before analysis begins.

## What you get

| View | What it helps you do |
| --- | --- |
| Overview and summary | See the repository shape, main language, entry points, and supported findings. |
| Architecture and reading path | Understand major components and where to start reading. |
| Graphs | Explore architecture, execution flow, and dependencies/impact through interactive, dark-theme diagrams. |
| Issues and recommendations | Review risks and next steps alongside their source references. |
| Inspector | Open the evidence and relationships behind a selected item. |

The workbench uses React Flow and ELKjs to lay out structured nodes and edges. The current graph UI does not render Mermaid.

## Quick start

**Requirements:** Go 1.26 or newer, a local repository checkout, and an OpenAI-compatible endpoint with a model and API key. Node.js is only needed when changing the frontend.

```bash
git clone https://github.com/Cyaside/codebase-explorer.git
cd codebase-explorer
go run ./cmd/codearch start
```

Open the local URL printed by the command if the browser does not open automatically. Then:

1. In **Connections**, enter the endpoint base URL, model, and API key. Use **Test model**, then **Save**. The key is stored by the local backend outside this repository, so you do not need to re-enter it for each run.
2. In **Project setup**, select a local repository path. Add support files such as an issue export or changelog if they are relevant.
3. Select **Analyze**. Follow progress in the workbench, then open the summary, architecture, graphs, and evidence inspector.

For a CLI run, use a saved connection:

```bash
go run ./cmd/codearch doctor --connection openai
go run ./cmd/codearch analyze <repo-path> --connection openai
```

Or provide the same connection through environment variables, for example in PowerShell:

```powershell
$env:CODEARCH_BASE_URL = "https://your-endpoint.example/v1"
$env:CODEARCH_MODEL = "your-model"
$env:CODEARCH_API_KEY = "your-key"
go run ./cmd/codearch doctor
go run ./cmd/codearch analyze C:\path\to\repository
```

The base URL, model, and key are all required. Keep the key in your local environment or the workbench connection form; never add it to a repository file.

## How it works

1. **Check the connection.** The model configuration and API key are validated before the repository is scanned.
2. **Scan and select evidence.** Local analysis finds entry points, modules, dependencies, hotspots, and relevant support files. It limits and redacts the material sent to the endpoint.
3. **Run the instruction pack.** Workers load the actual `.agents/` instructions. Compact context uses one structured model request; larger context can use two independent requests in parallel. The dashboard is assembled locally.
4. **Validate and assemble.** Returned sections are checked against the evidence paths that were read. Only a failed section may get one focused repair request. The run reports `succeeded`, `partial`, or `failed` rather than silently treating provider failure as success.
5. **Write a bundle.** Accepted results and scan data are written atomically with the instruction pack version/hash and run metrics.

Path and structure checks do **not** prove that every interpretation is semantically correct. Review important claims against the linked source.

## Workbench and bundles

The workbench serves on the local machine. Each completed analysis produces a portable bundle under `out/` by default. The bundle includes Markdown reports, structured JSON in `data/`, and a standalone viewer in `ui/`. The evidence manifest records paths and metadata without retaining the submitted source excerpts.

| Command | Purpose |
| --- | --- |
| `go run ./cmd/codearch start` | Start the local workbench. Use `--addr <host:port>` or `--no-browser` when needed. |
| `go run ./cmd/codearch analyze <repo-path>` | Analyze a checkout with a saved or environment connection. Accepts `--connection`, `--support`, `--issues`, `--changelog`, `--output`, and `--ignore`. |
| `go run ./cmd/codearch doctor` | Check the connection and local setup before a run. |
| `go run ./cmd/codearch open [bundle-path]` | Open the latest or a selected bundle. |
| `go run ./cmd/codearch export [bundle-path] --output <zip-path>` | Package an existing bundle without analyzing again. |
| `go run ./cmd/codearch cache clear` | Clear the analysis cache. |

Older bundles remain readable in the current viewer. Generated bundles and cache directories are ignored by Git. The default output retention is ten bundles; `CODEARCH_OUTPUT_ROOT` and `CODEARCH_OUTPUT_KEEP` change the output location and retention.

## Data and trust

- The workbench stores the API key in its local backend and does not return the saved key to the browser or write it into a bundle.
- Common secret files are excluded, and selected excerpts are checked for sensitive content before they are sent.
- Large files and total request size are bounded internally. Bundle metadata records the evidence manifest and request size.
- Model output is accepted only when its references pass the available evidence and structure checks. Partial results remain visibly partial.
- The endpoint you configure receives selected repository evidence. Use an endpoint you trust for the repository you analyze.

## Development

The Go binary embeds the built workbench assets, so a normal `go run` does not need an npm install. To change the frontend, use Node.js and npm:

```bash
cd ui
npm ci
npm run check
npm run build
```

Then run the backend checks from the repository root:

```bash
go test ./...
go vet ./...
```

The [CI workflow](.github/workflows/ci.yml) runs Go tests, vet, and a build on Linux and Windows. It also typechecks and builds the React workbench. It runs on pushes to `main` or `renovasi`, pull requests to `main`, and manual dispatch.

The project is licensed under [Apache 2.0](LICENSE).
