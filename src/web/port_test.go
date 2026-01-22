package web

import (
	"net"
	"testing"
)

func TestIsPortAvailable(t *testing.T) {
	// Test that a random high port is available
	if !IsPortAvailable("localhost", 59999) {
		t.Skip("Port 59999 unexpectedly in use")
	}

	// Occupy a port
	listener, err := net.Listen("tcp", "localhost:59998")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()

	// Test that occupied port is not available
	if IsPortAvailable("localhost", 59998) {
		t.Error("Expected port 59998 to be unavailable")
	}
}
