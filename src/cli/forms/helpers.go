package forms

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// IsInteractive returns true if stdin is a TTY (interactive terminal)
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ValidateServerName validates that a server name meets requirements:
// - Required (non-empty)
// - Alphanumeric with dashes and underscores
// - 1-50 characters
func ValidateServerName(s string) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return fmt.Errorf("server name is required")
	}

	if len(s) > 50 {
		return fmt.Errorf("server name must be 50 characters or less")
	}

	// Allow alphanumeric, dash, underscore
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("server name must contain only letters, numbers, dashes, and underscores")
	}

	return nil
}

// ValidateURL validates that a URL meets requirements:
// - Required (non-empty)
// - Valid URL format
// - Has http or https scheme
func ValidateURL(s string) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return fmt.Errorf("URL is required")
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateAPIKey validates that an API key meets requirements:
// - Required (non-empty)
// - Reasonable length (20-100 characters)
func ValidateAPIKey(s string) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return fmt.Errorf("API key is required")
	}

	if len(s) < 20 {
		return fmt.Errorf("API key is too short (minimum 20 characters)")
	}

	if len(s) > 100 {
		return fmt.Errorf("API key is too long (maximum 100 characters)")
	}

	return nil
}

// ValidateServerType validates that a server type is either "radarr" or "sonarr"
func ValidateServerType(s string) error {
	s = strings.ToLower(strings.TrimSpace(s))

	if s == "" {
		return fmt.Errorf("server type is required")
	}

	if s != "radarr" && s != "sonarr" {
		return fmt.Errorf("server type must be 'radarr' or 'sonarr'")
	}

	return nil
}
