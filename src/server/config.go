package main

// config.go owns environment-driven process configuration and keeps parsing
// rules out of the HTTP and runtime code paths.

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	ListenAddr     string
	IsController   bool
	ControllerURL  string
	NodeID         string
	PollInterval   time.Duration
	StaleNodeAfter time.Duration
	HTTPTimeout    time.Duration
}

func loadConfig() Config {
	return Config{
		ListenAddr:     getenvDefault("LISTEN_ADDR", ":8081"),
		IsController:   os.Getenv("IS_CONTROLLER") == "1",
		ControllerURL:  strings.TrimRight(os.Getenv("CONTROLLER_URL"), "/"),
		NodeID:         getenvDefault("NODE_ID", randomID("node")),
		PollInterval:   parseDurationEnv("POLL_INTERVAL", 10*time.Second),
		StaleNodeAfter: parseDurationEnv("STALE_NODE_AFTER", 2*time.Minute),
		HTTPTimeout:    parseDurationEnv("HTTP_TIMEOUT", 5*time.Second),
	}
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}

func getenvDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
