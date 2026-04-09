# Dashboard Function Guide

## Goal

Build the dashboard as the compressed control-plane view of the current project.

It should answer:
- what this project is
- what the most important activity surfaces are
- where the user should look next

## Inputs

Read from:
- summary output
- architecture output
- hotspots and dependencies output
- issues output
- recommendations output

## What The Agent Must Do

- identify the top-level project identity
- surface a few high-signal metrics
- highlight the most urgent or important areas
- provide a short view of recent or notable signals

## Evidence Priority

1. real repository facts
2. issue/support-file signals
3. architecture and hotspot findings
4. recommendations

## Output Shape

The dashboard should include:
- project identity
- top metrics
- top hotspots
- top issue signals if present
- top recommendations

## Do Not

- invent dashboard metrics that are not backed by repo evidence
- overload the dashboard with every detail
- duplicate whole sections from deeper tabs

## Done Condition

The dashboard feels like a compact entry surface, not a dump.
