package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
	"github.com/user/janitarr/src/web"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Janitarr in production mode (scheduler + web server)",
	Long:  `Starts the automation scheduler and web server for production use.`,
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntP("port", "p", 3434, "Web server port")
	startCmd.Flags().String("host", "localhost", "Web server host")
}

func runStart(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")

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

	// Initialize logger
	appLogger := logger.NewLogger(db)

	// Initialize services
	detector := services.NewDetector(db)
	searchTrigger := services.NewSearchTrigger(db)
	automation := services.NewAutomation(db, detector, searchTrigger, appLogger)

	// Create scheduler with automation callback wrapper
	schedulerCallback := func(ctx context.Context, isManual bool) error {
		_, err := automation.RunCycle(ctx, isManual, false) // dryRun = false
		return err
	}
	scheduler := services.NewScheduler(db, config.Schedule.IntervalHours, schedulerCallback)

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

	// Initialize web server
	server := web.NewServer(web.ServerConfig{
		Port:      port,
		Host:      host,
		DB:        db,
		Logger:    appLogger,
		Scheduler: scheduler,
		IsDev:     false,
	})

	// Display startup information
	fmt.Println("\n" + success("Janitarr started successfully!"))
	fmt.Printf("Web UI:  http://%s:%d\n", host, port)
	fmt.Printf("API:     http://%s:%d/api\n", host, port)
	fmt.Printf("Metrics: http://%s:%d/metrics\n", host, port)
	fmt.Println("\nPress Ctrl+C to stop")

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

func gracefulShutdown(scheduler *services.Scheduler, server *web.Server, db *database.DB) error {
	fmt.Println("Stopping services...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Stop scheduler (wait for active cycle)
	fmt.Println("  Stopping scheduler...")
	scheduler.Stop()
	fmt.Println("  ✓ Scheduler stopped")

	// 2. Close WebSocket connections
	fmt.Println("  Closing WebSocket connections...")
	server.CloseWebSockets()
	fmt.Println("  ✓ WebSocket connections closed")

	// 3. Stop web server (wait for in-flight requests)
	fmt.Println("  Stopping web server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("  ⚠ Web server shutdown error: %v\n", err)
	} else {
		fmt.Println("  ✓ Web server stopped")
	}

	// 4. Close database
	fmt.Println("  Closing database...")
	if err := db.Close(); err != nil {
		fmt.Printf("  ⚠ Database close error: %v\n", err)
	} else {
		fmt.Println("  ✓ Database closed")
	}

	fmt.Println("\n" + success("Shutdown complete"))
	return nil
}
