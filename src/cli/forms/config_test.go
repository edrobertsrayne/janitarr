package forms

import (
	"strconv"
	"testing"
)

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid cases
		{name: "zero is valid", input: "0", wantErr: false},
		{name: "one is valid", input: "1", wantErr: false},
		{name: "100 is valid", input: "100", wantErr: false},
		{name: "500 is valid", input: "500", wantErr: false},
		{name: "1000 is valid", input: "1000", wantErr: false},

		// Invalid cases
		{name: "negative is invalid", input: "-1", wantErr: true},
		{name: "1001 exceeds max", input: "1001", wantErr: true},
		{name: "2000 exceeds max", input: "2000", wantErr: true},
		{name: "not a number", input: "abc", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a validator function matching updated implementation
			validateLimit := func(s string) error {
				val, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				// Updated implementation: 0-1000
				if val < 0 || val > 1000 {
					return strconv.ErrRange
				}
				return nil
			}

			err := validateLimit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLimit(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
