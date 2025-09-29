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

	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/internal/sql-data-collector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// ConnectionManager handles the connection to the OceanBase cluster.
type ConnectionManager struct {
	k8sClient        client.Client
	logger           logr.Logger
	obcluster        *v1alpha1.OBCluster
	cachedConnection *operation.OceanbaseOperationManager
	mu               sync.Mutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(k8sClient client.Client, logger logr.Logger, obcluster *v1alpha1.OBCluster) *ConnectionManager {
	return &ConnectionManager{
		k8sClient: k8sClient,
		logger:    logger,
		obcluster: obcluster,
	}
}

// GetConnection returns a valid OceanBaseOperationManager, handling reconnection if necessary.
func (cm *ConnectionManager) GetConnection(ctx context.Context) (*operation.OceanbaseOperationManager, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cachedConnection != nil && cm.cachedConnection.Connector.IsAlive() {
		log.Println("Using cached connection.")
		return cm.cachedConnection, nil
	}

	log.Println("Cached connection is not alive, creating a new one...")
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}

	manager, err := utils.GetSysOperationClient(cm.k8sClient, &cm.logger, cm.obcluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get OceanBase operation manager: %w", err)
	}

	cm.cachedConnection = manager
	log.Println("Successfully created a new connection.")
	return cm.cachedConnection, nil
}

// Close closes the cached connection.
func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}
}

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
	connManager := NewConnectionManager(k8sClient, logger, obcluster)
	defer connManager.Close()

	// Get an initial connection to retrieve the tenant ID.
	var obTenantID int64
	var initialManager *operation.OceanbaseOperationManager

	for {
		var err error
		initialManager, err = connManager.GetConnection(ctx)
		if err != nil {
			log.Printf("Failed to get OceanBase connection: %v. Retrying in 10 seconds...", err)
		} else {
			tenantID, err := getTenantIDByName(ctx, initialManager, obTenant)
			if err == nil {
				obTenantID = tenantID
				log.Printf("Found tenant '%s' with ID %d", obTenant, obTenantID)
				break // Success
			}
			log.Printf("Failed to get tenant ID for tenant %s: %v. Retrying in 10 seconds...", obTenant, err)
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
	runCollection(ctx, connManager, collector, duckdbManager)

	for {
		select {
		case <-ticker.C:
			runCollection(ctx, connManager, collector, duckdbManager)
		case <-ctx.Done():
			log.Println("Collector stopped.")
			return
		}
	}
}

// runCollection performs one full collection and insertion cycle.
func runCollection(ctx context.Context, connMgr *ConnectionManager, coll *sqldatacollector.Collector, duckdbMgr *sqldatacollector.DuckDBManager) {
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
		}
	}
}

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, manager *operation.OceanbaseOperationManager, tenantName string) (int64, error) {
	var tenant sqldatacollector.Tenant
	err := manager.QueryRow(ctx, &tenant, "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?", tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}
