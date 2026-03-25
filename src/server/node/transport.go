package node

// transport.go implements outbound HTTP communication from a node to its
// upstream controller.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (ctx *Context) doJSON(requestCtx context.Context, method string, path string, reqBody any, out any) error {
	return ctx.doJSONWithHeaders(requestCtx, method, path, reqBody, out, nil)
}

func (ctx *Context) doJSONWithHeaders(requestCtx context.Context, method string, path string, reqBody any, out any, headers map[string]string) error {
	endpoint := strings.TrimRight(ctx.Config.UpstreamURL, "/") + path

	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(requestCtx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := ctx.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("upstream returned %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}

	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (ctx *Context) pollHeaders() map[string]string {
	return map[string]string{
		"X-Node-ID":           ctx.Config.Identity.NodeID,
		"X-Node-Role":         string(ctx.Config.Identity.Role),
		"X-Node-Organization": ctx.Config.Identity.Organization,
		"X-Node-Lat":          strconv.FormatFloat(ctx.Config.Identity.Location.Lat, 'f', 6, 64),
		"X-Node-Lon":          strconv.FormatFloat(ctx.Config.Identity.Location.Lon, 'f', 6, 64),
	}
}
