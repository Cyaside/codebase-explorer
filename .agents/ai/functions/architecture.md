# Architecture Function Guide

## Goal

Describe the structural shape of the system in a way that is useful for onboarding and reasoning.

## Questions To Answer

- what are the main modules or subsystems
- how do responsibilities appear to be split
- where do entry points connect into deeper code
- what are the likely boundaries and seams

## Evidence To Inspect

Prioritize:
- entry points
- routing/bootstrap files
- module boundaries
- dependency-heavy files
- service composition points
- config that reveals runtime wiring

## Output Shape

The architecture output should include:
- one narrative paragraph
- a list of major modules
- notable boundaries or seams
- uncertainty notes where wiring is ambiguous

## Do Not

- confuse folder names with true runtime boundaries
- invent service relationships not supported by imports or call structure
- turn architecture into a prose-only essay

## Done Condition

The architecture writeup agrees with the visible code structure and helps a reader choose where to inspect next.
