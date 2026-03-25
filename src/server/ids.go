package main

// ids.go owns local identifier generation for nodes and ping requests.

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

func randomID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "_" + strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
