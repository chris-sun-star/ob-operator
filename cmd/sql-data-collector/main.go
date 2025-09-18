package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("SQL Data Collector starting...")

	// Read database credentials from environment variables.
	obUser := os.Getenv("OB_USER")
	obPassword := os.Getenv("OB_PASSWORD")

	if obUser == "" {
		log.Fatal("Environment variable OB_USER is not set.")
	}
	if obPassword == "" {
		log.Fatal("Environment variable OB_PASSWORD is not set.")
	}

	// TODO: The host and port should be discovered dynamically for all observers.
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:2881)/oceanbase?parseTime=true", obUser, obPassword)

	config := &Config{
		DSN:      dsn,
		BatchSize: 1000,
		Interval:  30 * time.Second,
	}

	duckDBPath := "sql_audit.duckdb"

	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping collector...")
		cancel()
	}()

	// Initialize the OceanBase collector.
	collector, err := NewCollector(config)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Close()

	// Initialize the DuckDB manager.
	duckdbManager, err := NewDuckDBManager(duckDBPath)
	if err != nil {
		log.Fatalf("Failed to create DuckDB manager: %v", err)
	}
	defer duckdbManager.Close()

	// Run the collection loop.
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Run a collection immediately at startup.
	runCollection(ctx, collector, duckdbManager)

	for {
		select {
		case <-ticker.C:
			runCollection(ctx, collector, duckdbManager)
		case <-ctx.Done():
			log.Println("Collector stopped.")
			return
		}
	}
}

// runCollection performs one full collection and insertion cycle.
func runCollection(ctx context.Context, coll *Collector, mgr *DuckDBManager) {
	log.Println("Running collection cycle...")
	results, err := coll.Collect(ctx)
	if err != nil {
		log.Printf("Error during collection: %v", err)
		return
	}

	if len(results) > 0 {
		if err := mgr.InsertBatch(results); err != nil {
			log.Printf("Error inserting data into DuckDB: %v", err)
		}
	}
}
