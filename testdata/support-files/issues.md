# Issue Notes

## internal/app bootstrap race

internal/app still fails during startup when configuration loading lags behind initialization.

## internal/core parser regression

internal/core keeps resurfacing in parser cleanup work and needs closer review.
