package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/mrckurz/CI-CD-MCM/internal/handler"
	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

var listenAndServe = http.ListenAndServe

func run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		pgStore, err := store.NewPostgresStore(
			dbHost,
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "catalog"),
			getEnv("DB_PASSWORD", "catalog123"),
			getEnv("DB_NAME", "productcatalog"),
		)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		if err := pgStore.EnsureTable(); err != nil {
			pgStore.DB.Close()
			log.Fatalf("Failed to create table: %v", err)
			return fmt.Errorf("failed to create table: %w", err)
		}

		h := handler.NewPostgresHandler(pgStore)
		h.RegisterRoutes(r)
		fmt.Printf("Product Catalog API (PostgreSQL) listening on :%s\n", port)
	} else {
		memStore := store.NewMemoryStore()
		h := handler.NewHandler(memStore)
		h.RegisterRoutes(r)
		fmt.Printf("Product Catalog API (in-memory) listening on :%s\n", port)
	}

	return listenAndServe(":"+port, r)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
