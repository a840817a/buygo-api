package handler

import "testing"

func TestDecodePageToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    int
		wantErr bool
	}{
		{name: "empty token", token: "", want: 0, wantErr: false},
		{name: "zero", token: "0", want: 0, wantErr: false},
		{name: "positive", token: "40", want: 40, wantErr: false},
		{name: "negative", token: "-1", wantErr: true},
		{name: "invalid", token: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodePageToken(tt.token)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("decodePageToken() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEncodePageToken(t *testing.T) {
	if got := encodePageToken(0); got != "" {
		t.Fatalf("encodePageToken(0) = %q, want empty", got)
	}
	if got := encodePageToken(20); got != "20" {
		t.Fatalf("encodePageToken(20) = %q, want %q", got, "20")
	}
}

func TestNormalizePageSize(t *testing.T) {
	if got := normalizePageSize(0); got != 20 {
		t.Fatalf("normalizePageSize(0) = %d, want 20", got)
	}
	if got := normalizePageSize(-1); got != 20 {
		t.Fatalf("normalizePageSize(-1) = %d, want 20", got)
	}
	if got := normalizePageSize(200); got != 100 {
		t.Fatalf("normalizePageSize(200) = %d, want 100", got)
	}
	if got := normalizePageSize(50); got != 50 {
		t.Fatalf("normalizePageSize(50) = %d, want 50", got)
	}
}
