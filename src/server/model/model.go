package model

// model.go defines the transport and state models shared across the node,
// controller, and HTTP layers.

import "time"

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type NodeRole string

const (
	NodeRoleWorker     NodeRole = "worker"
	NodeRoleController NodeRole = "controller"
	NodeRoleHybrid     NodeRole = "hybrid"
)

func DetectRole(isController bool, hasUpstream bool) NodeRole {
	switch {
	case isController && hasUpstream:
		return NodeRoleHybrid
	case isController:
		return NodeRoleController
	default:
		return NodeRoleWorker
	}
}

type NodeIdentity struct {
	NodeID       string   `json:"node_id"`
	Organization string   `json:"organization,omitempty"`
	Location     Location `json:"location"`
	Role         NodeRole `json:"role"`
}

type NodeInfo struct {
	NodeIdentity
	PublicIP string `json:"public_ip,omitempty"`
}

type ConnectedNode struct {
	NodeIdentity
	LastContact time.Time `json:"last_contact"`
}

type PingRequest struct {
	ID           string    `json:"id"`
	Target       string    `json:"target"`
	CreatedAt    time.Time `json:"created_at"`
	CollectUntil time.Time `json:"collect_until"`
	SourceNodeID string    `json:"source_node_id,omitempty"`
}

type PingResult struct {
	RequestID      string       `json:"request_id"`
	Executor       NodeIdentity `json:"executor"`
	ReporterNodeID string       `json:"reporter_node_id,omitempty"`
	Success        bool         `json:"success"`
	LatencyMs      float64      `json:"latency_ms,omitempty"`
	Error          string       `json:"error,omitempty"`
	ObservedAt     time.Time    `json:"observed_at"`
}

func (r PingResult) Key() string {
	return r.RequestID + ":" + r.Executor.NodeID
}

type SubmittedPingRequest struct {
	Request           PingRequest  `json:"request"`
	LastResultsSyncAt time.Time    `json:"last_results_sync_at,omitempty"`
	CachedResults     []PingResult `json:"cached_results,omitempty"`
}
