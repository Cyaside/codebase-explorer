# Recommendations Function Guide

## Goal

Give the developer a practical next-step plan for exploring the repository.

## Questions To Answer

- what should the user read first
- what should they inspect after that
- what are the riskiest files or modules
- what sequence best reduces confusion quickly

## Evidence To Inspect

Prioritize:
- summary
- entry points
- architecture modules
- hotspot analysis
- dependency concentration
- issue/change signals

## Output Shape

Recommendations should include:
- reading path order
- short rationale per path
- likely high-value follow-up actions
- warnings about ambiguity or risk where relevant

## Recommendation Rules

- paths must exist
- rationales must be short and specific
- order must make onboarding easier
- recommendations must not depend on hidden assumptions

## Do Not

- recommend reading random low-signal files first
- produce a reading path without rationale
- recommend based only on file size

## Done Condition

A user can follow the recommendations as a concrete onboarding path.
