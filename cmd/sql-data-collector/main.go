package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func main() {
	log.Println("SQL Data Collector starting...")

	// Read database credentials from environment variables.
	obUser := os.Getenv("OB_USER")
	obPassword := os.Getenv("OB_PASSWORD")
	obHost := os.Getenv("OB_HOST")
	obPortStr := os.Getenv("OB_PORT")

	if obUser == "" {
		log.Fatal("Environment variable OB_USER is not set.")
	}
	if obPassword == "" {
		log.Fatal("Environment variable OB_PASSWORD is not set.")
	}
	if obHost == "" {
		obHost = "127.0.0.1"
	}
	if obPortStr == "" {
		obPortStr = "2881"
	}

	obPort, err := strconv.ParseInt(obPortStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid OB_PORT: %v", err)
	}

	// Create a new OceanBase data source
	ds := connector.NewOceanBaseDataSource(obHost, obPort, obUser, "sys", obPassword, "oceanbase")

	// Create a new connector
	conn := database.NewConnector(ds)

	if err := conn.Init(); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Create a new operation manager
	manager := operation.NewOceanbaseOperationManager(conn)

	config := &Config{
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
	collector, err := NewCollector(config, manager)
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
