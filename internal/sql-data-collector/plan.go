package sqldatacollector

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// PlanIdentifier holds the identifiers for a plan.
type PlanIdentifier struct {
	TenantID uint64
	SvrIP    string
	PlanID   int64
}

// PlanCollector manages the collection of SQL plan data.
type PlanCollector struct {
	existingPlans map[string]bool
	mu            sync.Mutex
	inputChan     chan PlanIdentifier
}

// NewPlanCollector creates a new PlanCollector.
func NewPlanCollector(existingPlans map[string]bool, inputChan chan PlanIdentifier) *PlanCollector {
	return &PlanCollector{
		existingPlans: existingPlans,
		inputChan:     inputChan,
	}
}

// CollectAsync sends new plan identifiers to the input channel.
func (c *PlanCollector) CollectAsync(audits []SQLAudit) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, audit := range audits {
		key := fmt.Sprintf("%d-%s-%d", audit.TenantId, audit.SvrIP, audit.PlanId)
		if !c.existingPlans[key] {
			c.existingPlans[key] = true
			c.inputChan <- PlanIdentifier{
				TenantID: audit.TenantId,
				SvrIP:    audit.SvrIP,
				PlanID:   audit.PlanId,
			}
		}
	}
}

// PlanWorker fetches plan details from the database.
type PlanWorker struct {
	manager   *operation.OceanbaseOperationManager
	inputChan chan PlanIdentifier
	outputChan chan SQLPlan
	wg        *sync.WaitGroup
}

// NewPlanWorker creates a new PlanWorker.
func NewPlanWorker(manager *operation.OceanbaseOperationManager, inputChan chan PlanIdentifier, outputChan chan SQLPlan, wg *sync.WaitGroup) *PlanWorker {
	return &PlanWorker{
		manager:   manager,
		inputChan: inputChan,
		outputChan: outputChan,
		wg:        wg,
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
			query := `SELECT TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, SQL_ID, DB_ID, PLAN_HASH, GMT_CREATE, OPERATOR, OBJECT_NODE, OBJECT_ID, OBJECT_OWNER, OBJECT_NAME, OBJECT_ALIAS, OBJECT_TYPE, OPTIMIZER, ID, PARENT_ID, DEPTH, POSITION, COST, REAL_COST, CARDINALITY, REAL_CARDINALITY, IO_COST, CPU_COST, BYTES, ROWSET, OTHER_TAG, PARTITION_START, OTHER, ACCESS_PREDICATES, FILTER_PREDICATES, STARTUP_PREDICATES, PROJECTION, SPECIAL_PREDICATES, QBLOCK_NAME, REMARKS, OTHER_XML FROM GV$OB_SQL_PLAN WHERE TENANT_ID = ? AND SVR_IP = ? AND PLAN_ID = ?`
			var plans []SQLPlan
			if err := w.manager.QueryList(ctx, &plans, query, ident.TenantID, ident.SvrIP, ident.PlanID); err != nil {
				log.Printf("failed to query sql plan: %v", err)
				continue
			}
			for _, plan := range plans {
				w.outputChan <- plan
			}
		}
	}
}