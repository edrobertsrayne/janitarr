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

func TestServerList(t *testing.T) {
	assert := assert.New(t)

	// Override NewServerManagerFunc for testing
	originalNewServerManagerFunc := services.NewServerManagerFunc
	defer func() { services.NewServerManagerFunc = originalNewServerManagerFunc }()

	mockServerManager := new(MockServerManager)
	services.NewServerManagerFunc = func(db *database.DB) services.ServerManagerInterface {
		return mockServerManager
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(serverCmd) // Ensure serverCmd is added

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDB(t), nil
	}

	t.Run("list servers - no servers", func(t *testing.T) {
		mockServerManager.On("ListServers").Return([]services.ServerInfo{}, nil).Once()

		output, err := executeCommand(rootCmd, "server", "list")
		assert.NoError(err)
		assert.Contains(output, info("No servers configured."))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("list servers - with servers (table format)", func(t *testing.T) {
		server1 := services.ServerInfo{
			ID:        uuid.New().String(),
			Name:      "MyRadarr",
			URL:       "http://localhost:7878",
			Type:      "radarr",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		server2 := services.ServerInfo{
			ID:        uuid.New().String(),
			Name:      "MySonarr",
			URL:       "http://localhost:8989",
			Type:      "sonarr",
			Enabled:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockServers := []services.ServerInfo{server1, server2}
		mockServerManager.On("ListServers").Return(mockServers, nil).Once()

		output, err := executeCommand(rootCmd, "server", "list")
		assert.NoError(err)
		assert.Contains(output, "Configured Servers")
		assert.Contains(output, server1.Name)
		assert.Contains(output, server1.URL)
		assert.Contains(output, strings.Title(server1.Type))
		assert.Contains(output, success("Yes"))
		assert.Contains(output, server2.Name)
		assert.Contains(output, server2.URL)
		assert.Contains(output, strings.Title(server2.Type))
		assert.Contains(output, warning("No"))
		mockServerManager.AssertExpectations(t)
	})

	t.Run("list servers - with servers (JSON format)", func(t *testing.T) {
		server1 := services.ServerInfo{
			ID:        uuid.New().String(),
			Name:      "MyRadarr",
			URL:       "http://localhost:7878",
			Type:      "radarr",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockServers := []services.ServerInfo{server1}
		mockServerManager.On("ListServers").Return(mockServers, nil).Once()

		output, err := executeCommand(rootCmd, "server", "list", "--json")
		assert.NoError(err)

		var actualServers []services.ServerInfo
		err = json.Unmarshal([]byte(output), &actualServers)
		assert.NoError(err)
		assert.Len(actualServers, 1)
		assert.Equal(server1.Name, actualServers[0].Name)
		assert.Equal(server1.Type, actualServers[0].Type)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("list servers - error from server manager", func(t *testing.T) {
		mockError := errors.New("failed to retrieve servers")
		mockServerManager.On("ListServers").Return([]services.ServerInfo{}, mockError).Once()

		output, err := executeCommand(rootCmd, "server", "list")
		assert.Error(err)
		assert.Contains(output, "Error: failed to list servers: failed to retrieve servers")
		mockServerManager.AssertExpectations(t)
	})
}