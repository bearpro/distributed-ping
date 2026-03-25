package main

// app.go wires process configuration, node metadata, controller state, and
// runtime workers into a single application object used by the HTTP layer.

import (
	"context"
	"log"
	"time"

	"github.com/bearpro/distributed-ping/controller"
	"github.com/bearpro/distributed-ping/model"
	"github.com/bearpro/distributed-ping/node"
)

type Application struct {
	Config     Config
	Controller *controller.Context
	Node       *node.Context
}

type ApplicationSnapshot struct {
	Config     Config               `json:"config"`
	Node       node.Snapshot        `json:"node"`
	Controller *controller.Snapshot `json:"controller,omitempty"`
}

func newApplication(cfg Config, logger *log.Logger) Application {
	nodeInfo := loadNodeInfo(cfg, logger)

	var controllerCtx *controller.Context
	if cfg.IsController {
		controllerCtx = controller.NewContext(controller.Configuration{
			StaleNodeAfter: cfg.StaleNodeAfter,
		})
	}

	nodeCtx := node.NewContext(node.Configuration{
		Identity:     nodeInfo,
		UpstreamURL:  cfg.ControllerURL,
		PollInterval: cfg.PollInterval,
		HTTPTimeout:  cfg.HTTPTimeout,
	}, controllerCtx, logger)

	return Application{
		Config:     cfg,
		Controller: controllerCtx,
		Node:       nodeCtx,
	}
}

func (app Application) Snapshot() ApplicationSnapshot {
	snapshot := ApplicationSnapshot{
		Config: app.Config,
		Node:   app.Node.Snapshot(),
	}
	if app.Controller != nil {
		controllerSnapshot := app.Controller.Snapshot()
		snapshot.Controller = &controllerSnapshot
	}
	return snapshot
}

func (app Application) Start(ctx context.Context) {
	app.Node.Start(ctx)
	if app.Controller != nil {
		go cleanupLoop(ctx, app.Controller, cleanupInterval(app.Config.StaleNodeAfter))
	}
}

func loadNodeInfo(cfg Config, logger *log.Logger) model.NodeInfo {
	info := model.NodeInfo{}
	fetched, err := node.FetchNodeInfo()
	if err != nil {
		logger.Printf("node discovery failed: %v", err)
	} else {
		info = *fetched
	}

	info.NodeIdentity.NodeID = cfg.NodeID
	info.NodeIdentity.Role = model.DetectRole(cfg.IsController, cfg.ControllerURL != "")
	return info
}

func cleanupLoop(ctx context.Context, controllerCtx *controller.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			controllerCtx.CleanupStaleNodes(now.UTC())
		}
	}
}

func cleanupInterval(staleAfter time.Duration) time.Duration {
	if staleAfter <= 30*time.Second {
		return 15 * time.Second
	}
	return staleAfter / 2
}
