package node

// context.go defines the node context and separates immutable configuration
// from mutable runtime state.

import (
	"log"
	"net/http"
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
