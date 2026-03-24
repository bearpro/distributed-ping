package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Location struct {
	Lat float64
	Lon float64
}

type NodeInfo struct {
	IP           string
	Organization string
	Location     Location
}

type ipInfoResponse struct {
	IP  string `json:"ip"`
	Org string `json:"org"`
	Loc string `json:"loc"` // "52.3068,4.9453"
}

func FetchNodeInfo() (*NodeInfo, error) {
	resp, err := http.Get("https://ipinfo.io/json")
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

	return &NodeInfo{
		IP:           raw.IP,
		Organization: raw.Org,
		Location: Location{
			Lat: lat,
			Lon: lon,
		},
	}, nil
}
