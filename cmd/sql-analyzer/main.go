package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	logger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	sqlanalyzer "github.com/oceanbase/ob-operator/internal/sql-analyzer"
	"github.com/oceanbase/ob-operator/pkg/log"
)

const (
	CompactionThreshold = 120
	PlanWorkerCount     = 4
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "log/sql-analyzer.log"
	}
	log.InitLogger(
		log.LoggerConfig{
			Level:      logLevel,
			Filename:   logFile,
			MaxSize:    256,
			MaxAge:     7,
			MaxBackups: 5,
			LocalTime:  true,
			Compress:   true,
		},
	)
}

func main() {

	// Init
	// launch collector
	// launch server

	// Configure logging
	logger.Info("SQL Data Collector starting...")

	// Read configuration from environment variables.
	obClusterName := os.Getenv("OB_CLUSTER_NAME")
	obClusterNamespace := os.Getenv("OB_CLUSTER_NAMESPACE")
	obTenant := os.Getenv("OB_TENANT")
	dataPath := os.Getenv("DATA_PATH")

	if obClusterName == "" || obClusterNamespace == "" || obTenant == "" {
		logger.Fatalf("OB_CLUSTER_NAME, OB_CLUSTER_NAMESPACE, and OB_TENANT environment variables must be set.")
	}
	if dataPath == "" {
		dataPath = "."
	}
	planDir := filepath.Join(dataPath, "sql_plan")
	if err := os.MkdirAll(planDir, 0755); err != nil {
		logger.Fatalf("Failed to create plan data directory: %v", err)
	}
	planDataDb := filepath.Join(planDir, "sql_plan.duckdb")

	// Create a Kubernetes client.
	k8sConfig, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("Failed to get Kubernetes config: %v", err)
	}
	v1alpha1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(k8sConfig, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		logger.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Println("Shutdown signal received, stopping collector...")
		cancel()
	}()

	// Get the OBCluster resource.
	obcluster := &v1alpha1.OBCluster{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: obClusterName, Namespace: obClusterNamespace}, obcluster); err != nil {
		logger.Fatalf("Failed to get OBCluster resource: %v", err)
	}

	// Create the connection manager.
	connManager := sqlanalyzer.NewConnectionManager(k8sClient, logr.FromContextOrDiscard(ctx), obcluster)
	defer connManager.Close()

	// Get an initial connection to retrieve the tenant ID.
	var obTenantID int64

	for {
		tenantID, err := getTenantIDByName(ctx, connManager, obTenant)
		if err != nil {
			logger.Printf("Failed to get tenant ID for tenant %s: %v. Retrying in 10 seconds...", obTenant, err)
		} else {
			obTenantID = tenantID
			logger.Printf("Found tenant '%s' with ID %d", obTenant, obTenantID)
			break // Success
		}

		// Wait before retrying or exit if context is cancelled.
		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			logger.Println("Collector stopped during tenant ID retrieval.")
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
			logger.Printf("Invalid COLLECTION_INTERVAL_SECONDS value '%s', using default of 30 seconds.", intervalStr)
		}
	}

	config := &sqlanalyzer.Config{
		Interval: time.Duration(intervalSeconds) * time.Second,
	}

	duckDBPath := filepath.Join(dataPath, "sql_audit")

	// Initialize the DuckDB manager.
	duckdbManager, err := sqlanalyzer.NewDuckDBManager(duckDBPath)
	if err != nil {
		logger.Fatalf("Failed to create DuckDB manager: %v", err)
	}
	defer duckdbManager.Close()

	// Initialize the PlanStore.
	planStore, err := sqlanalyzer.NewPlanStore(planDataDb)
	if err != nil {
		logger.Fatalf("Failed to create PlanStore: %v", err)
	}
	defer planStore.Close()

	// Create channels for plan collection
	planIdentifierChan := make(chan sqlanalyzer.PlanIdentifier, 100)

	// Initialize the PlanCollector.
	planCollector := sqlanalyzer.NewPlanCollector(planIdentifierChan)

	// Start plan workers
	var wg sync.WaitGroup
	wg.Add(PlanWorkerCount)
	for i := 0; i < PlanWorkerCount; i++ {
		worker := sqlanalyzer.NewPlanWorker(connManager, planIdentifierChan, planStore, &wg)
		go worker.Start(ctx)
	}

	// Retrieve the last known request IDs from DuckDB to resume progress.
	lastRequestIDs, err := duckdbManager.GetLastRequestIDs()
	if err != nil {
		logger.Fatalf("Failed to get last request IDs from DuckDB: %v", err)
	}
	logger.Printf("Retrieved progress for %d observers from DuckDB.", len(lastRequestIDs))

	// Initialize the OceanBase collector with the retrieved progress.
	collector := sqlanalyzer.NewCollector(config, obTenantID, lastRequestIDs)

	// Run the collection loop.
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Start the cleanup routine for old data
	retentionStr := os.Getenv("DATA_RETENTION_DAYS")
	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil {
		logger.Fatalf("Invalid or missing DATA_RETENTION_DAYS environment variable: %v", err)
	}

	go func() {
		// Run cleanup once at startup
		logger.Println("Running initial cleanup of old data...")
		if err := duckdbManager.DeleteOldData(retentionDays); err != nil {
			logger.Printf("Error during initial data cleanup: %v", err)
		}

		// Then run periodically
		cleanupTicker := time.NewTicker(24 * time.Hour)
		defer cleanupTicker.Stop()
		for {
			select {
			case <-cleanupTicker.C:
				logger.Println("Running periodic cleanup of old data...")
				if err := duckdbManager.DeleteOldData(retentionDays); err != nil {
					logger.Printf("Error during periodic data cleanup: %v", err)
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
			logger.Println("Collector stopped.")
			close(planIdentifierChan)
			wg.Wait()
			return
		}
	}
}

// runCollection performs one full collection and insertion cycle.
func runCollection(ctx context.Context, connMgr *sqlanalyzer.ConnectionManager, coll *sqlanalyzer.Collector, duckdbMgr *sqlanalyzer.DuckDBManager, planColl *sqlanalyzer.PlanCollector, compactionCounter *int) {
	logger.Println("Running collection cycle...")

	// Get a valid connection for this cycle.
	manager, err := connMgr.GetConnection(ctx)
	if err != nil {
		logger.Printf("Error getting connection: %v", err)
		return
	}

	results, err := coll.Collect(ctx, manager)
	if err != nil {
		logger.Printf("Error during collection: %v", err)
		return
	}

	if len(results) > 0 {
		if err := duckdbMgr.InsertBatch(results); err != nil {
			logger.Printf("Error inserting data into DuckDB: %v", err)
		} else {
			logger.Printf("Successfully inserted %d records.", len(results))
			(*compactionCounter)++

			// Collect and store SQL plans asynchronously
			planColl.CollectAsync(results)
		}
	}

	if *compactionCounter >= CompactionThreshold {
		logger.Println("Compaction threshold reached, running compaction...")
		if err := duckdbMgr.Compact(); err != nil {
			logger.Printf("Error during compaction: %v", err)
		} else {
			*compactionCounter = 0
		}
	}
}

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, connMgr *sqlanalyzer.ConnectionManager, tenantName string) (int64, error) {
	manager, err := connMgr.GetConnection(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection for tenant ID retrieval: %w", err)
	}
	var tenant sqlanalyzer.Tenant
	err = manager.QueryRow(ctx, &tenant, "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?", tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}
