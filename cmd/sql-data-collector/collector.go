package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// Config holds the configuration for the collector.
type Config struct {
	Interval time.Duration `yaml:"interval"`
}

// Collector manages the data collection from OceanBase.
type Collector struct {
	config       *Config
	tenantID     int64
	requestIdMap sync.Map
}

// Tenant represents an OceanBase tenant, used for discovery.
type Tenant struct {
	ID   int64  `db:"tenant_id"`
	Name string `db:"tenant_name"`
}

// ObserverStat holds the latest request_id for an observer.
type ObserverStat struct {
	SvrIP        string `db:"svr_ip"`
	MaxRequestId int64  `db:"max_request_id"`
}

// NewCollector creates a new Collector instance.
func NewCollector(config *Config, tenantID int64) *Collector {
	return &Collector{
		config:   config,
		tenantID: tenantID,
	}
}

// Collect performs a single collection cycle for the configured tenant.
func (c *Collector) Collect(ctx context.Context, manager *operation.OceanbaseOperationManager) ([]SqlAuditResult, error) {
	log.Println("Starting SQL audit collection...")

	observerStats, err := c.getObserverStats(ctx, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to get observer stats: %w", err)
	}

	log.Printf("Found %d observers with data for tenant %d.", len(observerStats), c.tenantID)

	var allResults []SqlAuditResult
	var wg sync.WaitGroup
	resultChan := make(chan []SqlAuditResult, len(observerStats))

	for _, stat := range observerStats {
		key := fmt.Sprintf("%s-%d", stat.SvrIP, c.tenantID)
		lastMaxRequestId, _ := c.requestIdMap.Load(key)
		if lastMaxRequestId == nil {
			lastMaxRequestId = int64(0)
		}

		if stat.MaxRequestId != lastMaxRequestId.(int64) {
			wg.Add(1)
			go func(s ObserverStat) {
				defer wg.Done()
				startId := lastMaxRequestId.(int64) + 1
				// Handle observer restart
				if s.MaxRequestId < lastMaxRequestId.(int64) {
					log.Printf("Observer restart detected for %s. Resetting start request_id.", key)
					startId = 0 // Reset to collect from the beginning
				}

				log.Printf("Collecting for server %s, tenant %d", s.SvrIP, c.tenantID)
				results, err := c.collectAndAggregate(ctx, manager, s.SvrIP, c.tenantID, startId, s.MaxRequestId)
				if err != nil {
					log.Printf("Failed to collect for server %s, tenant %d: %v", s.SvrIP, c.tenantID, err)
					return
				}
				if len(results) > 0 {
					resultChan <- results
				}
			}(stat)
		} else {
			log.Printf("No new data for server %s, tenant %d. Skipping.", stat.SvrIP, c.tenantID)
		}
	}

	wg.Wait()
	close(resultChan)

	for res := range resultChan {
		allResults = append(allResults, res...)
	}

	log.Printf("Collection finished. Aggregated %d SQL statements in total.", len(allResults))
	return allResults, nil
}

func (c *Collector) getObserverStats(ctx context.Context, manager *operation.OceanbaseOperationManager) ([]ObserverStat, error) {
	var stats []ObserverStat
	query := fmt.Sprintf(selectObserverStats, globalSqlAuditTableName)
	err := manager.Connector.GetClient().SelectContext(ctx, &stats, query, c.tenantID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (c *Collector) collectAndAggregate(ctx context.Context, manager *operation.OceanbaseOperationManager, svrIP string, tenantID int64, startRequestId, maxRequestId int64) ([]SqlAuditResult, error) {
	key := fmt.Sprintf("%s-%d", svrIP, tenantID)

	var results []SqlAuditResult
	query := fmt.Sprintf(aggregatedSqlAuditQuery, time.Now().UnixMicro(), globalSqlAuditTableName)
	args := []interface{}{svrIP, tenantID, startRequestId, maxRequestId}

	err := manager.Connector.GetClient().SelectContext(ctx, &results, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated sql audit: %w", err)
	}

	var newMaxRequestId int64
	for _, res := range results {
		if res.MaxRequestId > newMaxRequestId {
			newMaxRequestId = res.MaxRequestId
		}
	}

	if newMaxRequestId > 0 {
		c.requestIdMap.Store(key, newMaxRequestId)
	}

	return results, nil
}

// --- Data Structures ---

type SqlAuditResult struct {
	CollectTime          int64  `db:"collect_time"`
	SvrIp                string `db:"svr_ip"`
	TenantId             uint64 `db:"tenant_id"`
	UserId               int64  `db:"user_id"`
	DbId                 uint64 `db:"db_id"`
	SqlId                string `db:"sql_id"`
	MaxRequestId         int64  `db:"max_request_id"`
	MinRequestTime       int64  `db:"min_request_time"`
	MaxRequestTime       int64  `db:"max_request_time"`
	Executions           int64  `db:"executions"`
	AffectedRows         int64  `db:"affected_rows"`
	ReturnRows           int64  `db:"return_rows"`
	FailCount            int64  `db:"fail_count"`
	ElapsedTime          int64  `db:"elapsed_time"`
	MaxElapsedTime       int64  `db:"max_elapsed_time"`
	CpuTime              int64  `db:"cpu_time"`
	MaxCpuTime           int64  `db:"max_cpu_time"`
	NetTime              int64  `db:"net_time"`
	NetWaitTime          int64  `db:"net_wait_time"`
	QueueTime            int64  `db:"queue_time"`
	DecodeTime           int64  `db:"decode_time"`
	GetPlanTime          int64  `db:"get_plan_time"`
	ExecuteTime          int64  `db:"execute_time"`
	ApplicationWaitTime  int64  `db:"application_wait_time"`
	ConcurrencyWaitTime  int64  `db:"concurrency_wait_time"`
	UserIoWaitTime       int64  `db:"user_io_wait_time"`
	ScheduleTime         int64  `db:"schedule_time"`
	RpcCount             int64  `db:"rpc_count"`
	MissPlanCount        int64  `db:"miss_plan_count"`
	MemstoreReadRowCount int64  `db:"memstore_read_row_count"`
	SSStoreReadRowCount  int64  `db:"ssstore_read_row_count"`
}

// --- SQL Constants ---

const (
	globalSqlAuditTableName = "gv$ob_sql_audit"
)

const (
	selectObserverStats = `SELECT svr_ip, MAX(request_id) as max_request_id FROM %s WHERE tenant_id = ? GROUP BY svr_ip`

	aggregatedSqlAuditQuery = `
    SELECT
        %d AS collect_time,
        svr_ip,
        tenant_id,
        user_id,
        db_id,
        sql_id,
        MAX(request_id) as max_request_id,
        MIN(request_time) as min_request_time,
        MAX(request_time) as max_request_time,
        COUNT(*) as executions,
        SUM(affected_rows) as affected_rows,
        SUM(return_rows) as return_rows,
        SUM(CASE WHEN ret_code = 0 THEN 0 ELSE 1 END) as fail_count,
        SUM(elapsed_time) as elapsed_time,
        MAX(elapsed_time) as max_elapsed_time,
        SUM(execute_time - total_wait_time_micro + get_plan_time) as cpu_time,
        MAX(execute_time - total_wait_time_micro + get_plan_time) as max_cpu_time,
        SUM(net_time) as net_time,
        SUM(net_wait_time) as net_wait_time,
        SUM(queue_time) as queue_time,
        SUM(decode_time) as decode_time,
        SUM(get_plan_time) as get_plan_time,
        SUM(execute_time) as execute_time,
        SUM(application_wait_time) as application_wait_time,
        SUM(concurrency_wait_time) as concurrency_wait_time,
        SUM(user_io_wait_time) as user_io_wait_time,
        SUM(schedule_time) as schedule_time,
        SUM(rpc_count) as rpc_count,
        SUM(CASE WHEN is_hit_plan=1 THEN 0 ELSE 1 END) as miss_plan_count,
        SUM(memstore_read_row_count) as memstore_read_row_count,
        SUM(ssstore_read_row_count) as ssstore_read_row_count
    FROM %s
    WHERE svr_ip = ? AND tenant_id = ? AND request_id >= ? AND request_id <= ?
    GROUP BY svr_ip, tenant_id, user_id, db_id, sql_id`
)
