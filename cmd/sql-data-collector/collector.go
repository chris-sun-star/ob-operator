package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// Config holds the configuration for the collector.
type Config struct {
	Interval  time.Duration `yaml:"interval"`
	BatchSize int64         `yaml:"batchSize"`
}

// Collector manages the data collection from OceanBase.
type Collector struct {
	manager      *operation.OceanbaseOperationManager
	config       *Config
	tenantID     int64
	requestIdMap sync.Map
}

// Tenant represents an OceanBase tenant, used for discovery.
type Tenant struct {
	ID   int64  `db:"tenant_id"`
	Name string `db:"tenant_name"`
}

// NewCollector creates a new Collector instance.
func NewCollector(config *Config, manager *operation.OceanbaseOperationManager, tenantID int64) (*Collector, error) {
	return &Collector{
		manager:  manager,
		config:   config,
		tenantID: tenantID,
	}, nil
}

// Close is a no-op since the manager's lifecycle is handled in main.
func (c *Collector) Close() error {
	return nil
}

// Collect performs a single collection cycle for the configured tenant.
func (c *Collector) Collect(ctx context.Context) ([]SqlAuditResult, error) {
	log.Println("Starting SQL audit collection...")

	servers, err := c.manager.ListServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	log.Printf("Discovered %d servers for tenant %d.", len(servers), c.tenantID)

	var allResults []SqlAuditResult
	var wg sync.WaitGroup
	resultChan := make(chan []SqlAuditResult, len(servers))

	for _, server := range servers {
		wg.Add(1)
		go func(s model.OBServer) {
			defer wg.Done()
			log.Printf("Collecting for server %s, tenant %d", s.Ip, c.tenantID)
			results, err := c.collectAndAggregate(ctx, s.Ip, c.tenantID)
			if err != nil {
				log.Printf("Failed to collect for server %s, tenant %d: %v", s.Ip, c.tenantID, err)
				return
			}
			if len(results) > 0 {
				resultChan <- results
			}
		}(server)
	}

	wg.Wait()
	close(resultChan)

	for res := range resultChan {
		allResults = append(allResults, res...)
	}

	log.Printf("Collection finished. Aggregated %d SQL statements in total.", len(allResults))
	return allResults, nil
}

func (c *Collector) collectAndAggregate(ctx context.Context, svrIP string, tenantID int64) ([]SqlAuditResult, error) {
	key := fmt.Sprintf("%s-%d", svrIP, tenantID)

	maxRequestId, minRequestId, err := c.getMaxMinRequestID(ctx, svrIP, tenantID)
	if err != nil {
		return nil, fmt.Errorf("could not get max/min request_id: %w", err)
	}

	lastStartRequestId, _ := c.requestIdMap.Load(key)
	if lastStartRequestId == nil {
		lastStartRequestId = int64(0)
	}
	startRequestId := lastStartRequestId.(int64) + 1

	if startRequestId == 1 { // First run for this target
		startRequestId = minRequestId
	}

	if maxRequestId > 0 && maxRequestId < (startRequestId-1) {
		log.Printf("Request ID reset for %s. Max ID (%d) < Start ID (%d). Resetting to min ID (%d).", key, maxRequestId, startRequestId, minRequestId)
		startRequestId = minRequestId
	}

	if maxRequestId <= lastStartRequestId.(int64) {
		log.Printf("No new data for %s.", key)
		return nil, nil
	}

	var results []SqlAuditResult
	query := fmt.Sprintf(aggregatedSqlAuditQuery, globalSqlAuditTableName)
	args := []interface{}{svrIP, tenantID, startRequestId, maxRequestId}

	err = c.manager.Connector.GetClient().SelectContext(ctx, &results, query, args...)
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

func (c *Collector) getMaxMinRequestID(ctx context.Context, svrIP string, tenantID int64) (max int64, min int64, err error) {
	endTime := time.Now()
	startTime := endTime.Add(c.config.Interval * -2)

	query := fmt.Sprintf(selectMaxMinRequestId, globalSqlAuditTableName)
	row := c.manager.Connector.GetClient().QueryRowxContext(ctx, query, svrIP, tenantID, startTime.UnixMicro(), endTime.UnixMicro())
	err = row.Scan(&max, &min)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return max, min, nil
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
	selectMaxMinRequestId = `SELECT IFNULL(MAX(request_id), 0), IFNULL(MIN(request_id), 0) FROM %s WHERE svr_ip = ? AND tenant_id = ? AND request_time >= ? AND request_time < ?`

	aggregatedSqlAuditQuery = `
    SELECT
        time_to_usec(now()) AS collect_time,
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
