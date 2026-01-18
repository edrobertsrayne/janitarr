package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// MockDBLogs is a mock database for log-related operations.
type MockDBLogs struct {
	mock.Mock
}

func (m *MockDBLogs) GetLogs(ctx context.Context, limit, offset int, logTypeFilter, serverNameFilter *string) ([]logger.LogEntry, error) {
	args := m.Called(ctx, limit, offset, logTypeFilter, serverNameFilter)
	return args.Get(0).([]logger.LogEntry), args.Error(1)
}

func (m *MockDBLogs) ClearLogs() error {
	args := m.Called()
	return args.Error(0)
}

func createTestDBLogs(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.janitarr.key")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestLogsCommand(t *testing.T) {
	assert := assert.New(t)

	// Override database.New, database.GetLogsFunc, and database.ClearLogsFunc for testing
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	originalGetLogsFunc := database.GetLogsFunc
	defer func() { database.GetLogsFunc = originalGetLogsFunc }()
	originalClearLogsFunc := database.ClearLogsFunc
	defer func() { database.ClearLogsFunc = originalClearLogsFunc }()

	mockDB := new(MockDBLogs)
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDBLogs(t), nil
	}
	database.GetLogsFunc = func(db *database.DB, limit, offset int, logTypeFilter, serverNameFilter *string) ([]logger.LogEntry, error) {
		return mockDB.GetLogs(context.Background(), limit, offset, logTypeFilter, serverNameFilter)
	}
	database.ClearLogsFunc = func(db *database.DB) error {
		return mockDB.ClearLogs()
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(logsCmd)

	t.Run("logs command - default (20 entries)", func(t *testing.T) {
		entries := []logger.LogEntry{}
		for i := 0; i < 5; i++ { // Just 5 for brevity in test output
			entries = append(entries, logger.LogEntry{
				ID:        uuid.NewString(),
				Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
				Type:      logger.LogTypeCycleStart,
				Message:   fmt.Sprintf("Cycle started #%d", i+1),
			})
		}
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		output, err := executeCommand(rootCmd, "logs")
		assert.NoError(err)
		assert.Contains(output, "Activity Logs:")
		assert.Contains(output, "Cycle started #1")
		assert.Contains(output, "Cycle started #5")
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - limit 2", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeCycleStart, Message: "Cycle started #1"},
			{ID: uuid.NewString(), Timestamp: time.Now().Add(-time.Minute), Type: logger.LogTypeCycleEnd, Message: "Cycle ended #1"},
		}
		mockDB.On("GetLogs", mock.Anything, 2, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		output, err := executeCommand(rootCmd, "logs", "-n", "2")
		assert.NoError(err)
		assert.Contains(output, "Cycle started #1")
		assert.Contains(output, "Cycle ended #1")
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - no entries", func(t *testing.T) {
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return([]logger.LogEntry{}, nil).Once()

		output, err := executeCommand(rootCmd, "logs")
		assert.NoError(err)
		assert.Contains(output, info("No log entries found."))
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - json output", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeCycleStart, Message: "Cycle started #1"},
		}
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		output, err := executeCommand(rootCmd, "logs", "--json")
		assert.NoError(err)

		var actualEntries []logger.LogEntry
		err = json.Unmarshal([]byte(output), &actualEntries)
		assert.NoError(err)
		assert.Len(actualEntries, 1)
		assert.Equal(entries[0].Message, actualEntries[0].Message)
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - clear with confirmation", func(t *testing.T) {
		// Mock confirmation input
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = strings.NewReader("y\n")

		mockDB.On("ClearLogs").Return(nil).Once()

		output, err := executeCommand(rootCmd, "logs", "--clear")
		assert.NoError(err)
		assert.Contains(output, "Are you sure you want to clear all logs?")
		assert.Contains(output, success("All logs cleared successfully."))
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - clear cancelled", func(t *testing.T) {
		// Mock confirmation input
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = strings.NewReader("n\n")

		output, err := executeCommand(rootCmd, "logs", "--clear")
		assert.NoError(err)
		assert.Contains(output, info("Log clearing cancelled."))
		mockDB.AssertNotCalled(t, "ClearLogs")
	})

	t.Run("logs command - error retrieving logs", func(t *testing.T) {
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return(([]logger.LogEntry)(nil), errors.New("db read error")).Once()

		output, err := executeCommand(rootCmd, "logs")
		assert.Error(err)
		assert.Contains(output, errorMsg("failed to retrieve logs: db read error"))
		mockDB.AssertExpectations(t)
	})

	t.Run("logs command - error clearing logs", func(t *testing.T) {
		// Mock confirmation input
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()
		os.Stdin = strings.NewReader("y\n")

		mockDB.On("ClearLogs").Return(errors.New("db clear error")).Once()

		output, err := executeCommand(rootCmd, "logs", "--clear")
		assert.Error(err)
		assert.Contains(output, errorMsg("failed to clear logs: db clear error"))
		mockDB.AssertExpectations(t)
	})
}

// executeCommand is a helper to execute a cobra command and capture its output
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

// keyValue is a helper to format key-value pairs for table output verification
func keyValue(key, value string) string {
	return fmt.Sprintf("%-" + "20" + "s %s", key+":", value)
}

// confirmAction is a helper function to simulate user confirmation for tests.
// This overwrites the global confirmAction for the duration of the test.
var originalConfirmAction func(prompt string) bool

func init() {
	originalConfirmAction = confirmAction
	confirmAction = func(prompt string) bool {
		// Default to yes for tests, unless specifically overridden in a test case
		return true
	}
}
