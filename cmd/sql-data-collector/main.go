package main

import (
	"context"
	"database/sql"
	"fmt"
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

	// Read configuration from environment variables.
	obUser := os.Getenv("OB_USER")
	obPassword := os.Getenv("OB_PASSWORD")
	obHost := os.Getenv("OB_HOST")
	obPortStr := os.Getenv("OB_PORT")
	obTenant := os.Getenv("OB_TENANT")

	if obUser == "" || obPassword == "" || obTenant == "" {
		log.Fatal("OB_USER, OB_PASSWORD, and OB_TENANT environment variables must be set.")
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

	// Create a new OceanBase data source for the sys tenant to discover the target tenant ID.
	ds := connector.NewOceanBaseDataSource(obHost, obPort, obUser, "sys", obPassword, "oceanbase")

	// Create a new connector
	conn := database.NewConnector(ds)
	if err := conn.Init(); err != nil {
		log.Fatalf("Failed to initialize connector: %v", err)
	}

	// Create a new operation manager
	manager := operation.NewOceanbaseOperationManager(conn)
	defer manager.Close()

	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping collector...")
		cancel()
	}()

	// Get the tenant ID for the specified tenant name.
	obTenantID, err := getTenantIDByName(ctx, manager, obTenant)
	if err != nil {
		log.Fatalf("Failed to get tenant ID for tenant %s: %v", obTenant, err)
	}
	log.Printf("Found tenant '%s' with ID %d", obTenant, obTenantID)

	config := &Config{
		BatchSize: 1000,
		Interval:  30 * time.Second,
	}

	duckDBPath := fmt.Sprintf("sql_audit_tenant_%s.duckdb", obTenant)

	// Initialize the OceanBase collector.
	collector, err := NewCollector(config, manager, obTenantID)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

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

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, manager *operation.OceanbaseOperationManager, tenantName string) (int64, error) {
	var tenant Tenant
	err := manager.QueryRow(ctx, &tenant, "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?", tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}