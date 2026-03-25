package node

// context.go defines the node context and separates immutable configuration
// from mutable runtime state.

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/bearpro/distributed-ping/controller"
	"github.com/bearpro/distributed-ping/model"
)

type Configuration struct {
	Identity     model.NodeInfo
	UpstreamURL  string
	PollInterval time.Duration
	HTTPTimeout  time.Duration
}

type State struct {
	mu               sync.RWMutex
	upstreamOffset   int
	offsetKnown      bool
	submitted        map[string]model.SubmittedPingRequest
	executedRequests map[string]time.Time
	relayedResults   map[string]time.Time
}

type Context struct {
	Config     Configuration
	State      *State
	controller *controller.Context
	httpClient *http.Client
	logger     *log.Logger
}

type Snapshot struct {
	Config Configuration `json:"config"`
	State  StateSnapshot `json:"state"`
}

type StateSnapshot struct {
	CurrentOffset      int                          `json:"current_offset"`
	OffsetKnown        bool                         `json:"offset_known"`
	SubmittedRequests  []model.SubmittedPingRequest `json:"submitted_requests"`
	ExecutedRequestIDs []string                     `json:"executed_request_ids"`
	RelayedResultKeys  []string                     `json:"relayed_result_keys"`
}

func NewContext(cfg Configuration, controllerCtx *controller.Context, logger *log.Logger) *Context {
	httpTimeout := cfg.HTTPTimeout
	if httpTimeout <= 0 {
		httpTimeout = 5 * time.Second
	}

	return &Context{
		Config: cfg,
		State: &State{
			submitted:        make(map[string]model.SubmittedPingRequest),
			executedRequests: make(map[string]time.Time),
			relayedResults:   make(map[string]time.Time),
		},
		controller: controllerCtx,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		logger: logger,
	}
}

func (ctx *Context) Snapshot() Snapshot {
	ctx.State.mu.RLock()
	submitted := make([]model.SubmittedPingRequest, 0, len(ctx.State.submitted))
	for _, request := range ctx.State.submitted {
		submitted = append(submitted, request)
	}

	executedRequestIDs := make([]string, 0, len(ctx.State.executedRequests))
	for requestID := range ctx.State.executedRequests {
		executedRequestIDs = append(executedRequestIDs, requestID)
	}

	relayedResultKeys := make([]string, 0, len(ctx.State.relayedResults))
	for key := range ctx.State.relayedResults {
		relayedResultKeys = append(relayedResultKeys, key)
	}

	currentOffset := ctx.State.upstreamOffset
	offsetKnown := ctx.State.offsetKnown
	ctx.State.mu.RUnlock()

	sortSubmittedRequests(submitted)
	sort.Strings(executedRequestIDs)
	sort.Strings(relayedResultKeys)

	return Snapshot{
		Config: ctx.Config,
		State: StateSnapshot{
			CurrentOffset:      currentOffset,
			OffsetKnown:        offsetKnown,
			SubmittedRequests:  submitted,
			ExecutedRequestIDs: executedRequestIDs,
			RelayedResultKeys:  relayedResultKeys,
		},
	}
}

func (ctx *Context) CurrentOffset() int {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	return ctx.State.upstreamOffset
}

func (ctx *Context) SubmittedRequests() []model.SubmittedPingRequest {
	ctx.State.mu.RLock()
	defer ctx.State.mu.RUnlock()

	items := make([]model.SubmittedPingRequest, 0, len(ctx.State.submitted))
	for _, request := range ctx.State.submitted {
		items = append(items, request)
	}

	sortSubmittedRequests(items)
	return items
}
