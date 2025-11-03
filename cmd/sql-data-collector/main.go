package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	sqldatacollector "github.com/oceanbase/ob-operator/internal/sql-data-collector"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	CompactionThreshold = 120
	PlanWorkerCount     = 4
)

func main() {
	// Configure logging
	logDir := "./log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "sql-data-collector.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(multiWriter)))

	log.Println("SQL Data Collector starting...")

	// Read configuration from environment variables.
	obClusterName := os.Getenv("OB_CLUSTER_NAME")
	obClusterNamespace := os.Getenv("OB_CLUSTER_NAMESPACE")
	obTenant := os.Getenv("OB_TENANT")
	dataPath := os.Getenv("DATA_PATH")

	if obClusterName == "" || obClusterNamespace == "" || obTenant == "" {
		log.Fatal("OB_CLUSTER_NAME, OB_CLUSTER_NAMESPACE, and OB_TENANT environment variables must be set.")
	}
	if dataPath == "" {
		dataPath = "."
	}
	planDir := filepath.Join(dataPath, "sql_plan")
	if err := os.MkdirAll(planDir, 0755); err != nil {
		log.Fatalf("Failed to create plan data directory: %v", err)
	}
	planDataDb := filepath.Join(planDir, "sql_plan.duckdb")

	// Create a Kubernetes client.
	k8sConfig, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes config: %v", err)
	}
	v1alpha1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(k8sConfig, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping collector...")
		cancel()
	}()

	// Get the OBCluster resource.
	obcluster := &v1alpha1.OBCluster{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: obClusterName, Namespace: obClusterNamespace}, obcluster); err != nil {
		log.Fatalf("Failed to get OBCluster resource: %v", err)
	}

	// Create the connection manager.
	logger := logf.Log.WithName("collector")
	connManager := sqldatacollector.NewConnectionManager(k8sClient, logger, obcluster)
	defer connManager.Close()

	// Get an initial connection to retrieve the tenant ID.
	var obTenantID int64

	for {
		tenantID, err := getTenantIDByName(ctx, connManager, obTenant)
		if err != nil {
			log.Printf("Failed to get tenant ID for tenant %s: %v. Retrying in 10 seconds...", obTenant, err)
		} else {
			obTenantID = tenantID
			log.Printf("Found tenant '%s' with ID %d", obTenant, obTenantID)
			break // Success
		}

		// Wait before retrying or exit if context is cancelled.
		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			log.Println("Collector stopped during tenant ID retrieval.")
			return
		}
	}

	// Configure collection interval
	intervalSeconds := 30
	intervalStr := os.Getenv("COLLECTION_INTERVAL_SECONDS")
	if intervalStr != "" {
		if val, err := strconv.Atoi(intervalStr); err == nil && val > 0 {
			intervalSeconds = val
		} else {
			log.Printf("Invalid COLLECTION_INTERVAL_SECONDS value '%s', using default of 30 seconds.", intervalStr)
		}
	}

	config := &sqldatacollector.Config{
		Interval: time.Duration(intervalSeconds) * time.Second,
	}

	duckDBPath := filepath.Join(dataPath, "sql_audit")

	// Initialize the DuckDB manager.
	duckdbManager, err := sqldatacollector.NewDuckDBManager(duckDBPath)
	if err != nil {
		log.Fatalf("Failed to create DuckDB manager: %v", err)
	}
	defer duckdbManager.Close()

	// Initialize the PlanStore.
	planStore, err := sqldatacollector.NewPlanStore(planDataDb)
	if err != nil {
		log.Fatalf("Failed to create PlanStore: %v", err)
	}
	defer planStore.Close()

	// Create channels for plan collection
	planIdentifierChan := make(chan sqldatacollector.PlanIdentifier, 100)

	// Initialize the PlanCollector.
	planCollector := sqldatacollector.NewPlanCollector(planIdentifierChan)

	// Start plan workers
	var wg sync.WaitGroup
	wg.Add(PlanWorkerCount)
	for i := 0; i < PlanWorkerCount; i++ {
		worker := sqldatacollector.NewPlanWorker(connManager, planIdentifierChan, planStore, &wg)
		go worker.Start(ctx)
	}

	// Retrieve the last known request IDs from DuckDB to resume progress.
	lastRequestIDs, err := duckdbManager.GetLastRequestIDs()
	if err != nil {
		log.Fatalf("Failed to get last request IDs from DuckDB: %v", err)
	}
	log.Printf("Retrieved progress for %d observers from DuckDB.", len(lastRequestIDs))

	// Initialize the OceanBase collector with the retrieved progress.
	collector := sqldatacollector.NewCollector(config, obTenantID, lastRequestIDs)

	// Run the collection loop.
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Start the cleanup routine for old data
	retentionStr := os.Getenv("DATA_RETENTION_DAYS")
	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil {
		log.Fatalf("Invalid or missing DATA_RETENTION_DAYS environment variable: %v", err)
	}

	go func() {
		// Run cleanup once at startup
		log.Println("Running initial cleanup of old data...")
		if err := duckdbManager.DeleteOldData(retentionDays); err != nil {
			log.Printf("Error during initial data cleanup: %v", err)
		}

		// Then run periodically
		cleanupTicker := time.NewTicker(24 * time.Hour)
		defer cleanupTicker.Stop()
		for {
			select {
			case <-cleanupTicker.C:
				log.Println("Running periodic cleanup of old data...")
				if err := duckdbManager.DeleteOldData(retentionDays); err != nil {
					log.Printf("Error during periodic data cleanup: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Run a collection immediately at startup.
	compactionCounter := 0
	runCollection(ctx, connManager, collector, duckdbManager, planCollector, &compactionCounter)

	for {
		select {
		case <-ticker.C:
			runCollection(ctx, connManager, collector, duckdbManager, planCollector, &compactionCounter)
		case <-ctx.Done():
			log.Println("Collector stopped.")
			close(planIdentifierChan)
			wg.Wait()
			return
		}
	}
}

// runCollection performs one full collection and insertion cycle.
func runCollection(ctx context.Context, connMgr *sqldatacollector.ConnectionManager, coll *sqldatacollector.Collector, duckdbMgr *sqldatacollector.DuckDBManager, planColl *sqldatacollector.PlanCollector, compactionCounter *int) {
	log.Println("Running collection cycle...")

	// Get a valid connection for this cycle.
	manager, err := connMgr.GetConnection(ctx)
	if err != nil {
		log.Printf("Error getting connection: %v", err)
		return
	}

	results, err := coll.Collect(ctx, manager)
	if err != nil {
		log.Printf("Error during collection: %v", err)
		return
	}

	if len(results) > 0 {
		if err := duckdbMgr.InsertBatch(results); err != nil {
			log.Printf("Error inserting data into DuckDB: %v", err)
		} else {
			log.Printf("Successfully inserted %d records.", len(results))
			(*compactionCounter)++

			// Collect and store SQL plans asynchronously
			planColl.CollectAsync(results)
		}
	}

	if *compactionCounter >= CompactionThreshold {
		log.Println("Compaction threshold reached, running compaction...")
		if err := duckdbMgr.Compact(); err != nil {
			log.Printf("Error during compaction: %v", err)
		} else {
			*compactionCounter = 0
		}
	}
}

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, connMgr *sqldatacollector.ConnectionManager, tenantName string) (int64, error) {
	manager, err := connMgr.GetConnection(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection for tenant ID retrieval: %w", err)
	}
	var tenant sqldatacollector.Tenant
	err = manager.QueryRow(ctx, &tenant, "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?", tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}
