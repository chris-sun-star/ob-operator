package collector

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-logr/logr"
	logger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/const/parquet"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

type Collector struct {
	Ctx               context.Context
	Config            *config.Config
	ConnectionManager *oceanbase.ConnectionManager
	SqlAuditStore     *store.SqlAuditStore
	SqlPlanStore      *store.PlanStore
	RequestIdMap      map[string]uint64

	//TODO This should be a cache with limited number
	CollectedSqlPlans  map[model.SqlPlanIdentifier]struct{}
	TenantID           uint64
	PlanIdentifierChan chan *model.SqlPlanIdentifier
	CompactionChan     chan struct{}
}

// NewCollector creates a new Collector.
func NewCollector(ctx context.Context, config *config.Config) *Collector {
	c := &Collector{
		Ctx:                ctx,
		Config:             config,
		PlanIdentifierChan: make(chan *model.SqlPlanIdentifier, config.QueueSize),
		CollectedSqlPlans:  make(map[model.SqlPlanIdentifier]struct{}),
		CompactionChan:     make(chan struct{}, 1),
	}
	return c
}

func (c *Collector) Init() error {
	sqlAuditStore, err := store.NewSqlAuditStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_audit"))
	if err != nil {
		return fmt.Errorf("failed to initialize sql audit store: %w", err)
	}
	c.SqlAuditStore = sqlAuditStore

	planStore, err := store.NewPlanStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_plan"), false)
	if err != nil {
		return fmt.Errorf("failed to initialize sql plan store: %w", err)
	}
	c.SqlPlanStore = planStore

	// init duckdb if necessary
	obtenant, err := clients.GetOBTenant(c.Ctx, types.NamespacedName{
		Namespace: c.Config.Namespace,
		Name:      c.Config.OBTenant,
	})

	if err != nil {
		return fmt.Errorf("failed to get OBTenant resource: %w", err)
	}

	// Get the OBCluster resource.
	obcluster, err := clients.GetOBCluster(c.Ctx, c.Config.Namespace, obtenant.Spec.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to get OBCluster resource: %w", err)
	}

	connectionManager := oceanbase.NewConnectionManager(logger.StandardLogger(), obcluster)
	c.ConnectionManager = connectionManager

	lastRequestIDs, err := sqlAuditStore.GetLastRequestIDs()
	if err != nil {
		return fmt.Errorf("failed to load request id from duckdb: %w", err)
	} else {
		for k, v := range lastRequestIDs {
			logger.Infof("Retrieved progress for %s with request id %d from DuckDB.", k, v)
		}
	}
	c.RequestIdMap = lastRequestIDs

	// what if there's a huge number of plans
	existingPlans, err := planStore.LoadExistingPlans()
	if err != nil {
		return fmt.Errorf("failed to load plan identities from duckdb: %w", err)
	}
	for _, plan := range existingPlans {
		c.CollectedSqlPlans[plan] = struct{}{}
	}

	tenantID, err := getTenantIDByName(c.Ctx, connectionManager, obtenant.Spec.TenantName)
	if err != nil {
		return fmt.Errorf("failed to get tenant id from oceanbase: %w", err)
	}
	c.TenantID = tenantID
	return nil
}

func (c *Collector) Stop() {
	defer c.SqlAuditStore.Close()
	defer c.SqlPlanStore.Close()
	defer c.ConnectionManager.Close()
}

func (c *Collector) Start() {
	var wg sync.WaitGroup
	wg.Add(c.Config.WorkerNum)

	for i := 0; i < c.Config.WorkerNum; i++ {
		worker := NewPlanWorker(c.ConnectionManager, c.PlanIdentifierChan, c.SqlPlanStore, &wg)
		go worker.Start(c.Ctx, i)
	}

	// Start compaction worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-c.CompactionChan:
				logger.Println("Compaction signal received, running compaction...")
				if err := c.SqlAuditStore.Compact(); err != nil {
					logger.Errorf("Failed to compact sql audit data: %v", err)
				} else {
					logger.Println("Sql audit data compacted successfully.")
				}
			case <-c.Ctx.Done():
				logger.Println("Compaction worker stopped.")
				return
			}
		}
	}()

	// Run the collection loop.
	ticker := time.NewTicker(c.Config.Interval)
	defer ticker.Stop()

	// Run a collection immediately at startup.
	compactionCounter := 0

	for {
		select {
		case <-ticker.C:
			c.collectSqlAuditData()
			compactionCounter++
			if compactionCounter >= parquet.CompactionThreshold {
				select {
				case c.CompactionChan <- struct{}{}: // Send compaction signal
					compactionCounter = 0 // Reset counter after sending signal
				default:
					logger.Warn("Compaction channel is full, skipping compaction signal.")
				}
			}
		case <-c.Ctx.Done():
			logger.Println("Collector stopped. Stopping plan workers...")
			close(c.PlanIdentifierChan)
			wg.Wait() // Wait for all workers (plan and compaction) to finish
			return
		}
	}
}
