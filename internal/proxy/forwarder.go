package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ForwardResult holds the response from a real HTTP microservice.
type ForwardResult struct {
	StatusCode int
	Body       map[string]interface{}
	RawBody    string
	ErrorMsg   string
	DurationMs float64
}

// HTTPForwarder manages real reverse-proxy HTTP dispatches to live microservices.
type HTTPForwarder struct {
	client *http.Client
}

// NewHTTPForwarder creates a production-grade HTTP forwarder with resilient timeouts.
func NewHTTPForwarder() *HTTPForwarder {
	return &HTTPForwarder{
		client: &http.Client{
			Timeout: 12 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ForwardRequest dispatches an HTTP POST request with the given payload to a real target URL.
// It returns the HTTP status code, parsed JSON response map, and error trace if status >= 400.
func (f *HTTPForwarder) ForwardRequest(
	ctx context.Context,
	targetURL string,
	payload map[string]interface{},
	headers map[string]string,
) *ForwardResult {
	start := time.Now()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		dur := float64(time.Since(start).Microseconds()) / 1000.0
		return &ForwardResult{
			StatusCode: http.StatusBadRequest,
			ErrorMsg:   fmt.Sprintf("Failed to marshal request payload: %v", err),
			DurationMs: dur,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		dur := float64(time.Since(start).Microseconds()) / 1000.0
		return &ForwardResult{
			StatusCode: http.StatusBadGateway,
			ErrorMsg:   fmt.Sprintf("Invalid upstream URL '%s': %v", targetURL, err),
			DurationMs: dur,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NeuroProxy-Gateway/2.0 (Google Cloud Run)")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := f.client.Do(req)
	dur := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return &ForwardResult{
			StatusCode: http.StatusBadGateway,
			Body: map[string]interface{}{
				"status": "UPSTREAM_UNREACHABLE",
				"error":  err.Error(),
				"target": targetURL,
			},
			ErrorMsg:   fmt.Sprintf("Upstream connection failure to %s: %v", targetURL, err),
			DurationMs: dur,
		}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	rawStr := string(respBytes)

	var respMap map[string]interface{}
	_ = json.Unmarshal(respBytes, &respMap)
	if respMap == nil {
		respMap = map[string]interface{}{
			"raw_response": rawStr,
			"status_code":  resp.StatusCode,
		}
	}

	var errMsg string
	if resp.StatusCode >= 400 {
		if e, ok := respMap["error"].(string); ok && e != "" {
			errMsg = e
		} else if m, ok := respMap["message"].(string); ok && m != "" {
			errMsg = m
		} else if d, ok := respMap["details"].([]interface{}); ok && len(d) > 0 {
			errMsg = fmt.Sprintf("%v", d)
		} else {
			errMsg = fmt.Sprintf("HTTP %d from %s: %s", resp.StatusCode, targetURL, rawStr)
		}
	}

	return &ForwardResult{
		StatusCode: resp.StatusCode,
		Body:       respMap,
		RawBody:    rawStr,
		ErrorMsg:   errMsg,
		DurationMs: dur,
	}
}
