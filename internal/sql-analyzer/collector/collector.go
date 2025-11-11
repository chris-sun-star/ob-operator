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
	CollectedSqlPlans  []model.SqlPlanIdentifier
	TenantID           uint64
	PlanIdentifierChan chan *model.SqlPlanIdentifier
}

// NewCollector creates a new Collector.
func NewCollector(ctx context.Context, config *config.Config) *Collector {
	c := &Collector{
		Ctx:                ctx,
		Config:             config,
		PlanIdentifierChan: make(chan *model.SqlPlanIdentifier, config.QueueSize),
	}
	return c
}

func (c *Collector) Init() {
	sqlAuditStore, err := store.NewSqlAuditStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_audit"))
	if err != nil {
		logger.Fatalf("Failed to initialize sql audit store: %v", err)
	}
	c.SqlAuditStore = sqlAuditStore

	planStore, err := store.NewPlanStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_plan"), false)
	if err != nil {
		logger.Fatalf("Failed to initialize sql plan store: %v", err)
	}
	c.SqlPlanStore = planStore

	// init duckdb if necessary
	obtenant, err := clients.GetOBTenant(c.Ctx, types.NamespacedName{
		Namespace: c.Config.Namespace,
		Name:      c.Config.OBTenant,
	})

	if err != nil {
		logger.Fatalf("Failed to get OBTenant resource: %v", err)
	}

	// Get the OBCluster resource.
	obcluster, err := clients.GetOBCluster(c.Ctx, c.Config.Namespace, obtenant.Spec.ClusterName)
	if err != nil {
		logger.Fatalf("Failed to get OBCluster resource: %v", err)
	}

	connectionManager := oceanbase.NewConnectionManager(logr.FromContextOrDiscard(c.Ctx), obcluster)
	c.ConnectionManager = connectionManager

	lastRequestIDs, err := sqlAuditStore.GetLastRequestIDs()
	if err != nil {
		logger.Fatalf("Failed to load request id from duckdb: %v", err)
	} else {
		for k, v := range lastRequestIDs {
			logger.Infof("Retrieved progress for %s with request id %d from DuckDB.", k, v)
		}
	}
	c.RequestIdMap = lastRequestIDs

	// what if there's a huge number of plans
	collectedSqlPlans, err := planStore.LoadExistingPlans()
	if err != nil {
		logger.Fatalf("Failed to load plan identities from duckdb: %v", err)
	}
	c.CollectedSqlPlans = collectedSqlPlans

	tenantID, err := getTenantIDByName(c.Ctx, connectionManager, obtenant.Spec.TenantName)
	if err != nil {
		logger.Fatalf("Failed to get tenant id from oceanbase: %v", err)
	}
	c.TenantID = tenantID

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

	// Run the collection loop.
	ticker := time.NewTicker(c.Config.Interval)
	defer ticker.Stop()

	// Run a collection immediately at startup.
	compactionCounter := 0

	for {
		select {
		case <-ticker.C:
			c.collectSqlAuditData()
			if compactionCounter%parquet.CompactionThreshold == 0 {
				// TODO make compaction async
				err := c.SqlAuditStore.Compact()
				if err != nil {
					logger.Errorf("Failed to compact sql audit data %v", err)
				}
			}
			compactionCounter = (compactionCounter + 1) % parquet.CompactionThreshold
		case <-c.Ctx.Done():
			logger.Println("Collector stopped.")
			close(c.PlanIdentifierChan)
			wg.Wait()
			return
		}
	}
}
