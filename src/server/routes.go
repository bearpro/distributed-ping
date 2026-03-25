package main

// routes.go contains the HTTP API surface for both node-facing and
// controller-facing endpoints. Its only responsibility is request handling.

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bearpro/distributed-ping/controller"
	"github.com/bearpro/distributed-ping/model"
	"github.com/gin-gonic/gin"
)

type routeHandlers struct {
	app Application
}

func registerRoutes(router *gin.Engine, app Application) {
	handlers := routeHandlers{app: app}

	router.GET("/healthz", handlers.healthz)
	router.GET("/api/application", handlers.getApplication)
	router.GET("/api/node", handlers.getNode)
	router.GET("/api/node/ping-requests", handlers.listNodePingRequests)
	router.POST("/api/node/ping-requests", handlers.createNodePingRequest)
	router.GET("/api/node/ping-requests/:id/results", handlers.getNodePingRequestResults)
	router.GET("/api/controller", handlers.getController)
	router.GET("/api/controller/nodes", handlers.listControllerNodes)
	router.GET("/api/controller/ping-requests", handlers.listControllerPingRequests)
	router.POST("/api/controller/ping-requests", handlers.createControllerPingRequest)
	router.POST("/api/controller/ping-results", handlers.createControllerPingResult)
	router.GET("/api/controller/ping-requests/:id/results", handlers.getControllerPingRequestResults)
}

func (h routeHandlers) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h routeHandlers) getApplication(c *gin.Context) {
	c.JSON(http.StatusOK, h.app.Snapshot())
}

func (h routeHandlers) getNode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"node":                h.app.Node.Config.Identity,
		"is_controller":       h.app.Config.IsController,
		"upstream_controller": h.app.Config.ControllerURL,
		"current_offset":      h.app.Node.CurrentOffset(),
		"submitted_requests":  len(h.app.Node.SubmittedRequests()),
	})
}

func (h routeHandlers) listNodePingRequests(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"requests": h.app.Node.SubmittedRequests(),
	})
}

func (h routeHandlers) createNodePingRequest(c *gin.Context) {
	var body struct {
		Target string `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := h.app.Node.CreatePingRequest(c.Request.Context(), body.Target, randomID("req"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"request": req})
}

func (h routeHandlers) getNodePingRequestResults(c *gin.Context) {
	results, err := h.app.Node.RefreshRequestResults(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := http.StatusBadGateway
		if strings.Contains(err.Error(), "unknown") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h routeHandlers) getController(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	nodesCount, requestsCount, resultsCount := h.app.Controller.Counts()
	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"current_offset":  h.app.Controller.CurrentOffset(),
		"connected_nodes": nodesCount,
		"ping_requests":   requestsCount,
		"ping_results":    resultsCount,
	})
}

func (h routeHandlers) listControllerNodes(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": h.app.Controller.NodesSnapshot()})
}

func (h routeHandlers) listControllerPingRequests(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	if identity, ok := identityFromHeaders(c); ok {
		h.app.Controller.RegisterNode(identity, time.Now().UTC())
	}

	fromOffset, _ := strconv.Atoi(c.DefaultQuery("from_offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	requests, nextOffset := h.app.Controller.RequestsFrom(fromOffset, limit)
	c.JSON(http.StatusOK, gin.H{
		"requests":       requests,
		"next_offset":    nextOffset,
		"current_offset": h.app.Controller.CurrentOffset(),
	})
}

func (h routeHandlers) createControllerPingRequest(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	var body struct {
		Target       string `json:"target" binding:"required"`
		SourceNodeID string `json:"source_node_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := h.app.Controller.CreateRequest(body.Target, body.SourceNodeID, time.Now().UTC(), randomID("req"))
	h.app.Node.ProcessRequest(req)
	c.JSON(http.StatusCreated, gin.H{"request": req})
}

func (h routeHandlers) createControllerPingResult(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	var result model.PingResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result.RequestID == "" || result.Executor.NodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id and executor.node_id are required"})
		return
	}
	if result.ObservedAt.IsZero() {
		result.ObservedAt = time.Now().UTC()
	}
	if result.ReporterNodeID == "" {
		result.ReporterNodeID = result.Executor.NodeID
	}

	if ok := h.app.Controller.StoreResult(result); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (h routeHandlers) getControllerPingRequestResults(c *gin.Context) {
	if !ensureController(c, h.app.Controller) {
		return
	}

	requestID := c.Param("id")
	request, ok := h.app.Controller.RequestByID(requestID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}

	results := h.app.Controller.ResultsForRequest(requestID)
	c.JSON(http.StatusOK, gin.H{
		"request":    request,
		"collecting": time.Now().UTC().Before(request.CollectUntil),
		"results":    results,
	})
}

func ensureController(c *gin.Context, controllerCtx *controller.Context) bool {
	if controllerCtx != nil {
		return true
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "node is not running as controller"})
	return false
}

func identityFromHeaders(c *gin.Context) (model.NodeIdentity, bool) {
	nodeID := strings.TrimSpace(c.GetHeader("X-Node-ID"))
	if nodeID == "" {
		return model.NodeIdentity{}, false
	}

	lat, _ := strconv.ParseFloat(strings.TrimSpace(c.GetHeader("X-Node-Lat")), 64)
	lon, _ := strconv.ParseFloat(strings.TrimSpace(c.GetHeader("X-Node-Lon")), 64)

	return model.NodeIdentity{
		NodeID:       nodeID,
		Organization: strings.TrimSpace(c.GetHeader("X-Node-Organization")),
		Role:         model.NodeRole(strings.TrimSpace(c.GetHeader("X-Node-Role"))),
		Location: model.Location{
			Lat: lat,
			Lon: lon,
		},
	}, true
}
