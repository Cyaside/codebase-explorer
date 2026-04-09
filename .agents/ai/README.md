# AI Orchestrator Guide

This is the master instruction file for any AI agent that needs to deeply explore a repository and produce structured outputs for Codebase Explorer.

Use this file as the orchestrator.

## Purpose

The agent must not behave like a generic chat summarizer.
It must behave like a repository orientation worker that:
- inspects the codebase directly
- gathers evidence before claiming structure
- fills each analysis surface with grounded information
- keeps outputs aligned with the product and bundle model

## Mandatory Read Order For The Agent

Before doing deep repo exploration, the agent must read:
1. `.agents/README.md`
2. `.agents/agents.md`
3. `.agents/ai/README.md`

Then the agent must read the function-specific files it needs under:
- `.agents/ai/functions/`

## Exploration Contract

The agent must:
- inspect the repository directly
- prefer evidence from actual files over assumptions
- identify uncertainty explicitly
- distinguish facts from interpretations
- keep outputs structured enough to be reused in the product

The agent must not:
- invent files, modules, flows, or incidents
- pretend it inspected files it did not read
- replace evidence with stylistic prose
- produce graph output that cannot be traced back to repository structure

## Global Execution Loop

The agent should work in this order:

1. Establish repo facts
   - detect language mix
   - detect entry points
   - detect top-level modules
   - detect important files and directories

2. Build evidence map
   - high-signal files
   - dependency concentration
   - hotspot candidates
   - reading-path candidates
   - support files if available

3. Run the function specs
   - summary
   - architecture
   - flowchart
   - issues
   - recommendations
   - dashboard assembly

4. Verify consistency
   - paths in recommendations must exist
   - flowchart nodes must map to real modules or files
   - issue claims must trace to support files or code evidence
   - architecture narrative must agree with flowchart and summary

5. Produce outputs
   - concise summaries
   - structured evidence-backed narratives
   - explicit uncertainties

## Function Guides

Use these documents as specialized playbooks:

- `.agents/ai/functions/dashboard.md`
- `.agents/ai/functions/summary.md`
- `.agents/ai/functions/architecture.md`
- `.agents/ai/functions/flowchart.md`
- `.agents/ai/functions/issues.md`
- `.agents/ai/functions/recommendations.md`
- `.agents/ai/functions/hotspots-and-dependencies.md`

## Output Quality Rules

Every function output must be:
- grounded
- concise
- implementation-focused
- path-aware
- usable without rereading the whole repository

Every function output should include:
- the main claim
- the evidence basis
- the most important paths
- a short uncertainty note when needed

## Default Function Order

When nothing else is specified, run in this order:
1. summary
2. architecture
3. hotspots-and-dependencies
4. flowchart
5. issues
6. recommendations
7. dashboard

## Relationship To The Product

This AI orchestrator does not authorize reckless repository summarization.

For the product:
- deterministic analysis remains the default baseline
- deeper AI exploration is optional when enabled by the runtime
- AI output must still be reusable, structured, and evidence-backed

## Done Condition

The agent is done only when:
- the repository has been explored enough to support all required function outputs
- each output is aligned with its function spec
- claims are evidence-backed
- uncertainty is explicitly surfaced where evidence is thin
