package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
)

// MockDBConfig is a mock database for config-related operations.
type MockDBConfig struct {
	mock.Mock
}

func (m *MockDBConfig) GetAppConfig() database.AppConfig {
	args := m.Called()
	return args.Get(0).(database.AppConfig)
}

func (m *MockDBConfig) SetAppConfig(config database.AppConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func TestConfigHandlers(t *testing.T) {
	assert := assert.New(t)

	mockDB := new(MockDBConfig)
	handlers := NewConfigHandlers(nil) // DB will be mocked via functions

	// Override database.GetAppConfigFunc and database.SetAppConfigFunc for testing
	originalGetAppConfigFunc := database.GetAppConfigFunc
	defer func() { database.GetAppConfigFunc = originalGetAppConfigFunc }()
	originalSetAppConfigFunc := database.SetAppConfigFunc
	defer func() { database.SetAppConfigFunc = originalSetAppConfigFunc }()

	database.GetAppConfigFunc = func(db *database.DB) database.AppConfig {
		return mockDB.GetAppConfig()
	}
	database.SetAppConfigFunc = func(db *database.DB, config database.AppConfig) error {
		return mockDB.SetAppConfig(config)
	}

	t.Run("GetConfig - success", func(t *testing.T) {
		expectedConfig := database.DefaultAppConfig()
		mockDB.On("GetAppConfig").Return(expectedConfig).Once()

		req, _ := http.NewRequest("GET", "/api/config", nil)
		rr := httptest.NewRecorder()
		handlers.GetConfig(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var actualConfig database.AppConfig
		err := json.Unmarshal(rr.Body.Bytes(), &actualConfig)
		assert.NoError(err)
		assert.Equal(expectedConfig, actualConfig)
		mockDB.AssertExpectations(t)
	})

	t.Run("PatchConfig - success with multiple fields", func(t *testing.T) {
		initialConfig := database.DefaultAppConfig()
		updatedConfig := initialConfig
		updatedConfig.Schedule.IntervalHours = 12
		updatedConfig.Schedule.Enabled = false
		updatedConfig.SearchLimits.MissingMoviesLimit = 50

		patchPayload := map[string]any{
			"schedule.intervalHours":    12.0,
			"schedule.enabled":          false,
			"limits.missingMoviesLimit": 50.0,
		}
		body, _ := json.Marshal(patchPayload)

		mockDB.On("GetAppConfig").Return(initialConfig).Once()
		mockDB.On("SetAppConfig", updatedConfig).Return(nil).Once()

		req, _ := http.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handlers.PatchConfig(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(err)
		assert.Equal("Configuration updated successfully", resp.Message)
		mockDB.AssertExpectations(t)
	})

	t.Run("PatchConfig - invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/config", strings.NewReader(`{invalid json`))
		rr := httptest.NewRecorder()
		handlers.PatchConfig(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Invalid request payload", resp.Error)
	})

	t.Run("PatchConfig - unknown key", func(t *testing.T) {
		initialConfig := database.DefaultAppConfig()
		patchPayload := map[string]any{
			"unknown.key": "value",
		}
		body, _ := json.Marshal(patchPayload)

		mockDB.On("GetAppConfig").Return(initialConfig).Once()

		req, _ := http.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handlers.PatchConfig(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Unknown configuration key")
		mockDB.AssertExpectations(t)
	})

	t.Run("PatchConfig - invalid value type", func(t *testing.T) {
		initialConfig := database.DefaultAppConfig()
		patchPayload := map[string]any{
			"schedule.intervalHours": "not-an-int",
		}
		body, _ := json.Marshal(patchPayload)

		mockDB.On("GetAppConfig").Return(initialConfig).Once()

		req, _ := http.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handlers.PatchConfig(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Invalid value type for schedule.intervalHours")
		mockDB.AssertExpectations(t)
	})

	t.Run("PatchConfig - database error on SetAppConfig", func(t *testing.T) {
		initialConfig := database.DefaultAppConfig()
		patchPayload := map[string]any{
			"schedule.enabled": false,
		}
		body, _ := json.Marshal(patchPayload)

		mockDB.On("GetAppConfig").Return(initialConfig).Once()
		mockDB.On("SetAppConfig", mock.AnythingOfType("database.AppConfig")).Return(errors.New("db write error")).Once()

		req, _ := http.NewRequest("PATCH", "/api/config", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handlers.PatchConfig(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Failed to update configuration")
		mockDB.AssertExpectations(t)
	})

	t.Run("ResetConfig - success", func(t *testing.T) {
		defaultConfig := database.DefaultAppConfig()

		mockDB.On("SetAppConfig", defaultConfig).Return(nil).Once()

		req, _ := http.NewRequest("PUT", "/api/config/reset", nil)
		rr := httptest.NewRecorder()
		handlers.ResetConfig(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(err)
		assert.Equal("Configuration reset to defaults successfully", resp.Message)
		mockDB.AssertExpectations(t)
	})

	t.Run("ResetConfig - database error", func(t *testing.T) {
		defaultConfig := database.DefaultAppConfig()

		mockDB.On("SetAppConfig", defaultConfig).Return(errors.New("db reset error")).Once()

		req, _ := http.NewRequest("PUT", "/api/config/reset", nil)
		rr := httptest.NewRecorder()
		handlers.ResetConfig(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Failed to reset configuration")
		mockDB.AssertExpectations(t)
	})
}
