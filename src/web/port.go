package web

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)

	// Try to listen on the port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()

	// Small delay to ensure port is fully released
	time.Sleep(10 * time.Millisecond)
	return true
}
