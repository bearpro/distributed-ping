package node

// state.go contains small mutable-state helpers for node-local bookkeeping.

import (
	"sort"
	"time"

	"github.com/bearpro/distributed-ping/model"
)

func sortSubmittedRequests(items []model.SubmittedPingRequest) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Request.CreatedAt.After(items[j].Request.CreatedAt)
	})
}

func (ctx *Context) trackSubmitted(req model.PingRequest) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	current := ctx.State.submitted[req.ID]
	current.Request = req
	ctx.State.submitted[req.ID] = current
}

func (ctx *Context) updateCachedResults(requestID string, results []model.PingResult) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	current, ok := ctx.State.submitted[requestID]
	if !ok {
		return
	}
	current.CachedResults = append([]model.PingResult(nil), results...)
	current.LastResultsSyncAt = time.Now().UTC()
	ctx.State.submitted[requestID] = current
}

func (ctx *Context) appendTrackedResult(result model.PingResult) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	current, ok := ctx.State.submitted[result.RequestID]
	if !ok {
		return
	}

	replaced := false
	for i := range current.CachedResults {
		if current.CachedResults[i].Key() != result.Key() {
			continue
		}
		current.CachedResults[i] = result
		replaced = true
		break
	}
	if !replaced {
		current.CachedResults = append(current.CachedResults, result)
	}
	current.LastResultsSyncAt = time.Now().UTC()
	ctx.State.submitted[result.RequestID] = current
}

func (ctx *Context) hasKnownOffset() bool {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	return ctx.State.offsetKnown
}

func (ctx *Context) setOffset(offset int) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	ctx.State.upstreamOffset = offset
	ctx.State.offsetKnown = true
}

func (ctx *Context) markExecuted(requestID string) bool {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	if _, exists := ctx.State.executedRequests[requestID]; exists {
		return false
	}
	ctx.State.executedRequests[requestID] = time.Now().UTC()
	return true
}

func (ctx *Context) resultRelayed(key string) bool {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	_, exists := ctx.State.relayedResults[key]
	return exists
}

func (ctx *Context) markRelayed(key string) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	ctx.State.relayedResults[key] = time.Now().UTC()
}
