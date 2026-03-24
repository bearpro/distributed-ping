package main

import (
	"os"
	"time"

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
}

func mkAppCtx(cfg Config) ApplicationContext {
	return ApplicationContext{
		Config: cfg,
	}
}

func main() {
	cfg := loadConfig()
	ctx := mkAppCtx(cfg)

	router := gin.Default()
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
