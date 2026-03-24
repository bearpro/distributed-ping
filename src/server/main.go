package main

import (
	"os"
	"time"

	"github.com/bearpro/distributed-ping/node"

	"github.com/gin-gonic/gin"
)

type Config struct {
	IsController bool
}

func loadConfig() Config {
	return Config{
		IsController: os.Getenv("IS_CONTROLLER") == "1",
	}
}

type ConnectedNodeInfo struct {
	LastContact  time.Time
	Location     string
	Organization string
}

type PingRequest struct {
	IP       string
	PostedAt time.Time
}

type ControllerState struct {
	ConnectedNodes []ConnectedNodeInfo
	PingRequests   []PingRequest
}

type ApplicationContext struct {
	Config          Config
	ControllerState *ControllerState
	NodeInfo        node.NodeInfo
}

func mkAppCtx(
	cfg Config,
	node_info node.NodeInfo,
) ApplicationContext {
	return ApplicationContext{
		Config:   cfg,
		NodeInfo: node_info,
	}
}

func main() {
	cfg := loadConfig()
	node_info, err := node.FetchNodeInfo()
	if err != nil {
		os.Exit(1)
	}

	ctx := mkAppCtx(cfg, *node_info)

	router := gin.Default()

	router.GET("/api/node", func(c *gin.Context) {
		c.JSON(200, ctx.NodeInfo)
	})

	router.GET("/api/controller", func(c *gin.Context) {
		if ctx.Config.IsController {
			response := gin.H{
				"status":         0,
				"message":        "ok",
				"current_offset": len(ctx.ControllerState.PingRequests),
			}
			c.JSON(200, response)
		} else {
			c.JSON(400, gin.H{"status": 1, "message": "Not a controller"})
		}
	})

	router.Run("localhost:8081")
}
