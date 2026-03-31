package installer

import (
	"sync"

	"github.com/chenasraf/sofmani/logger"
)

var (
	repoUpdateMu   sync.Mutex
	repoUpdateDone = map[string]bool{}
	repoUpdateErr  = map[string]error{}
)

// RunRepoUpdateOnce runs fn at most once per key during the process lifetime.
// Subsequent calls with the same key return the cached error without running fn again.
func RunRepoUpdateOnce(key string, fn func() error) error {
	repoUpdateMu.Lock()
	defer repoUpdateMu.Unlock()
	if repoUpdateDone[key] {
		logger.Debug("Repo update already done for %s, skipping", key)
		return repoUpdateErr[key]
	}
	err := fn()
	repoUpdateDone[key] = true
	repoUpdateErr[key] = err
	return err
}

// MarkRepoUpdated marks a key as done without running a function.
func MarkRepoUpdated(key string) {
	repoUpdateMu.Lock()
	defer repoUpdateMu.Unlock()
	repoUpdateDone[key] = true
}

// IsRepoUpdated returns whether a key has been marked as done.
func IsRepoUpdated(key string) bool {
	repoUpdateMu.Lock()
	defer repoUpdateMu.Unlock()
	return repoUpdateDone[key]
}

// ResetRepoUpdateTracker resets the tracker state. Intended for testing.
func ResetRepoUpdateTracker() {
	repoUpdateMu.Lock()
	defer repoUpdateMu.Unlock()
	repoUpdateDone = map[string]bool{}
	repoUpdateErr = map[string]error{}
}
