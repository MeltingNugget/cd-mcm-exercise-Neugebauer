package main

import (
	"net/http"
	"os"
	"testing"
)

func TestGetEnvFallback(t *testing.T) {
	os.Unsetenv("SOME_NON_EXISTENT")
	got := getEnv("SOME_NON_EXISTENT", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestRunUsesListenAndServe(t *testing.T) {
	// mock listenAndServe to avoid binding and to verify argument
	called := false
	orig := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error {
		called = true
		if addr == "" {
			t.Fatalf("expected non-empty addr")
		}
		return nil
	}
	defer func() { listenAndServe = orig }()
	os.Unsetenv("DB_HOST")
	os.Setenv("PORT", "0") // use "0" to indicate no preference, still not actually bound because of mock

	if err := run(); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected listenAndServe to be called")
	}
}
