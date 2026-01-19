package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/janitarr/src/database"
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
