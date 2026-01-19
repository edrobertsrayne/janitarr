package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
	"github.com/user/janitarr/src/web"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start Janitarr in development mode (verbose logging)",
	Long:  `Starts the automation scheduler and web server with verbose logging and debug output.`,
	RunE:  runDev,
}

func init() {
	devCmd.Flags().IntP("port", "p", 3434, "Web server port")
	devCmd.Flags().String("host", "localhost", "Web server host")
}

func runDev(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")

	// Display development mode banner
	fmt.Println("========================================")
	fmt.Println("  " + warning("DEVELOPMENT MODE"))
	fmt.Println("  Verbose logging enabled")
	fmt.Println("  Stack traces in error responses")
	fmt.Println("  HTTP request logging enabled")
	fmt.Println("========================================")
	fmt.Println()

	// Validate port range
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", port)
	}

	// Ensure data directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize database
	keyPath := filepath.Join(dbDir, ".janitarr.key")
	db, err := database.New(dbPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get configuration
	config := db.GetAppConfig()

	// Parse log level from flag, default to debug in dev mode if not explicitly set
	level := logger.LevelDebug
	if cmd.Flags().Changed("log-level") {
		parsedLevel, err := logger.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level: %w", err)
		}
		level = parsedLevel
	}

	// Initialize logger with configured level in development mode
	appLogger := logger.NewLogger(db, level, true)

	// Initialize services
	detector := services.NewDetector(db)
	searchTrigger := services.NewSearchTrigger(db, appLogger)
	automation := services.NewAutomation(db, detector, searchTrigger, appLogger)

	// Create scheduler with automation callback wrapper
	schedulerCallback := func(ctx context.Context, isManual bool) error {
		_, err := automation.RunCycle(ctx, isManual, false) // dryRun = false
		return err
	}
	scheduler := services.NewScheduler(db, config.Schedule.IntervalHours, schedulerCallback).WithLogger(appLogger)

	// Start scheduler if enabled
	ctx := context.Background()
	if config.Schedule.Enabled {
		if err := scheduler.Start(ctx); err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}
		fmt.Printf("✓ Scheduler started (interval: %d hours)\n", config.Schedule.IntervalHours)
	} else {
		fmt.Println("⚠ Warning: Scheduler is disabled in configuration")
		fmt.Println("  Use 'janitarr config set schedule.enabled true' to enable")
	}

	// Initialize web server with development mode enabled
	server := web.NewServer(web.ServerConfig{
		Port:      port,
		Host:      host,
		DB:        db,
		Logger:    appLogger,
		Scheduler: scheduler,
		IsDev:     true, // Enable development features
	})

	// Display startup information
	fmt.Println("\n" + success("Janitarr started successfully!"))
	fmt.Printf("Web UI:  http://%s:%d\n", host, port)
	fmt.Printf("API:     http://%s:%d/api\n", host, port)
	fmt.Printf("Metrics: http://%s:%d/metrics\n", host, port)
	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println()

	// Setup graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-stopChan:
		fmt.Println("\n\nShutdown signal received...")
	case err := <-serverErrChan:
		return fmt.Errorf("web server error: %w", err)
	}

	// Graceful shutdown
	return gracefulShutdown(scheduler, server, db)
}
