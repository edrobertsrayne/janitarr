package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edrobertsrayne/janitarr/src/database"
)

func TestGetConfig(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	req := httptest.NewRequest("GET", "/api/config", nil)
	rr := httptest.NewRecorder()

	handlers.GetConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v\nBody: %s", err, rr.Body.String())
	}

	// Extract the config from the data field
	configBytes, _ := json.Marshal(resp.Data)
	var config database.AppConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Should return default config on first access
	expected := database.DefaultAppConfig()
	if config.Schedule.IntervalHours != expected.Schedule.IntervalHours {
		t.Errorf("expected interval %d, got %d", expected.Schedule.IntervalHours, config.Schedule.IntervalHours)
	}
	if config.Schedule.Enabled != expected.Schedule.Enabled {
		t.Errorf("expected enabled %v, got %v", expected.Schedule.Enabled, config.Schedule.Enabled)
	}
}

func TestPatchConfig_Success(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	updates := map[string]any{
		"schedule.intervalHours":    12.0,
		"schedule.enabled":          false,
		"limits.missingMoviesLimit": 50.0,
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.PatchConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify changes were persisted
	config := db.GetAppConfig()
	if config.Schedule.IntervalHours != 12 {
		t.Errorf("expected interval 12, got %d", config.Schedule.IntervalHours)
	}
	if config.Schedule.Enabled != false {
		t.Errorf("expected enabled false, got %v", config.Schedule.Enabled)
	}
	if config.SearchLimits.MissingMoviesLimit != 50 {
		t.Errorf("expected missing movies limit 50, got %d", config.SearchLimits.MissingMoviesLimit)
	}
}

func TestPatchConfig_InvalidJSON(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	req := httptest.NewRequest("PATCH", "/api/config", strings.NewReader(`{invalid json`))
	rr := httptest.NewRecorder()

	handlers.PatchConfig(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Error != "Invalid request payload" {
		t.Errorf("unexpected error message: %s", resp.Error)
	}
}

func TestPatchConfig_UnknownKey(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	updates := map[string]any{
		"unknown.key": "value",
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.PatchConfig(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if !strings.Contains(resp.Error, "Unknown configuration key") {
		t.Errorf("expected unknown key error, got: %s", resp.Error)
	}
}

func TestPatchConfig_InvalidValueType(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	updates := map[string]any{
		"schedule.intervalHours": "not-an-int",
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.PatchConfig(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if !strings.Contains(resp.Error, "Invalid value type") {
		t.Errorf("expected invalid value type error, got: %s", resp.Error)
	}
}

func TestResetConfig_Success(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	// First, modify the config
	config := db.GetAppConfig()
	config.Schedule.IntervalHours = 99
	db.SetAppConfig(config)

	// Now reset it
	req := httptest.NewRequest("PUT", "/api/config/reset", nil)
	rr := httptest.NewRecorder()

	handlers.ResetConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify it was reset to defaults
	config = db.GetAppConfig()
	expected := database.DefaultAppConfig()
	if config.Schedule.IntervalHours != expected.Schedule.IntervalHours {
		t.Errorf("expected default interval %d, got %d", expected.Schedule.IntervalHours, config.Schedule.IntervalHours)
	}
}

func TestPostConfig_HighLimitWarning(t *testing.T) {
	db := testDB(t)
	handlers := NewConfigHandlers(db)

	tests := []struct {
		name        string
		formData    string
		wantWarning bool
	}{
		{
			name:        "limit 100 - no warning",
			formData:    "limits.missing.movies=100",
			wantWarning: false,
		},
		{
			name:        "limit 101 - warning",
			formData:    "limits.missing.movies=101",
			wantWarning: true,
		},
		{
			name:        "limit 500 - warning",
			formData:    "limits.cutoff.episodes=500",
			wantWarning: true,
		},
		{
			name:        "all limits <= 100 - no warning",
			formData:    "limits.missing.movies=50&limits.missing.episodes=75&limits.cutoff.movies=100&limits.cutoff.episodes=25",
			wantWarning: false,
		},
		{
			name:        "one limit > 100 - warning",
			formData:    "limits.missing.movies=50&limits.missing.episodes=150&limits.cutoff.movies=100&limits.cutoff.episodes=25",
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/config", strings.NewReader(tt.formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handlers.PostConfig(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
			}

			var resp SuccessResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v\nBody: %s", err, rr.Body.String())
			}

			// Check if warning is present in Data field
			hasWarning := false
			if data, ok := resp.Data.(map[string]any); ok {
				if warning, exists := data["warning"]; exists && warning != "" {
					hasWarning = true
				}
			}

			if hasWarning != tt.wantWarning {
				t.Errorf("expected warning=%v, got warning=%v\nResponse: %+v", tt.wantWarning, hasWarning, resp)
			}
		})
	}
}
