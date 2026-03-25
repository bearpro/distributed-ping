package controller

// controller.go keeps the in-memory controller state: connected nodes, ordered
// ping requests, and deduplicated ping results.

import (
	"sort"
	"sync"
	"time"

	"github.com/bearpro/distributed-ping/model"
)

type Configuration struct {
	StaleNodeAfter time.Duration
}

type State struct {
	mu          sync.RWMutex
	nodes       map[string]model.ConnectedNode
	requests    []model.PingRequest
	requestByID map[string]model.PingRequest
	results     map[string]map[string]model.PingResult
}

type Context struct {
	Config Configuration
	State  *State
}

type Snapshot struct {
	Config Configuration `json:"config"`
	State  StateSnapshot `json:"state"`
}

type StateSnapshot struct {
	ConnectedNodes []model.ConnectedNode `json:"connected_nodes"`
	PingRequests   []model.PingRequest   `json:"ping_requests"`
	PingResults    []model.PingResult    `json:"ping_results"`
}

func NewContext(config Configuration) *Context {
	return &Context{
		Config: config,
		State: &State{
			nodes:       make(map[string]model.ConnectedNode),
			requestByID: make(map[string]model.PingRequest),
			results:     make(map[string]map[string]model.PingResult),
		},
	}
}

func (ctx *Context) Snapshot() Snapshot {
	ctx.State.mu.RLock()
	requests := make([]model.PingRequest, len(ctx.State.requests))
	copy(requests, ctx.State.requests)
	ctx.State.mu.RUnlock()

	return Snapshot{
		Config: ctx.Config,
		State: StateSnapshot{
			ConnectedNodes: ctx.NodesSnapshot(),
			PingRequests:   requests,
			PingResults:    ctx.AllResults(),
		},
	}
}

func (ctx *Context) RegisterNode(identity model.NodeIdentity, now time.Time) (model.ConnectedNode, int) {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	connected := model.ConnectedNode{
		NodeIdentity: identity,
		LastContact:  now.UTC(),
	}
	ctx.State.nodes[identity.NodeID] = connected

	return connected, len(ctx.State.requests)
}

func (ctx *Context) CleanupStaleNodes(now time.Time) int {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	removed := 0
	for nodeID, info := range ctx.State.nodes {
		if now.Sub(info.LastContact) <= ctx.Config.StaleNodeAfter {
			continue
		}
		delete(ctx.State.nodes, nodeID)
		removed++
	}

	return removed
}

func (ctx *Context) CurrentOffset() int {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	return len(ctx.State.requests)
}

func (ctx *Context) CreateRequest(target string, sourceNodeID string, now time.Time, requestID string) model.PingRequest {
	req := model.PingRequest{
		ID:           requestID,
		Target:       target,
		CreatedAt:    now.UTC(),
		CollectUntil: now.UTC().Add(2 * time.Minute),
		SourceNodeID: sourceNodeID,
	}
	ctx.StoreRequest(req)
	return req
}

func (ctx *Context) StoreRequest(req model.PingRequest) bool {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	if _, exists := ctx.State.requestByID[req.ID]; exists {
		return false
	}

	req.CreatedAt = req.CreatedAt.UTC()
	req.CollectUntil = req.CollectUntil.UTC()
	ctx.State.requests = append(ctx.State.requests, req)
	ctx.State.requestByID[req.ID] = req

	return true
}

func (ctx *Context) RequestsFrom(offset int, limit int) ([]model.PingRequest, int) {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}
	if offset > len(ctx.State.requests) {
		offset = len(ctx.State.requests)
	}
	if limit <= 0 {
		limit = 100
	}

	end := offset + limit
	if end > len(ctx.State.requests) {
		end = len(ctx.State.requests)
	}

	items := make([]model.PingRequest, end-offset)
	copy(items, ctx.State.requests[offset:end])

	return items, end
}

func (ctx *Context) RequestByID(requestID string) (model.PingRequest, bool) {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	req, ok := ctx.State.requestByID[requestID]
	return req, ok
}

func (ctx *Context) StoreResult(result model.PingResult) bool {
	ctx.State.mu.Lock()
	defer ctx.State.mu.Unlock()

	if _, exists := ctx.State.requestByID[result.RequestID]; !exists {
		return false
	}

	result.ObservedAt = result.ObservedAt.UTC()
	byRequest, ok := ctx.State.results[result.RequestID]
	if !ok {
		byRequest = make(map[string]model.PingResult)
		ctx.State.results[result.RequestID] = byRequest
	}
	byRequest[result.Executor.NodeID] = result

	return true
}

func (ctx *Context) ResultsForRequest(requestID string) []model.PingResult {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	byRequest, ok := ctx.State.results[requestID]
	if !ok {
		return nil
	}

	items := make([]model.PingResult, 0, len(byRequest))
	for _, result := range byRequest {
		items = append(items, result)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ObservedAt.Equal(items[j].ObservedAt) {
			return items[i].Executor.NodeID < items[j].Executor.NodeID
		}
		return items[i].ObservedAt.Before(items[j].ObservedAt)
	})

	return items
}

func (ctx *Context) AllResults() []model.PingResult {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	items := make([]model.PingResult, 0)
	for _, byRequest := range ctx.State.results {
		for _, result := range byRequest {
			items = append(items, result)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ObservedAt.Equal(items[j].ObservedAt) {
			return items[i].Key() < items[j].Key()
		}
		return items[i].ObservedAt.Before(items[j].ObservedAt)
	})

	return items
}

func (ctx *Context) NodesSnapshot() []model.ConnectedNode {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	items := make([]model.ConnectedNode, 0, len(ctx.State.nodes))
	for _, info := range ctx.State.nodes {
		items = append(items, info)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].LastContact.Equal(items[j].LastContact) {
			return items[i].NodeID < items[j].NodeID
		}
		return items[i].LastContact.After(items[j].LastContact)
	})

	return items
}

func (ctx *Context) Counts() (int, int, int) {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	resultsCount := 0
	for _, byRequest := range ctx.State.results {
		resultsCount += len(byRequest)
	}

	return len(ctx.State.nodes), len(ctx.State.requests), resultsCount
}
