package main

import "github.com/bearpro/distributed-ping/model"

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --dir . --generalInfo main.go --output ./docs --parseDependency --parseInternal

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"request not found"`
}

type NodeSummaryResponse struct {
	Node               model.NodeInfo `json:"node"`
	IsController       bool           `json:"is_controller"`
	UpstreamController string         `json:"upstream_controller,omitempty" example:"http://controller:8081"`
	CurrentOffset      int            `json:"current_offset" example:"12"`
	SubmittedRequests  int            `json:"submitted_requests" example:"3"`
}

type NodePingRequestsResponse struct {
	Requests []model.SubmittedPingRequest `json:"requests"`
}

type CreateNodePingRequestBody struct {
	Target string `json:"target" binding:"required" example:"1.1.1.1"`
}

type PingRequestResponse struct {
	Request model.PingRequest `json:"request"`
}

type PingResultsResponse struct {
	Results []model.PingResult `json:"results"`
}

type ControllerSummaryResponse struct {
	Status         string `json:"status" example:"ok"`
	CurrentOffset  int    `json:"current_offset" example:"12"`
	ConnectedNodes int    `json:"connected_nodes" example:"4"`
	PingRequests   int    `json:"ping_requests" example:"8"`
	PingResults    int    `json:"ping_results" example:"16"`
}

type ControllerNodesResponse struct {
	Nodes []model.ConnectedNode `json:"nodes"`
}

type ControllerPingRequestsResponse struct {
	Requests      []model.PingRequest `json:"requests"`
	NextOffset    int                 `json:"next_offset" example:"100"`
	CurrentOffset int                 `json:"current_offset" example:"140"`
}

type CreateControllerPingRequestBody struct {
	Target       string `json:"target" binding:"required" example:"example.com"`
	SourceNodeID string `json:"source_node_id,omitempty" example:"node-west-1"`
}

type AcceptedResponse struct {
	Status string `json:"status" example:"accepted"`
}

type ControllerPingRequestResultsResponse struct {
	Request    model.PingRequest  `json:"request"`
	Collecting bool               `json:"collecting"`
	Results    []model.PingResult `json:"results"`
}
