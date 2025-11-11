package oceanbase

import (
	"context"
	"fmt"
	"sync"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/logr_adapter" // New import
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	logger "github.com/sirupsen/logrus"
)

// ConnectionManager handles the connection to the OceanBase cluster.
type ConnectionManager struct {
	logger           logger.FieldLogger
	obcluster        *v1alpha1.OBCluster
	cachedConnection *operation.OceanbaseOperationManager
	mu               sync.Mutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(log logger.FieldLogger, obcluster *v1alpha1.OBCluster) *ConnectionManager {
	return &ConnectionManager{
		logger:    log,
		obcluster: obcluster,
	}
}

// GetConnection returns a valid OceanBaseOperationManager, handling reconnection if necessary.
func (cm *ConnectionManager) GetConnection(ctx context.Context) (*operation.OceanbaseOperationManager, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cachedConnection != nil && cm.cachedConnection.Connector.IsAlive() {
		cm.logger.Println("Using cached connection.")
		return cm.cachedConnection, nil
	}

	cm.logger.Println("Cached connection is not alive, creating a new one...")
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}

	// create a k8s client
	// Wrap cm.logger with the adapter
	logrLogger := logr_adapter.NewLogrAdapter(cm.logger)
	manager, err := utils.GetSysOperationClient(nil, &logrLogger, cm.obcluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get OceanBase operation manager: %w", err)
	}

	cm.cachedConnection = manager
	cm.logger.Println("Successfully created a new connection.")
	return cm.cachedConnection, nil
}

// Close closes the cached connection.
func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.logger.Println("Closing ConnectionManager")
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}
}
