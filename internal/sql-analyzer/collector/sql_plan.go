package collector

import (
	"context"
	"sync"

	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	logger "github.com/sirupsen/logrus"
)

type PlanWorker struct {
	collector   *Collector // New field
	connManager *oceanbase.ConnectionManager
	inputChan   chan *model.SqlPlanIdentifier
	planStore   *store.PlanStore
	wg          *sync.WaitGroup
}

// NewPlanWorker creates a new PlanWorker.
func NewPlanWorker(collector *Collector, connManager *oceanbase.ConnectionManager, inputChan chan *model.SqlPlanIdentifier, planStore *store.PlanStore, wg *sync.WaitGroup) *PlanWorker {
	return &PlanWorker{
		collector:   collector, // Assign the collector
		connManager: connManager,
		inputChan:   inputChan,
		planStore:   planStore,
		wg:          wg,
	}
}

// Start starts the worker.
func (w *PlanWorker) Start(ctx context.Context, idx int) {
	logger.Infof("plan worker %d started", idx)
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ident := <-w.inputChan:
			logger.Printf("Fetching plan for tenant %d, server %s, port %d, plan %d in worker %d", ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID, idx)
			cnx, err := w.connManager.GetConnection(ctx)
			if err != nil {
				logger.Printf("failed to get connection for plan worker: %v", err)
				// Remove from cache if failed
				w.collector.cacheMutex.Lock()
				w.collector.lruCache.Remove(*ident) // Remove from cache
				w.collector.cacheMutex.Unlock()
				continue
			}
			var plans []model.SqlPlan
			if err := cnx.QueryList(ctx, &plans, sqlconst.SelectSqlPlan, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID); err != nil {
				logger.Printf("failed to query sql plan: %v", err)
				// Remove from cache if failed
				w.collector.cacheMutex.Lock()
				w.collector.lruCache.Remove(*ident) // Remove from cache
				w.collector.cacheMutex.Unlock()
				continue
			}
			logger.Printf("Found %d plan details for tenant %d, server %s, port %d, plan %d", len(plans), ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID)
			allStored := true
			for _, plan := range plans {
				if err := w.planStore.Store(plan); err != nil {
					logger.Printf("Error inserting plan into DuckDB: %v", err)
					allStored = false
					break // Stop processing further plans for this identifier if one fails
				}
			}
			// Update cache status based on storage result
			w.collector.cacheMutex.Lock()
			if allStored {
				w.collector.lruCache.Add(*ident, struct{}{}) // Add to cache with empty struct
			} else {
				// If not all stored, remove from cache
				w.collector.lruCache.Remove(*ident) // Remove from cache
			}
			w.collector.cacheMutex.Unlock()
		}
	}
}
