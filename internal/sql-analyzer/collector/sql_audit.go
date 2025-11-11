package collector

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"

	logger "github.com/sirupsen/logrus"
)

// getMaxRequestIDs finds the latest request_id for each observer.
func (c *Collector) getMaxRequestIDs() (map[string]uint64, error) {
	var observers []struct {
		SvrIP        string `db:"svr_ip"`
		MaxRequestID uint64 `db:"max_request_id"`
	}

	cnx, err := c.ConnectionManager.GetConnection(c.Ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get oceanbase connection")
	}

	if err := cnx.QueryList(c.Ctx, &observers, sqlconst.GetMaxRequestIDByIP, c.TenantID); err != nil {
		return nil, errors.Wrap(err, "Failed to query max request ids")
	}

	maxRequestIDs := make(map[string]uint64)
	for _, o := range observers {
		maxRequestIDs[o.SvrIP] = o.MaxRequestID
	}
	return maxRequestIDs, nil
}

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func (c *Collector) collectSqlAuditData() {
	// Step 1: Find observers with new data.
	maxRequestIDs, err := c.getMaxRequestIDs()
	if err != nil {
		logger.Errorf("Failed to get max request ids %v", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []model.SqlAudit, len(maxRequestIDs))
	errChan := make(chan error, len(maxRequestIDs))

	// Step 2: For each observer with new data, dispatch a collection goroutine.
	for svrIP, maxRequestID := range maxRequestIDs {
		lastRequestID, ok := c.RequestIdMap[svrIP]
		if ok && lastRequestID == maxRequestID {
			continue
		}

		wg.Add(1)
		go func(svrIP string, lastRequestID uint64) {
			defer wg.Done()
			logger.Printf("Collecting from observer %s since request_id %d", svrIP, lastRequestID)
			data, err := c.collectSqlAuditByOBServer(svrIP, lastRequestID)
			if err != nil {
				errChan <- fmt.Errorf("failed to collect from observer %s: %w", svrIP, err)
				return
			}
			if len(data) > 0 {
				resultsChan <- data
			}
		}(svrIP, lastRequestID)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	// Consolidate results and errors.
	var allResults []model.SqlAudit
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	for err := range errChan {
		logger.Println("Error during collection:", err) // Log errors but don't fail the whole batch
	}

	// Step 3: Update the last request IDs for the next cycle.
	for _, audit := range allResults {
		c.PushPlan(&model.SqlPlanIdentifier{
			TenantID: c.TenantID,
			SvrIP:    audit.SvrIP,
			SvrPort:  audit.SvrPort,
			PlanID:   audit.PlanId,
		})
		lastRequestID, ok := c.RequestIdMap[audit.SvrIP]
		if !ok || lastRequestID < audit.MaxRequestId {
			c.RequestIdMap[audit.SvrIP] = audit.MaxRequestId
		}
	}
	logger.Printf("Collected %d new audit records.", len(allResults))

	// TODO persist sql audit data and send plan identities to plan channel
	if len(allResults) > 0 {
		if err := c.SqlAuditStore.InsertBatch(allResults); err != nil {
			logger.Printf("Error inserting data into DuckDB: %v", err)
		} else {
			logger.Printf("Saved %d sql audit records", len(allResults))
		}
	}

}

func (c *Collector) PushPlan(plan *model.SqlPlanIdentifier) {
	if _, ok := c.CollectedSqlPlans[*plan]; ok {
		logger.Debugf("Plan %v already collected, skipping.", plan)
		return
	}
	c.CollectedSqlPlans[*plan] = struct{}{}
	c.PlanIdentifierChan <- plan
}

func (c *Collector) collectSqlAuditByOBServer(svrIP string, lastRequestID uint64) ([]model.SqlAudit, error) {
	var results []model.SqlAudit
	cnx, err := c.ConnectionManager.GetConnection(c.Ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get oceanbase connection")
	}

	if err := cnx.QueryList(c.Ctx, &results, sqlconst.GetSqlStatistics, c.TenantID, svrIP, lastRequestID); err != nil {
		return nil, err
	}
	return results, nil
}
