package app

import (
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func makeModules(count int) []analyzer.ModuleInfo {
	modules := make([]analyzer.ModuleInfo, 0, count)
	for index := 0; index < count; index++ {
		modules = append(modules, analyzer.ModuleInfo{
			Path:       "module/path",
			FileCount:  index + 1,
			TotalLines: 100 + index,
			Languages:  []string{"Go"},
		})
	}
	return modules
}

func makeHotspots(count int) []analyzer.Hotspot {
	hotspots := make([]analyzer.Hotspot, 0, count)
	for index := 0; index < count; index++ {
		hotspots = append(hotspots, analyzer.Hotspot{
			Path:  "hotspot/path",
			Score: float64(index),
		})
	}
	return hotspots
}

func makeDependencyRisks(count int) []analyzer.DependencyRisk {
	risks := make([]analyzer.DependencyRisk, 0, count)
	for index := 0; index < count; index++ {
		risks = append(risks, analyzer.DependencyRisk{
			Path:        "dependency/path",
			ImportCount: index + 1,
			Reason:      "reason",
		})
	}
	return risks
}

func makeReadingPath(count int) []analyzer.ReadingPathItem {
	items := make([]analyzer.ReadingPathItem, 0, count)
	for index := 0; index < count; index++ {
		items = append(items, analyzer.ReadingPathItem{
			Path:   "reading/path",
			Reason: "reason",
		})
	}
	return items
}

func testTime() time.Time {
	return time.Date(2026, time.April, 3, 15, 0, 0, 0, time.UTC)
}
