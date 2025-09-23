package sqldatacollector

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// Config holds the configuration for the collector.
type Config struct {
	Interval time.Duration
}

// Collector manages the collection of SQL audit data.
type Collector struct {
	config         *Config
	tenantID       int64
	lastRequestIDs map[string]uint64
	mu             sync.Mutex
}

// NewCollector creates a new Collector.
func NewCollector(config *Config, tenantID int64, initialRequestIDs map[string]uint64) *Collector {
	c := &Collector{
		config:         config,
		tenantID:       tenantID,
		lastRequestIDs: initialRequestIDs,
		mu:             sync.Mutex{},
	}
	if c.lastRequestIDs == nil {
		c.lastRequestIDs = make(map[string]uint64)
	}
	return c
}

// Collect fetches new SQL audit data from the OceanBase cluster.
func (c *Collector) Collect(ctx context.Context, manager *operation.OceanbaseOperationManager) ([]SQLAudit, error) {
	// Step 1: Find observers with new data.
	maxRequestIDs, err := c.getMaxRequestIDs(ctx, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to get max request IDs: %w", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []SQLAudit, len(maxRequestIDs))
	errChan := make(chan error, len(maxRequestIDs))

	// Step 2: For each observer with new data, dispatch a collection goroutine.
	for svrIP, maxRequestID := range maxRequestIDs {
		c.mu.Lock()
		lastRequestID, ok := c.lastRequestIDs[svrIP]
		c.mu.Unlock()

		// If we have a last request ID and it's the same as the current max, skip.
		// We use != instead of > to handle observer restarts where request_id may reset.
		if ok && lastRequestID == maxRequestID {
			continue
		}

		wg.Add(1)
		go func(svrIP string, lastID uint64) {
			defer wg.Done()
			log.Printf("Collecting from observer %s since request_id %d", svrIP, lastID)
			data, err := c.collectFromObserver(ctx, manager, svrIP, lastID)
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
	var allResults []SQLAudit
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	for err := range errChan {
		log.Println("Error during collection:", err) // Log errors but don't fail the whole batch
	}

	// Step 3: Update the last request IDs for the next cycle.
	c.mu.Lock()
	for svrIP, maxRequestID := range maxRequestIDs {
		c.lastRequestIDs[svrIP] = maxRequestID
	}
	c.mu.Unlock()

	log.Printf("Collected %d new audit records.", len(allResults))
	return allResults, nil
}

// getMaxRequestIDs finds the latest request_id for each observer.
func (c *Collector) getMaxRequestIDs(ctx context.Context, manager *operation.OceanbaseOperationManager) (map[string]uint64, error) {
	query := "SELECT svr_ip, MAX(request_id) FROM gv$ob_sql_audit WHERE tenant_id = ? GROUP BY svr_ip"
	var observers []struct {
		SvrIP      string `db:"svr_ip"`
		MaxRequest uint64 `db:"MAX(request_id)"`
	}

	if err := manager.Query(ctx, &observers, query, c.tenantID); err != nil {
		return nil, err
	}

	maxRequestIDs := make(map[string]uint64)
	for _, o := range observers {
		maxRequestIDs[o.SvrIP] = o.MaxRequest
	}
	return maxRequestIDs, nil
}

// collectFromObserver collects aggregated SQL audit data from a specific observer.
func (c *Collector) collectFromObserver(ctx context.Context, manager *operation.OceanbaseOperationManager, svrIP string, lastRequestID uint64) ([]SQLAudit, error) {
	query := `
		SELECT
			svr_ip, tenant_id, tenant_name, user_name, database_name, sql_id, query_sql, plan_id,
			MAX(affected_rows) as affected_rows, MAX(return_rows) as return_rows, MAX(ret_code) as ret_code,
			MAX(request_id) as request_id, MAX(request_time) as request_time, MAX(elapsed_time) as elapsed_time,
			MAX(execute_time) as execute_time, MAX(queue_time) as queue_time
		FROM gv$ob_sql_audit
		WHERE tenant_id = ? AND svr_ip = ? AND request_id > ?
		GROUP BY svr_ip, tenant_id, tenant_name, user_name, database_name, sql_id, query_sql, plan_id
	`

	var results []SQLAudit
	if err := manager.Query(ctx, &results, query, c.tenantID, svrIP, lastRequestID); err != nil {
		return nil, err
	}
	return results, nil
}
