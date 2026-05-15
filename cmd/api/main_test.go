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

func TestGetEnvReturnsExistingValue(t *testing.T) {
	os.Setenv("MY_TEST_ENV", "value")
	defer os.Unsetenv("MY_TEST_ENV")

	if got := getEnv("MY_TEST_ENV", "fallback"); got != "value" {
		t.Fatalf("expected value, got %s", got)
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

func TestRunUsesDefaultPortWhenPortUnset(t *testing.T) {
	orig := listenAndServe
	defer func() { listenAndServe = orig }()

	called := false
	listenAndServe = func(addr string, handler http.Handler) error {
		called = true
		if addr != ":8080" {
			t.Fatalf("expected default address :8080, got %s", addr)
		}
		return nil
	}

	os.Unsetenv("PORT")
	os.Unsetenv("DB_HOST")

	if err := run(); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !called {
		t.Fatal("expected listenAndServe to be called")
	}
}

func TestMainExecutesRun(t *testing.T) {
	orig := listenAndServe
	defer func() { listenAndServe = orig }()

	called := false
	listenAndServe = func(addr string, handler http.Handler) error {
		called = true
		return nil
	}

	os.Unsetenv("DB_HOST")
	os.Setenv("PORT", "0")
	defer os.Unsetenv("PORT")

	main()
	if !called {
		t.Fatal("expected main to call run and invoke listenAndServe")
	}
}
