package forms

import (
	"strings"
	"testing"
)

func TestValidateServerName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid name",
			input:   "radarr-main",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			input:   "my_server_123",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric",
			input:   "server123",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 51),
			wantErr: true,
		},
		{
			name:    "invalid characters - spaces",
			input:   "my server",
			wantErr: true,
		},
		{
			name:    "invalid characters - special chars",
			input:   "server@123",
			wantErr: true,
		},
		{
			name:    "valid at max length",
			input:   strings.Repeat("a", 50),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServerName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			input:   "http://localhost:7878",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			input:   "https://radarr.example.com",
			wantErr: false,
		},
		{
			name:    "valid with path",
			input:   "http://192.168.1.100:8989/sonarr",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			input:   "localhost:7878",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			input:   "ftp://localhost:7878",
			wantErr: true,
		},
		{
			name:    "missing host",
			input:   "http://",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			input:   "not a url at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid API key",
			input:   "abcd1234efgh5678ijkl9012",
			wantErr: false,
		},
		{
			name:    "valid at min length",
			input:   strings.Repeat("a", 20),
			wantErr: false,
		},
		{
			name:    "valid at max length",
			input:   strings.Repeat("a", 100),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "tooshort",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 101),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServerType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid radarr",
			input:   "radarr",
			wantErr: false,
		},
		{
			name:    "valid sonarr",
			input:   "sonarr",
			wantErr: false,
		},
		{
			name:    "valid radarr uppercase",
			input:   "RADARR",
			wantErr: false,
		},
		{
			name:    "valid sonarr mixed case",
			input:   "SoNaRr",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "plex",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "lidarr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServerType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
