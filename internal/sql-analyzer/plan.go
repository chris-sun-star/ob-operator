package sqlanalyzer

import (
	"context"
	"log"
	"sync"
)

// PlanIdentifier holds the identifiers for a plan.
type PlanIdentifier struct {
	TenantID uint64
	SvrIP    string
	SvrPort  int64
	PlanID   int64
}

// PlanCollector manages the collection of SQL plan data.
type PlanCollector struct {
	mu        sync.Mutex
	inputChan chan PlanIdentifier
}

// NewPlanCollector creates a new PlanCollector.
func NewPlanCollector(inputChan chan PlanIdentifier) *PlanCollector {
	return &PlanCollector{
		inputChan: inputChan,
	}
}

// CollectAsync sends new plan identifiers to the input channel.
func (c *PlanCollector) CollectAsync(audits []SQLAudit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("Collecting plans for %d audit records", len(audits))
	newPlanCount := 0

	for _, audit := range audits {
		c.inputChan <- PlanIdentifier{
			TenantID: audit.TenantId,
			SvrIP:    audit.SvrIP,
			SvrPort:  audit.SvrPort,
			PlanID:   audit.PlanId,
		}
		newPlanCount++
	}
	log.Printf("Found %d new plans", newPlanCount)
}

// PlanWorker fetches plan details from the database.
type PlanWorker struct {
	connManager *ConnectionManager
	inputChan   chan PlanIdentifier
	planStore   *PlanStore
	wg          *sync.WaitGroup
}

// NewPlanWorker creates a new PlanWorker.
func NewPlanWorker(connManager *ConnectionManager, inputChan chan PlanIdentifier, planStore *PlanStore, wg *sync.WaitGroup) *PlanWorker {
	return &PlanWorker{
		connManager: connManager,
		inputChan:   inputChan,
		planStore:   planStore,
		wg:          wg,
	}
}

// Start starts the worker.
func (w *PlanWorker) Start(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ident := <-w.inputChan:
			log.Printf("Fetching plan for tenant %d, server %s, port %d, plan %d", ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID)
			manager, err := w.connManager.GetConnection(ctx)
			if err != nil {
				log.Printf("failed to get connection for plan worker: %v", err)
				continue
			}
			query := `SELECT TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, SQL_ID, DB_ID, PLAN_HASH, GMT_CREATE, OPERATOR, OBJECT_NODE, OBJECT_ID, OBJECT_OWNER, OBJECT_NAME, OBJECT_ALIAS, OBJECT_TYPE, OPTIMIZER, ID, PARENT_ID, DEPTH, POSITION, COST, REAL_COST, CARDINALITY, REAL_CARDINALITY, IO_COST, CPU_COST, BYTES, ROWSET, OTHER_TAG, PARTITION_START, OTHER, ACCESS_PREDICATES, FILTER_PREDICATES, STARTUP_PREDICATES, PROJECTION, SPECIAL_PREDICATES, QBLOCK_NAME, REMARKS, OTHER_XML FROM GV$OB_SQL_PLAN WHERE TENANT_ID = ? AND SVR_IP = ? AND SVR_PORT = ? AND PLAN_ID = ?`
			var plans []SQLPlan
			if err := manager.QueryList(ctx, &plans, query, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID); err != nil {
				log.Printf("failed to query sql plan: %v", err)
				continue
			}
			log.Printf("Found %d plan details for tenant %d, server %s, port %d, plan %d", len(plans), ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID)
			for _, plan := range plans {
				if err := w.insertPlanIntoDuckDB(ctx, plan); err != nil {
					log.Printf("Error inserting plan into DuckDB: %v", err)
				}
			}
		}
	}
}

// insertPlanIntoDuckDB inserts a single SQLPlan into the DuckDB.
func (w *PlanWorker) insertPlanIntoDuckDB(ctx context.Context, plan SQLPlan) error {
	return w.planStore.Store(plan)
}
