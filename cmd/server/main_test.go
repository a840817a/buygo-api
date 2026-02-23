package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	const key = "BUYGO_TEST_ENV_OR_DEFAULT"
	t.Setenv(key, "")
	if got := envOrDefault(key, "fallback"); got != "fallback" {
		t.Fatalf("envOrDefault fallback = %q, want %q", got, "fallback")
	}

	t.Setenv(key, "configured")
	if got := envOrDefault(key, "fallback"); got != "configured" {
		t.Fatalf("envOrDefault configured = %q, want %q", got, "configured")
	}

	// Ensure the environment remains visible to subprocesses if needed.
	if got := os.Getenv(key); got != "configured" {
		t.Fatalf("os.Getenv(%q) = %q, want %q", key, got, "configured")
	}
}

func TestCORSHandler_OptionsAndPassThrough(t *testing.T) {
	c := newCORS("https://example.com")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := c.Handler(next)

	optionsReq := httptest.NewRequest(http.MethodOptions, "/health", nil)
	optionsReq.Header.Set("Origin", "https://example.com")
	optionsReq.Header.Set("Access-Control-Request-Method", "POST")
	optionsRec := httptest.NewRecorder()
	handler.ServeHTTP(optionsRec, optionsReq)

	if optionsRec.Code != http.StatusOK {
		t.Fatalf("OPTIONS status = %d, want %d", optionsRec.Code, http.StatusOK)
	}
	if got := optionsRec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("allow origin = %q, want %q", got, "https://example.com")
	}

	// Cross-origin GET should receive CORS response headers
	getReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	getReq.Header.Set("Origin", "https://example.com")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNoContent {
		t.Fatalf("GET status = %d, want %d", getRec.Code, http.StatusNoContent)
	}
	if got := getRec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q, want %q", got, "true")
	}
}
