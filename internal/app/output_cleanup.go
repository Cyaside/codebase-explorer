package app

import "fmt"

func cleanupStatus(prunedBundles int) string {
	if prunedBundles > 0 {
		return "pruned"
	}
	return "kept"
}

func cleanupDetail(prunedBundles, retentionLimit int) string {
	if retentionLimit <= 0 {
		return "automatic bundle cleanup disabled"
	}
	if prunedBundles > 0 {
		return fmt.Sprintf("removed %d older bundle(s); keeping latest %d", prunedBundles, retentionLimit)
	}
	return fmt.Sprintf("no older bundle removed; keeping latest %d", retentionLimit)
}
