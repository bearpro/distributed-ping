package node

// ping.go is responsible for executing local ping commands and translating
// command output into transport-level ping results.

import (
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bearpro/distributed-ping/model"
)

var pingLatencyPattern = regexp.MustCompile(`time=([0-9.]+)`)

func executePing(identity model.NodeIdentity, req model.PingRequest) model.PingResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", req.Target)
	output, err := cmd.CombinedOutput()

	result := model.PingResult{
		RequestID:      req.ID,
		Executor:       identity,
		ReporterNodeID: identity.NodeID,
		ObservedAt:     time.Now().UTC(),
	}

	out := string(output)
	if err != nil {
		result.Error = summarizeCommandError(out, err)
		return result
	}

	result.Success = true
	result.LatencyMs = extractLatency(out)
	return result
}

func summarizeCommandError(output string, err error) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return err.Error()
	}
	if len(trimmed) > 240 {
		trimmed = trimmed[:240]
	}
	return trimmed
}

func extractLatency(output string) float64 {
	matches := pingLatencyPattern.FindStringSubmatch(output)
	if len(matches) != 2 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	return value
}
