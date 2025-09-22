package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	k8sClient          client.Client
	logger             logr.Logger
	obcluster          *v1alpha1.OBCluster
	cachedConnection   *operation.OceanbaseOperationManager
	mu                 sync.Mutex
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
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	log.Println("SQL Data Collector starting...")

	// Read configuration from environment variables.
	obClusterName := os.Getenv("OB_CLUSTER_NAME")
	obClusterNamespace := os.Getenv("OB_CLUSTER_NAMESPACE")
	obTenant := os.Getenv("OB_TENANT")

	if obClusterName == "" || obClusterNamespace == "" || obTenant == "" {
		log.Fatal("OB_CLUSTER_NAME, OB_CLUSTER_NAMESPACE, and OB_TENANT environment variables must be set.")
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
	initialManager, err := connManager.GetConnection(ctx)
	if err != nil {
		log.Fatalf("Failed to get initial OceanBase connection: %v", err)
	}

	obTenantID, err := getTenantIDByName(ctx, initialManager, obTenant)
	if err != nil {
		log.Fatalf("Failed to get tenant ID for tenant %s: %v", obTenant, err)
	}
	log.Printf("Found tenant '%s' with ID %d", obTenant, obTenantID)

	config := &sqldatacollector.Config{
		Interval: 30 * time.Second,
	}

	duckDBPath := fmt.Sprintf("sql_audit_tenant_%s.duckdb", obTenant)

	// Initialize the OceanBase collector.
	collector := sqldatacollector.NewCollector(config, obTenantID)

	// Initialize the DuckDB manager.
	duckdbManager, err := sqldatacollector.NewDuckDBManager(duckDBPath)
	if err != nil {
		log.Fatalf("Failed to create DuckDB manager: %v", err)
	}
	defer duckdbManager.Close()

	// Run the collection loop.
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

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