//go:build integration
// +build integration

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startPostgresContainer(t *testing.T, hostPort int) string {
	containerName := fmt.Sprintf("cdmcm-test-postgres-%d", hostPort)
	cmd := exec.Command("docker", "run", "-d", "--name", containerName,
		"-e", "POSTGRES_USER=catalog",
		"-e", "POSTGRES_PASSWORD=catalog123",
		"-e", "POSTGRES_DB=productcatalog",
		"-p", fmt.Sprintf("%d:5432", hostPort),
		"postgres:15-alpine",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run failed: %v - output: %s", err, string(out))
	}
	return containerName
}

func stopContainer(t *testing.T, containerName string) {
	cmd := exec.Command("docker", "rm", "-f", containerName)
	_ = cmd.Run()
}

func waitForPostgres(t *testing.T, host string, port int) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		db, err := store.NewPostgresStore(host, strconv.Itoa(port), "catalog", "catalog123", "productcatalog")
		if err == nil {
			_ = db.DB.Close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("postgres did not become ready within timeout")
}

func TestRunWithPostgresBranch(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is not installed")
	}

	hostPort := freePort(t)
	containerName := startPostgresContainer(t, hostPort)
	defer stopContainer(t, containerName)

	waitForPostgres(t, "127.0.0.1", hostPort)

	orig := listenAndServe
	defer func() { listenAndServe = orig }()

	called := false
	listenAndServe = func(addr string, handler http.Handler) error {
		called = true
		return nil
	}

	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", strconv.Itoa(hostPort))
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

	if err := run(); err != nil {
		t.Fatalf("run with postgres branch failed: %v", err)
	}
	if !called {
		t.Fatal("expected listenAndServe to be called")
	}
}
