package node

// orchestration.go coordinates node lifecycle tasks: polling for work,
// relaying results upstream, and executing controller interactions.

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bearpro/distributed-ping/model"
)

func (ctx *Context) Start(runCtx context.Context) {
	if ctx.Config.UpstreamURL == "" {
		return
	}

	go ctx.pollLoop(runCtx)
	if ctx.controller != nil {
		go ctx.relayLoop(runCtx)
	}
}

func (ctx *Context) CreatePingRequest(requestCtx context.Context, target string, requestID string) (model.PingRequest, error) {
	if ctx.Config.UpstreamURL == "" {
		if ctx.controller == nil {
			return model.PingRequest{}, fmt.Errorf("controller is not configured")
		}

		req := ctx.controller.CreateRequest(target, ctx.Config.Identity.NodeID, time.Now().UTC(), requestID)
		ctx.trackSubmitted(req)
		go ctx.processRequest(req)

		return req, nil
	}

	body := struct {
		Target       string `json:"target"`
		SourceNodeID string `json:"source_node_id"`
	}{
		Target:       target,
		SourceNodeID: ctx.Config.Identity.NodeID,
	}

	var response struct {
		Request model.PingRequest `json:"request"`
	}
	if err := ctx.doJSON(requestCtx, http.MethodPost, "/api/controller/ping-requests", body, &response); err != nil {
		return model.PingRequest{}, err
	}

	req := response.Request
	ctx.trackSubmitted(req)
	if ctx.controller != nil {
		ctx.controller.StoreRequest(req)
	}
	go ctx.processRequest(req)

	return req, nil
}

func (ctx *Context) ProcessRequest(req model.PingRequest) {
	go ctx.processRequest(req)
}

func (ctx *Context) RefreshRequestResults(requestCtx context.Context, requestID string) ([]model.PingResult, error) {
	if ctx.controller != nil {
		if _, ok := ctx.controller.RequestByID(requestID); ok {
			results := ctx.controller.ResultsForRequest(requestID)
			ctx.updateCachedResults(requestID, results)
			return results, nil
		}
	}

	if ctx.Config.UpstreamURL == "" {
		return nil, fmt.Errorf("request %s is unknown", requestID)
	}

	var response struct {
		Results []model.PingResult `json:"results"`
	}
	path := fmt.Sprintf("/api/controller/ping-requests/%s/results", url.PathEscape(requestID))
	if err := ctx.doJSON(requestCtx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	ctx.updateCachedResults(requestID, response.Results)
	return response.Results, nil
}

func (ctx *Context) pollLoop(runCtx context.Context) {
	ticker := time.NewTicker(ctx.Config.PollInterval)
	defer ticker.Stop()

	for {
		if err := ctx.pollOnce(runCtx); err != nil {
			ctx.logger.Printf("poll failed: %v", err)
		}

		select {
		case <-runCtx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (ctx *Context) relayLoop(runCtx context.Context) {
	ticker := time.NewTicker(ctx.Config.PollInterval)
	defer ticker.Stop()

	for {
		if err := ctx.relayPendingResults(runCtx); err != nil {
			ctx.logger.Printf("relay failed: %v", err)
		}

		select {
		case <-runCtx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (ctx *Context) pollOnce(runCtx context.Context) error {
	if !ctx.hasKnownOffset() {
		if err := ctx.fetchCurrentOffset(runCtx); err != nil {
			return err
		}
	}

	offset := ctx.CurrentOffset()
	var response struct {
		Requests      []model.PingRequest `json:"requests"`
		NextOffset    int                 `json:"next_offset"`
		CurrentOffset int                 `json:"current_offset"`
	}
	path := fmt.Sprintf("/api/controller/ping-requests?from_offset=%d&limit=%d", offset, 100)
	if err := ctx.doJSONWithHeaders(runCtx, http.MethodGet, path, nil, &response, ctx.pollHeaders()); err != nil {
		return err
	}

	ctx.setOffset(response.NextOffset)
	for _, req := range response.Requests {
		if ctx.controller != nil {
			ctx.controller.StoreRequest(req)
		}
		go ctx.processRequest(req)
	}

	return nil
}

func (ctx *Context) fetchCurrentOffset(runCtx context.Context) error {
	var response struct {
		CurrentOffset int `json:"current_offset"`
	}
	if err := ctx.doJSON(runCtx, http.MethodGet, "/api/controller", nil, &response); err != nil {
		return err
	}

	ctx.setOffset(response.CurrentOffset)
	return nil
}

func (ctx *Context) relayPendingResults(runCtx context.Context) error {
	results := ctx.controller.AllResults()
	for _, result := range results {
		if ctx.resultRelayed(result.Key()) {
			continue
		}

		toRelay := result
		toRelay.ReporterNodeID = ctx.Config.Identity.NodeID
		if err := ctx.doJSON(runCtx, http.MethodPost, "/api/controller/ping-results", toRelay, nil); err != nil {
			return err
		}
		ctx.markRelayed(result.Key())
	}

	return nil
}

func (ctx *Context) processRequest(req model.PingRequest) {
	if !ctx.markExecuted(req.ID) {
		return
	}

	result := executePing(ctx.Config.Identity.NodeIdentity, req)
	ctx.appendTrackedResult(result)

	if ctx.controller != nil {
		ctx.controller.StoreResult(result)
		return
	}

	if ctx.Config.UpstreamURL == "" {
		return
	}

	requestCtx, cancel := context.WithTimeout(context.Background(), ctx.Config.HTTPTimeout)
	defer cancel()

	if err := ctx.doJSON(requestCtx, http.MethodPost, "/api/controller/ping-results", result, nil); err != nil {
		ctx.logger.Printf("submit result failed for %s: %v", req.ID, err)
	}
}
