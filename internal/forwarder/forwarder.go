package forwarder

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

// ForwardResult contains the outcome of an external HTTP upstream proxy call.
type ForwardResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// HTTPForwarder dispatches outgoing HTTP requests with connection pooling and timeouts.
type HTTPForwarder struct {
	client *http.Client
}

// NewHTTPForwarder returns a high-throughput, pooled HTTP forwarder.
func NewHTTPForwarder() *HTTPForwarder {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &HTTPForwarder{
		client: &http.Client{
			Transport: transport,
			Timeout:   20 * time.Second,
		},
	}
}

// Dispatch executes an HTTP POST to the target URL with provided body and headers.
func (f *HTTPForwarder) Dispatch(ctx context.Context, targetURL string, headers http.Header, body []byte) (*ForwardResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ForwardResult{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
		Duration:   time.Since(start),
	}, nil
}
