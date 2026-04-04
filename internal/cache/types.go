package cache

import (
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/provider"
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

type ProviderPayload struct {
	CachedAt   time.Time       `json:"cached_at"`
	Key        string          `json:"key"`
	Result     provider.Result `json:"result"`
	AppVersion string          `json:"app_version"`
}

type ClearResult struct {
	CacheRoot      string
	RemovedEntries int
}
