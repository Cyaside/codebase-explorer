# Flowchart Function Guide

## Goal

Produce a graph-friendly model of how the repository is structured or how execution likely flows.

## Questions To Answer

- what are the major graph nodes
- what are the major edges
- which nodes are entry-oriented
- which nodes are module-oriented
- which nodes represent issue or reading-path signals

## Evidence To Inspect

Prioritize:
- entry points
- import relationships
- routing or command dispatch
- initialization files
- service wiring
- module groupings

## Graph Rules

Nodes should map to real things:
- project
- module
- entry point
- hotspot
- issue signal
- reading-path checkpoint

Edges should mean something explicit:
- contains
- depends on
- dispatches to
- starts at
- highlights

## Do Not

- create decorative edges with no semantic meaning
- invent runtime flow from naming alone
- produce graph nodes that cannot be traced back to paths

## Done Condition

The graph is grounded enough that a developer can click around it and trust what it represents.
