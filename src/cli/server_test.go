package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

// MockServerManager for testing CLI commands
type MockServerManager struct {
	mock.Mock
}

func (m *MockServerManager) AddServer(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error) {
	args := m.Called(ctx, name, url, apiKey, serverType)
	return args.Get(0).(*services.ServerInfo), args.Error(1)
}

func (m *MockServerManager) UpdateServer(ctx context.Context, id string, updates services.ServerUpdate) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockServerManager) RemoveServer(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockServerManager) TestConnection(ctx context.Context, id string) (*services.ConnectionResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*services.ConnectionResult), args.Error(1)
}

func (m *MockServerManager) ListServers() ([]services.ServerInfo, error) {
	args := m.Called()
	return args.Get(0).([]services.ServerInfo), args.Error(1)
}

func (m *MockServerManager) GetServer(ctx context.Context, idOrName string) (*services.ServerInfo, error) {
	args := m.Called(ctx, idOrName)
	return args.Get(0).(*services.ServerInfo), args.Error(1)
}

func createTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.janitarr.key")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// executeCommand is a helper to execute Cobra commands and capture output
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

// Helper function to simulate stdin input
func simulateStdin(input string) *os.File {
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	_ = w.Close()
	return r
}

func TestServerAdd(t *testing.T) {
	assert := assert.New(t)

	// Override NewServerManagerFunc for testing
	originalNewServerManagerFunc := services.NewServerManagerFunc
	defer func() { services.NewServerManagerFunc = originalNewServerManagerFunc }()

	mockServerManager := new(MockServerManager)
	services.NewServerManagerFunc = func(db *database.DB) services.ServerManagerInterface {
		return mockServerManager
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(serverCmd)

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDB(t), nil
	}

	t.Run("add server - success", func(t *testing.T) {
		testServer := services.ServerInfo{
			ID:        uuid.New().String(),
			Name:      "TestRadarr",
			URL:       "http://localhost:7878",
			Type:      "radarr",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockServerManager.On("AddServer", mock.Anything, "TestRadarr", "http://localhost:7878", "test_apikey", "radarr").Return(&testServer, nil).Once()

		input := "TestRadarr\nradarr\nhttp://localhost:7878\ntest_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "add")
		assert.NoError(err)
		assert.Contains(output, "Add New Server")
		assert.Contains(output, success(fmt.Sprintf("Server '%s' (%s) added successfully!", testServer.Name, testServer.Type)))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("add server - empty name", func(t *testing.T) {
		input := "\nradarr\nhttp://localhost:7878\ntest_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "add")
		assert.Error(err)
		assert.Contains(output, "Server name cannot be empty")
	})

	t.Run("add server - invalid type", func(t *testing.T) {
		// Simulates entering "invalid", then "radarr" after error message
		input := "TestServer\ninvalid\nradarr\nhttp://localhost:7878\ntest_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		mockServerManager.On("AddServer", mock.Anything, "TestServer", "http://localhost:7878", "test_apikey", "radarr").Return(
			&services.ServerInfo{
				ID:        uuid.New().String(),
				Name:      "TestServer",
				URL:       "http://localhost:7878",
				Type:      "radarr",
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			nil).Once()

		output, err := executeCommand(rootCmd, "server", "add")
		assert.NoError(err) // Should not error if it retries and then succeeds
		assert.Contains(output, errorMsg("Invalid server type. Must be 'radarr' or 'sonarr'"))
		assert.Contains(output, success(fmt.Sprintf("Server '%s' (%s) added successfully!", "TestServer", "radarr")))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("add server - empty URL", func(t *testing.T) {
		input := "TestRadarr\nradarr\n\ntest_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "add")
		assert.Error(err)
		assert.Contains(output, "Server URL cannot be empty")
	})

	t.Run("add server - empty API key", func(t *testing.T) {
		input := "TestRadarr\nradarr\nhttp://localhost:7878\n\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "add")
		assert.Error(err)
		assert.Contains(output, "API Key cannot be empty")
	})

	t.Run("add server - AddServer returns error", func(t *testing.T) {
		mockError := errors.New("connection failed: failed to connect")
		mockServerManager.On("AddServer", mock.Anything, "TestRadarr", "http://localhost:7878", "test_apikey", "radarr").Return(nil, mockError).Once()

		input := "TestRadarr\nradarr\nhttp://localhost:7878\ntest_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "add")
		assert.Error(err)
		assert.Contains(output, "failed to add server: connection failed: failed to connect")
		mockServerManager.AssertExpectations(t)
	})
}

func TestServerEdit(t *testing.T) {
	assert := assert.New(t)

	// Override NewServerManagerFunc for testing
	originalNewServerManagerFunc := services.NewServerManagerFunc
	defer func() { services.NewServerManagerFunc = originalNewServerManagerFunc }()

	mockServerManager := new(MockServerManager)
	services.NewServerManagerFunc = func(db *database.DB) services.ServerManagerInterface {
		return mockServerManager
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(serverCmd)

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDB(t), nil
	}

	existingServer := services.ServerInfo{
		ID:        uuid.New().String(),
		Name:      "OldName",
		URL:       "http://oldurl.com:7878",
		Type:      "radarr",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		APIKey:    "old_apikey",
	}

	t.Run("edit server - success with all fields changed", func(t *testing.T) {
		mockServerManager.On("GetServer", mock.Anything, existingServer.ID).Return(&existingServer, nil).Once()
		updatedUpdates := services.ServerUpdate{
			Name:   services.StringPtr("NewName"),
			URL:    services.StringPtr("http://newurl.com:7878"),
			APIKey: services.StringPtr("new_apikey"),
		}
		mockServerManager.On("UpdateServer", mock.Anything, existingServer.ID, updatedUpdates).Return(nil).Once()

		input := "NewName\nhttp://newurl.com:7878\nnew_apikey\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "edit", existingServer.ID)
		assert.NoError(err)
		assert.Contains(output, success(fmt.Sprintf("Server '%s' updated successfully!", *updatedUpdates.Name)))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("edit server - success with some fields changed (leaving others blank)", func(t *testing.T) {
		mockServerManager.On("GetServer", mock.Anything, existingServer.Name).Return(&existingServer, nil).Once()
		updatedUpdates := services.ServerUpdate{
			Name: services.StringPtr("ChangedName"),
		}
		mockServerManager.On("UpdateServer", mock.Anything, existingServer.ID, updatedUpdates).Return(nil).Once()

		// User changes name, leaves URL and API key blank
		input := "ChangedName\n\n\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "edit", existingServer.Name)
		assert.NoError(err)
		assert.Contains(output, success(fmt.Sprintf("Server '%s' updated successfully!", *updatedUpdates.Name)))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("edit server - no changes made", func(t *testing.T) {
		mockServerManager.On("GetServer", mock.Anything, existingServer.ID).Return(&existingServer, nil).Once()
		// User leaves all inputs blank
		input := "\n\n\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "edit", existingServer.ID)
		assert.NoError(err)
		assert.Contains(output, info("No changes detected. Skipping update."))
		mockServerManager.AssertExpectations(t) // No UpdateServer call expected
	})

	t.Run("edit server - server not found", func(t *testing.T) {
		mockError := errors.New("server not found")
		mockServerManager.On("GetServer", mock.Anything, "nonexistent").Return(nil, mockError).Once()

		output, err := executeCommand(rootCmd, "server", "edit", "nonexistent")
		assert.Error(err)
		assert.Contains(output, "failed to find server 'nonexistent': server not found")
		mockServerManager.AssertExpectations(t)
	})

	t.Run("edit server - UpdateServer returns error", func(t *testing.T) {
		mockServerManager.On("GetServer", mock.Anything, existingServer.ID).Return(&existingServer, nil).Once()
		mockError := errors.New("update failed: connection refused")
		updatedUpdates := services.ServerUpdate{
			URL: services.StringPtr("http://newfailurl.com"),
		}
		mockServerManager.On("UpdateServer", mock.Anything, existingServer.ID, updatedUpdates).Return(mockError).Once()

		input := "\nhttp://newfailurl.com\n\n"
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = simulateStdin(input)

		output, err := executeCommand(rootCmd, "server", "edit", existingServer.ID)
		assert.Error(err)
		assert.Contains(output, "failed to update server: update failed: connection refused")
		mockServerManager.AssertExpectations(t)
	})
}
