# Hotspots And Dependencies Function Guide

## Goal

Explain where the repository looks concentrated, risky, central, or likely to deserve attention.

## Questions To Answer

- which files look central or risky
- which modules are dependency-heavy
- which files combine size, imports, or markers
- which areas deserve extra caution during onboarding

## Evidence To Inspect

Prioritize:
- hotspot candidates
- import concentration
- markers like TODO, FIXME, HACK
- entry points
- repeatedly mentioned issue areas

## Output Shape

This function should produce:
- ranked hotspot candidates
- short explanations
- dependency concentration notes
- overlaps with issue/change signals where available

## Do Not

- equate “large file” with “important file” without additional evidence
- treat every import-heavy file as risky by default
- ignore the relationship between hotspots and actual module roles

## Done Condition

The hotspot and dependency output helps the user decide where risk review and deeper reading should happen.
