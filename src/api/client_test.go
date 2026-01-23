package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "http://localhost:7878",
			expected: "http://localhost:7878",
		},
		{
			name:     "with trailing slash",
			input:    "http://localhost:7878/",
			expected: "http://localhost:7878",
		},
		{
			name:     "with multiple trailing slashes",
			input:    "http://localhost:7878///",
			expected: "http://localhost:7878",
		},
		{
			name:     "missing protocol",
			input:    "localhost:7878",
			expected: "http://localhost:7878",
		},
		{
			name:     "https protocol",
			input:    "https://localhost:7878",
			expected: "https://localhost:7878",
		},
		{
			name:     "with whitespace",
			input:    "  http://localhost:7878  ",
			expected: "http://localhost:7878",
		},
		{
			name:     "with path",
			input:    "http://localhost:7878/radarr",
			expected: "http://localhost:7878/radarr",
		},
		{
			name:     "with path and trailing slash",
			input:    "http://localhost:7878/radarr/",
			expected: "http://localhost:7878/radarr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:7878", "testapikey")

	if client.baseURL != "http://localhost:7878" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "http://localhost:7878")
	}
	if client.apiKey != "testapikey" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "testapikey")
	}
	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, DefaultTimeout)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := NewClientWithTimeout("http://localhost:7878", "testapikey", timeout)

	if client.httpClient.Timeout != timeout {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, timeout)
	}
}

func TestClientGet_Success(t *testing.T) {
	expected := SystemStatus{AppName: "Radarr", Version: "5.0.0"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "testapikey" {
			t.Errorf("expected api key header")
		}
		if !strings.HasSuffix(r.URL.Path, "/api/v3/system/status") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AppName != expected.AppName {
		t.Errorf("AppName = %q, want %q", result.AppName, expected.AppName)
	}
	if result.Version != expected.Version {
		t.Errorf("Version = %q, want %q", result.Version, expected.Version)
	}
}

func TestClientPost_Success(t *testing.T) {
	expected := CommandResponse{ID: 1, Name: "MoviesSearch", Status: "started"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content type")
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "MoviesSearch" {
			t.Errorf("expected name=MoviesSearch, got %v", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result CommandResponse
	err := client.Post(context.Background(), "/command", map[string]any{"name": "MoviesSearch", "movieIds": []int{1, 2, 3}}, &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("ID = %d, want %d", result.ID, expected.ID)
	}
}

func TestClientGet_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "badkey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error should mention unauthorized: %v", err)
	}
}

func TestClientGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/nonexistent", &result)

	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestClientGet_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestClientGet_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithTimeout(server.URL, "testapikey", 50*time.Millisecond)
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("error should mention timeout: %v", err)
	}
}

func TestClientGet_ConnectionRefused(t *testing.T) {
	// Connect to a port that's not listening
	client := NewClientWithTimeout("http://localhost:59999", "testapikey", 1*time.Second)
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestClientGet_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestClientGet_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(ctx, "/system/status", &result)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestClientGet_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected rate limit error")
	}

	// Check that it's a RateLimitError
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	// Check that Retry-After was parsed correctly
	expectedRetryAfter := 60 * time.Second
	if rateLimitErr.RetryAfter != expectedRetryAfter {
		t.Errorf("RetryAfter = %v, want %v", rateLimitErr.RetryAfter, expectedRetryAfter)
	}
}

func TestClientGet_TooManyRequests_DefaultRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Retry-After header
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL, "testapikey")
	var result SystemStatus
	err := client.Get(context.Background(), "/system/status", &result)

	if err == nil {
		t.Fatal("expected rate limit error")
	}

	// Check that it's a RateLimitError
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}

	// Check that default 30s was used
	expectedRetryAfter := 30 * time.Second
	if rateLimitErr.RetryAfter != expectedRetryAfter {
		t.Errorf("RetryAfter = %v, want %v (default)", rateLimitErr.RetryAfter, expectedRetryAfter)
	}
}
