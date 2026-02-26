package handler

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		label   string
		wantErr bool
	}{
		{"empty string returns error", "", "title", true},
		{"non-empty string passes", "hello", "title", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequired(tt.field, tt.label)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if connect.CodeOf(err) != connect.CodeInvalidArgument {
					t.Errorf("got code %v, want InvalidArgument", connect.CodeOf(err))
				}
				if !strings.Contains(err.Error(), tt.label) {
					t.Errorf("error %q should mention field name %q", err.Error(), tt.label)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		label   string
		max     int
		wantErr bool
	}{
		{"within limit", "abc", "name", 5, false},
		{"at limit", "abcde", "name", 5, false},
		{"exceeds limit", "abcdef", "name", 5, true},
		{"empty string passes", "", "name", 5, false},
		{"multibyte runes counted correctly", "你好世界！！", "name", 5, true},
		{"multibyte within limit", "你好", "name", 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaxLength(tt.field, tt.label, tt.max)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if connect.CodeOf(err) != connect.CodeInvalidArgument {
					t.Errorf("got code %v, want InvalidArgument", connect.CodeOf(err))
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePositiveInt64(t *testing.T) {
	tests := []struct {
		name    string
		val     int64
		wantErr bool
	}{
		{"positive value passes", 1, false},
		{"large positive passes", 9999999, false},
		{"zero returns error", 0, true},
		{"negative returns error", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveInt64(tt.val, "amount")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if connect.CodeOf(err) != connect.CodeInvalidArgument {
					t.Errorf("got code %v, want InvalidArgument", connect.CodeOf(err))
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePositiveFloat64(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		wantErr bool
	}{
		{"positive value passes", 0.01, false},
		{"large positive passes", 99999.99, false},
		{"zero returns error", 0, true},
		{"negative returns error", -0.5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveFloat64(tt.val, "price")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if connect.CodeOf(err) != connect.CodeInvalidArgument {
					t.Errorf("got code %v, want InvalidArgument", connect.CodeOf(err))
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
