# Agents Pack

This folder contains the AI instruction pack that ships with Codebase Explorer.

## Purpose

`.agents/` is embedded in the application and sent to AI workers with selected repository evidence. It tells workers how to:

- inspect supplied repository evidence
- cite the files supporting each claim
- produce structured outputs for each major analysis surface
- avoid hallucinating architecture or issue claims

## Structure

- `.agents/agents.md`
  top-level agent contract
- `.agents/ai/README.md`
  orchestrator for deep repository exploration
- `.agents/ai/functions/`
  function-specific instructions for summary, architecture, flowchart, issues, recommendations, dashboard, and hotspots
