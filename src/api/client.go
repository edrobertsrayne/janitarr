package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 15 * time.Second

	// APIPrefix is the API version path prefix for Radarr/Sonarr.
	APIPrefix = "/api/v3"
)

// DebugLogger is an interface for debug logging to avoid circular dependencies.
type DebugLogger interface {
	Debug(msg string, keyvals ...interface{})
}

// Client is a base HTTP client for Radarr/Sonarr APIs.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     DebugLogger
	serverName string // For logging context
}

// NormalizeURL ensures a URL has a protocol and removes trailing slashes.
func NormalizeURL(url string) string {
	normalized := strings.TrimSpace(url)

	// Add protocol if missing
	if !regexp.MustCompile(`^https?://`).MatchString(normalized) {
		normalized = "http://" + normalized
	}

	// Remove trailing slashes
	normalized = strings.TrimRight(normalized, "/")

	return normalized
}

// NewClient creates a new API client with default timeout.
func NewClient(url, apiKey string) *Client {
	return NewClientWithTimeout(url, apiKey, DefaultTimeout)
}

// NewClientWithTimeout creates a new API client with a custom timeout.
func NewClientWithTimeout(url, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: NormalizeURL(url),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs a GET request to the specified endpoint.
func (c *Client) Get(ctx context.Context, endpoint string, result any) error {
	return c.request(ctx, http.MethodGet, endpoint, nil, result)
}

// Post performs a POST request to the specified endpoint.
func (c *Client) Post(ctx context.Context, endpoint string, body, result any) error {
	return c.request(ctx, http.MethodPost, endpoint, body, result)
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// WithLogger attaches a logger to the client for debug logging.
func (c *Client) WithLogger(logger DebugLogger, serverName string) *Client {
	c.logger = logger
	c.serverName = serverName
	return c
}

// request performs an HTTP request to the API.
func (c *Client) request(ctx context.Context, method, endpoint string, body, result any) error {
	url := c.baseURL + APIPrefix + endpoint
	start := time.Now()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			return fmt.Errorf("request timeout: %w", err)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	// Log API request at debug level (without API key)
	if c.logger != nil {
		logFields := []interface{}{
			"endpoint", endpoint,
			"status", resp.StatusCode,
			"duration", duration.String(),
		}
		if c.serverName != "" {
			logFields = append([]interface{}{"server", c.serverName}, logFields...)
		}
		c.logger.Debug("API request", logFields...)
	}

	if err := c.checkStatusCode(resp.StatusCode); err != nil {
		return err
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// checkStatusCode returns an error for non-success status codes.
func (c *Client) checkStatusCode(code int) error {
	switch code {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: invalid API key")
	case http.StatusNotFound:
		return fmt.Errorf("not found: check server URL")
	default:
		if code >= 400 {
			return fmt.Errorf("server error: status %d", code)
		}
		return nil
	}
}
