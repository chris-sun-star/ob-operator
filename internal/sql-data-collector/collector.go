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

// collectFromObserver collects aggregated SQL audit data from a specific observer.
func (c *Collector) collectFromObserver(ctx context.Context, manager *operation.OceanbaseOperationManager, svrIP string, lastRequestID uint64) ([]SQLAudit, error) {
	query := `
		SELECT
			-- Grouping Keys
			svr_ip, tenant_id, tenant_name, user_id, user_name, db_id, db_name, sql_id, plan_id,

			-- Aggregated String/Identifier Values (using MAX to select one value from the group)
			MAX(query_sql) as query_sql,
			MAX(client_ip) as client_ip,
			MAX(event) as event,
			MAX(plan_type) as plan_type,
			MAX(consistency_level) as consistency_level,

			-- Aggregated Numeric Values
			COUNT(*) as executions,
			MIN(request_time) as min_request_time,
			MAX(request_time) as max_request_time,

			SUM(elapsed_time) as elapsed_time_sum, MAX(elapsed_time) as elapsed_time_max, MIN(elapsed_time) as elapsed_time_min,
			SUM(execute_time) as execute_time_sum, MAX(execute_time) as execute_time_max, MIN(execute_time) as execute_time_min,
			SUM(queue_time) as queue_time_sum, MAX(queue_time) as queue_time_max, MIN(queue_time) as queue_time_min,
			SUM(get_plan_time) as get_plan_time_sum, MAX(get_plan_time) as get_plan_time_max, MIN(get_plan_time) as get_plan_time_min,
			SUM(affected_rows) as affected_rows_sum, MAX(affected_rows) as affected_rows_max, MIN(affected_rows) as affected_rows_min,
			SUM(return_rows) as return_rows_sum, MAX(return_rows) as return_rows_max, MIN(return_rows) as return_rows_min,
			SUM(partition_cnt) as partition_count_sum, MAX(partition_cnt) as partition_count_max, MIN(partition_cnt) as partition_count_min,
			SUM(retry_cnt) as retry_count_sum, MAX(retry_cnt) as retry_count_max, MIN(retry_cnt) as retry_count_min,
			SUM(disk_reads) as disk_reads_sum, MAX(disk_reads) as disk_reads_max, MIN(disk_reads) as disk_reads_min,
			SUM(rpc_count) as rpc_count_sum, MAX(rpc_count) as rpc_count_max, MIN(rpc_count) as rpc_count_min,
			SUM(memstore_read_row_count) as memstore_read_row_count_sum, MAX(memstore_read_row_count) as memstore_read_row_count_max, MIN(memstore_read_row_count) as memstore_read_row_count_min,
			SUM(ssstore_read_row_count) as ssstore_read_row_count_sum, MAX(ssstore_read_row_count) as ssstore_read_row_count_max, MIN(ssstore_read_row_count) as ssstore_read_row_count_min,
			SUM(request_memory_used) as request_memory_used_sum, MAX(request_memory_used) as request_memory_used_max, MIN(request_memory_used) as request_memory_used_min,
			SUM(CASE WHEN ret_code = 0 THEN 0 ELSE 1 END) as fail_count_sum,

			SUM(CASE WHEN ret_code = -4012 THEN 1 ELSE 0 END) as ret_code_4012_count_sum,
			SUM(CASE WHEN ret_code = -4013 THEN 1 ELSE 0 END) as ret_code_4013_count_sum,
			SUM(CASE WHEN ret_code = -5001 THEN 1 ELSE 0 END) as ret_code_5001_count_sum,
			SUM(CASE WHEN ret_code = -5024 THEN 1 ELSE 0 END) as ret_code_5024_count_sum,
			SUM(CASE WHEN ret_code = -5167 THEN 1 ELSE 0 END) as ret_code_5167_count_sum,
			SUM(CASE WHEN ret_code = -5217 THEN 1 ELSE 0 END) as ret_code_5217_count_sum,
			SUM(CASE WHEN ret_code = -6002 THEN 1 ELSE 0 END) as ret_code_6002_count_sum,

			SUM(CASE event WHEN 'system internal wait' THEN wait_time_micro ELSE 0 END) as event_0_wait_time_sum,
			SUM(CASE event WHEN 'mysql response wait client' THEN wait_time_micro ELSE 0 END) as event_1_wait_time_sum,
			SUM(CASE event WHEN 'sync rpc' THEN wait_time_micro ELSE 0 END) as event_2_wait_time_sum,
			SUM(CASE event WHEN 'db file data read' THEN wait_time_micro ELSE 0 END) as event_3_wait_time_sum

		FROM gv$ob_sql_audit
		WHERE tenant_id = ? AND svr_ip = ? AND request_id > ?
		GROUP BY
			svr_ip, tenant_id, tenant_name, user_id, user_name, db_id, db_name, sql_id, plan_id
	`

	var results []SQLAudit
	if err := manager.QueryList(ctx, &results, query, c.tenantID, svrIP, lastRequestID); err != nil {
		return nil, err
	}
	return results, nil
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

		if ok && lastRequestID >= maxRequestID {
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

	if err := manager.QueryList(ctx, &observers, query, c.tenantID); err != nil {
		return nil, err
	}

	maxRequestIDs := make(map[string]uint64)
	for _, o := range observers {
		maxRequestIDs[o.SvrIP] = o.MaxRequest
	}
	return maxRequestIDs, nil
}