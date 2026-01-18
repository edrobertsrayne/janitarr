package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

func TestFormatServerTable(t *testing.T) {
	assert := assert.New(t)

	// Test case 1: No servers
	t.Run("no servers", func(t *testing.T) {
		servers := []services.ServerInfo{}
		expected := info("No servers configured.")
		assert.Equal(expected, formatServerTable(servers))
	})

	// Test case 2: Single server
	t.Run("single server", func(t *testing.T) {
		servers := []services.ServerInfo{
			{
				ID:        "1",
				Name:      "MyRadarr",
				URL:       "http://localhost:7878",
				Type:      "radarr",
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		output := formatServerTable(servers)
		assert.Contains(output, "Configured Servers")
		assert.Contains(output, "MyRadarr")
		assert.Contains(output, "Radarr")
		assert.Contains(output, "http://localhost:7878")
		assert.Contains(output, success("Yes"))
	})

	// Test case 3: Multiple servers, mixed types and enabled status
	t.Run("multiple servers", func(t *testing.T) {
		servers := []services.ServerInfo{
			{
				ID:        "1",
				Name:      "MyRadarr",
				URL:       "http://localhost:7878",
				Type:      "radarr",
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "2",
				Name:      "MySonarr",
				URL:       "http://localhost:8989",
				Type:      "sonarr",
				Enabled:   false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		output := formatServerTable(servers)
		assert.Contains(output, "Configured Servers")
		assert.Contains(output, "MyRadarr")
		assert.Contains(output, "Radarr")
		assert.Contains(output, success("Yes"))
		assert.Contains(output, "MySonarr")
		assert.Contains(output, "Sonarr")
		assert.Contains(output, warning("No"))
	})

	// Test case 4: Long names/URLs to check width calculation
	t.Run("long names/urls", func(t *testing.T) {
		servers := []services.ServerInfo{
			{
				ID:        "1",
				Name:      "VeryLongRadarrServerNameIndeed",
				URL:       "http://very.long.url.for.radarr.com:7878/api/v3",
				Type:      "radarr",
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		output := formatServerTable(servers)
		assert.Contains(output, "VeryLongRadarrServerNameIndeed")
		assert.Contains(output, "http://very.long.url.for.radarr.com:7878/api/v3")

		// Verify header and row line up based on computed width
		lines := strings.Split(output, "\n")
		// Check that header and underline have similar structure
		// This is a bit fragile but checks basic alignment
		assert.Equal(lines[3][0:len("VeryLongRadarrServerNameIndeed")], "VeryLongRadarrServerNameIndeed")
		assert.Equal(lines[4][0:len("------------------------------")], strings.Repeat("-", len("VeryLongRadarrServerNameIndeed")))
	})
}

func TestFormatLogTable(t *testing.T) {
	assert := assert.New(t)

	// Test case 1: No logs
	t.Run("no logs", func(t *testing.T) {
		logs := []logger.LogEntry{}
		expected := info("No log entries.")
		assert.Equal(expected, formatLogTable(logs))
	})

	// Test case 2: Mixed log types
	t.Run("mixed log types", func(t *testing.T) {
		testTime := time.Date(2026, time.January, 18, 10, 0, 0, 0, time.UTC)
		logs := []logger.LogEntry{
			{
				ID:        "1",
				Timestamp: testTime,
				Type:      logger.LogTypeCycleStart,
				Message:   "Automation cycle started.",
				IsManual:  true,
			},
			{
				ID:         "2",
				Timestamp:  testTime.Add(1 * time.Minute),
				Type:       logger.LogTypeSearch,
				ServerName: "MyRadarr",
				ServerType: database.ServerTypeRadarr,
				Category:   database.SearchCategoryMissing,
				Count:      5,
				Message:    "Triggered searches.",
				IsManual:   true,
			},
			{
				ID:        "3",
				Timestamp: testTime.Add(2 * time.Minute),
				Type:      logger.LogTypeError,
				Message:   "API connection failed.",
				IsManual:  false,
			},
			{
				ID:        "4",
				Timestamp: testTime.Add(3 * time.Minute),
				Type:      logger.LogTypeCycleEnd,
				Message:   "Automation cycle finished.",
				IsManual:  false,
			},
		}

		output := formatLogTable(logs)
		assert.Contains(output, "Activity Logs")
		assert.Contains(output, "Automation cycle started.")
		assert.Contains(output, "Triggered searches.")
		assert.Contains(output, "API connection failed.")
		assert.Contains(output, "Automation cycle finished.")

		// Check for specific formatting of search log details
		assert.Contains(output, "Radarr (MyRadarr) - missing: 5 items")
		assert.Contains(output, errorMsg("API connection failed."))
		assert.Contains(output, errorMsg(string(logger.LogTypeError)))
		assert.Contains(output, success(string(logger.LogTypeSearch)))
	})

	// Test case 3: Long message truncation
	t.Run("long message truncation", func(t *testing.T) {
		testTime := time.Date(2026, time.January, 18, 10, 0, 0, 0, time.UTC)
		longMessage := strings.Repeat("a", 100)
		logs := []logger.LogEntry{
			{
				ID:        "1",
				Timestamp: testTime,
				Type:      logger.LogTypeCycleStart,
				Message:   longMessage,
				IsManual:  false,
			},
		}
		output := formatLogTable(logs)
		assert.Contains(output, longMessage[:67]+"...") // 70-3
	})
}

func TestFormatConfigTable(t *testing.T) {
	assert := assert.New(t)

	// Test case 1: Default config
	t.Run("default config", func(t *testing.T) {
		config := &database.AppConfig{
			Schedule: database.ScheduleConfig{
				IntervalHours: 6,
				Enabled:       true,
			},
			SearchLimits: database.SearchLimits{
				MissingMoviesLimit:   10,
				MissingEpisodesLimit: 10,
				CutoffMoviesLimit:    5,
				CutoffEpisodesLimit:  5,
			},
		}

		output := formatConfigTable(config)
		assert.Contains(output, "Configuration")
		assert.Contains(output, colorBold+"Schedule:"+colorReset)
		assert.Contains(output, keyValue("Enabled", success("Yes")))
		assert.Contains(output, keyValue("Interval", "6 hours"))
		assert.Contains(output, colorBold+"Search Limits:"+colorReset)
		assert.Contains(output, keyValue("Missing Movies", "10 items"))
		assert.Contains(output, keyValue("Missing Episodes", "10 items"))
		assert.Contains(output, keyValue("Cutoff Movies", "5 items"))
		assert.Contains(output, keyValue("Cutoff Episodes", "5 items"))
	})

	// Test case 2: Custom config with disabled features
	t.Run("custom config", func(t *testing.T) {
		config := &database.AppConfig{
			Schedule: database.ScheduleConfig{
				IntervalHours: 12,
				Enabled:       false,
			},
			SearchLimits: database.SearchLimits{
				MissingMoviesLimit:   0,
				MissingEpisodesLimit: 20,
				CutoffMoviesLimit:    0,
				CutoffEpisodesLimit:  0,
			},
		}

		output := formatConfigTable(config)
		assert.Contains(output, keyValue("Enabled", warning("No")))
		assert.Contains(output, keyValue("Interval", "12 hours"))
		assert.Contains(output, keyValue("Missing Movies", warning("Disabled")))
		assert.Contains(output, keyValue("Missing Episodes", "20 items"))
		assert.Contains(output, keyValue("Cutoff Movies", warning("Disabled")))
		assert.Contains(output, keyValue("Cutoff Episodes", warning("Disabled")))
	})
}
