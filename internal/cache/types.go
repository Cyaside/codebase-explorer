package cache

import (
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

const (
	StatusDisabled = "disabled"
	StatusHit      = "hit"
	StatusMiss     = "miss"
)

type DeterministicPayload struct {
	CachedAt   time.Time       `json:"cached_at"`
	Key        string          `json:"key"`
	Scan       repo.ScanResult `json:"scan"`
	Analysis   analyzer.Result `json:"analysis"`
	Changes    changes.Result  `json:"changes"`
	AppVersion string          `json:"app_version"`
}

type AIExecutionPayload struct {
	CachedAt  time.Time        `json:"cached_at"`
	Key       string           `json:"key"`
	Execution fullai.Execution `json:"execution"`
}

type ClearResult struct {
	CacheRoot      string
	RemovedEntries int
}
