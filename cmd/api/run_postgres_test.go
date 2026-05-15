//go:build integration

package main

import (
	"net/http"
	"os"
	"testing"
)

func TestRun_WithPostgres(t *testing.T) {
	// Echte DB muss laufen (z.B. via docker-compose in CI)
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "catalog")
	os.Setenv("DB_PASSWORD", "catalog123")
	os.Setenv("DB_NAME", "productcatalog")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	orig := listenAndServe
	defer func() { listenAndServe = orig }()
	listenAndServe = func(addr string, handler http.Handler) error {
		return nil
	}

	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_WithPostgres_ConnectionError(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "wronguser")
	os.Setenv("DB_PASSWORD", "wrongpassword")
	os.Setenv("DB_NAME", "productcatalog")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	if err := run(); err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
}
