package node

// node.go owns only self-discovery of node metadata from an external IP info
// service. Runtime coordination lives in runtime.go.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bearpro/distributed-ping/model"
)

type ipInfoResponse struct {
	IP  string `json:"ip"`
	Org string `json:"org"`
	Loc string `json:"loc"` // "52.3068,4.9453"
}

func FetchNodeInfo() (*model.NodeInfo, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var raw ipInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	parts := strings.Split(raw.Loc, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid loc format: %q", raw.Loc)
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("parse latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("parse longitude: %w", err)
	}

	return &model.NodeInfo{
		NodeIdentity: model.NodeIdentity{
			Organization: raw.Org,
			Location: model.Location{
				Lat: lat,
				Lon: lon,
			},
		},
		PublicIP: raw.IP,
	}, nil
}
