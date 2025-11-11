package oceanbase

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr" // Use logr directly
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// ConnectionManager handles the connection to the OceanBase cluster.
type ConnectionManager struct {
	logger           logr.Logger // Change to logr.Logger
	obcluster        *v1alpha1.OBCluster
	cachedConnection *operation.OceanbaseOperationManager
	mu               sync.Mutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(logger logr.Logger, obcluster *v1alpha1.OBCluster) *ConnectionManager { // Change to logr.Logger
	return &ConnectionManager{
		logger:    logger,
		obcluster: obcluster,
	}
}

// GetConnection returns a valid OceanBaseOperationManager, handling reconnection if necessary.
func (cm *ConnectionManager) GetConnection(ctx context.Context) (*operation.OceanbaseOperationManager, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cachedConnection != nil && cm.cachedConnection.Connector.IsAlive() {
		cm.logger.Info("Using cached connection.") // Use logr.Logger.Info
		return cm.cachedConnection, nil
	}

	cm.logger.Info("Cached connection is not alive, creating a new one...") // Use logr.Logger.Info
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}

	// create a k8s client
	// TODO replace nil with a new k8s client
	manager, err := utils.GetSysOperationClient(nil, &cm.logger, cm.obcluster) // Pass &cm.logger directly
	if err != nil {
		return nil, fmt.Errorf("failed to get OceanBase operation manager: %w", err)
	}

	cm.cachedConnection = manager
	cm.logger.Info("Successfully created a new connection.") // Use logr.Logger.Info
	return cm.cachedConnection, nil
}

// Close closes the cached connection.
func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.logger.Info("Closing ConnectionManager") // Use logr.Logger.Info
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}
}
