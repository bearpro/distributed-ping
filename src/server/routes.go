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

func registerRoutes(router *gin.Engine, app Application) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/api/node", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"node":                app.Node.Config.Identity,
			"is_controller":       app.Config.IsController,
			"upstream_controller": app.Config.ControllerURL,
			"current_offset":      app.Node.CurrentOffset(),
			"submitted_requests":  len(app.Node.SubmittedRequests()),
		})
	})

	router.GET("/api/node/ping-requests", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"requests": app.Node.SubmittedRequests(),
		})
	})

	router.POST("/api/node/ping-requests", func(c *gin.Context) {
		var body struct {
			Target string `json:"target" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req, err := app.Node.CreatePingRequest(c.Request.Context(), body.Target, randomID("req"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"request": req})
	})

	router.GET("/api/node/ping-requests/:id/results", func(c *gin.Context) {
		results, err := app.Node.RefreshRequestResults(c.Request.Context(), c.Param("id"))
		if err != nil {
			status := http.StatusBadGateway
			if strings.Contains(err.Error(), "unknown") {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	})

	router.GET("/api/controller", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
			return
		}

		nodesCount, requestsCount, resultsCount := app.Controller.Counts()
		c.JSON(http.StatusOK, gin.H{
			"status":          "ok",
			"current_offset":  app.Controller.CurrentOffset(),
			"connected_nodes": nodesCount,
			"ping_requests":   requestsCount,
			"ping_results":    resultsCount,
		})
	})

	router.GET("/api/controller/nodes", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"nodes": app.Controller.NodesSnapshot()})
	})

	router.GET("/api/controller/ping-requests", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
			return
		}

		if identity, ok := identityFromHeaders(c); ok {
			app.Controller.RegisterNode(identity, time.Now().UTC())
		}

		fromOffset, _ := strconv.Atoi(c.DefaultQuery("from_offset", "0"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		requests, nextOffset := app.Controller.RequestsFrom(fromOffset, limit)
		c.JSON(http.StatusOK, gin.H{
			"requests":       requests,
			"next_offset":    nextOffset,
			"current_offset": app.Controller.CurrentOffset(),
		})
	})

	router.POST("/api/controller/ping-requests", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
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

		req := app.Controller.CreateRequest(body.Target, body.SourceNodeID, time.Now().UTC(), randomID("req"))
		app.Node.ProcessRequest(req)
		c.JSON(http.StatusCreated, gin.H{"request": req})
	})

	router.POST("/api/controller/ping-results", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
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

		if ok := app.Controller.StoreResult(result); !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
	})

	router.GET("/api/controller/ping-requests/:id/results", func(c *gin.Context) {
		if !ensureController(c, app.Controller) {
			return
		}

		requestID := c.Param("id")
		request, ok := app.Controller.RequestByID(requestID)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}

		results := app.Controller.ResultsForRequest(requestID)
		c.JSON(http.StatusOK, gin.H{
			"request":    request,
			"collecting": time.Now().UTC().Before(request.CollectUntil),
			"results":    results,
		})
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
