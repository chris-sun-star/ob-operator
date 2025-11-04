package sqlanalyzer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	log.Println("Closing ConnectionManager")
	if cm.cachedConnection != nil {
		cm.cachedConnection.Close()
	}
}
