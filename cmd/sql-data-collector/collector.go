
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config holds the configuration for the collector.
type Config struct {
	DSN       string        `yaml:"dsn"`
	BatchSize int64         `yaml:"batchSize"`
	Interval  time.Duration `yaml:"interval"`
}

// Collector manages the data collection from OceanBase.
type Collector struct {
	db           *sql.DB
	config       *Config
	requestIdMap sync.Map
}

// NewCollector creates a new Collector instance.
func NewCollector(config *Config) (*Collector, error) {
	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Collector{
		db:     db,
		config: config,
	}, nil
}

// Close closes the database connection.
func (c *Collector) Close() error {
	return c.db.Close()
}

// Collect performs a single collection cycle.
func (c *Collector) Collect(ctx context.Context) ([]SqlAuditResult, error) {
	log.Println("Starting SQL audit collection...")
	results, err := c.doCollect(ctx)
	if err != nil {
		return nil, fmt.Errorf("collection failed: %w", err)
	}
	log.Printf("Collection finished. Aggregated %d SQL statements.", len(results))
	return results, nil
}

func (c *Collector) doCollect(ctx context.Context) ([]SqlAuditResult, error) {
	// In this simplified version, we collect for all tenants at once.
	// The original had per-tenant collection logic which can be added back if needed.

	// 1. Get the max and min request_id to define the collection window.
	maxRequestId, minRequestId, err := c.getMaxMinRequestID(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get max/min request_id: %w", err)
	}

	// 2. Determine the starting request_id for this collection cycle.
	lastStartRequestId, _ := c.requestIdMap.Load("all_tenants")
	if lastStartRequestId == nil {
		lastStartRequestId = int64(0)
	}
	startRequestId := lastStartRequestId.(int64) + 1

	if startRequestId == 1 { // First run
		startRequestId = minRequestId
	}

	// Handle observer restarts where request_id might reset.
	if maxRequestId > 0 && maxRequestId < (startRequestId-1) {
		log.Printf("Request ID reset detected. Max ID (%d) < Start ID (%d). Resetting to min ID (%d).", maxRequestId, startRequestId, minRequestId)
		startRequestId = minRequestId
	}

	if maxRequestId <= lastStartRequestId.(int64) {
		log.Println("No new SQL audit data to collect.")
		return nil, nil
	}

	log.Printf("Collecting SQL audit data from request_id %d to %d", startRequestId, maxRequestId)

	// 3. Query the raw data.
	rows, err := c.querySqlAuditRaw(ctx, startRequestId, maxRequestId)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audit raw: %w", err)
	}
	defer rows.Close()

	// 4. Parse and aggregate the results.
	aggregatedResults, newMaxRequestId, err := c.parseSqlAuditResults(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sql audit results: %w", err)
	}

	// 5. Store the new max request_id for the next cycle.
	if newMaxRequestId > 0 {
		c.requestIdMap.Store("all_tenants", newMaxRequestId)
		log.Printf("Updated max request_id to %d for next cycle.", newMaxRequestId)
	}

	// 6. Convert map to slice
	finalResults := make([]SqlAuditResult, 0, len(aggregatedResults))
	for _, result := range aggregatedResults {
		finalResults = append(finalResults, result)
	}

	return finalResults, nil
}

func (c *Collector) getMaxMinRequestID(ctx context.Context) (max int64, min int64, err error) {
	// Simplified query to get the min/max request_id in the last 2 intervals.
	// This helps define a reasonable window to avoid collecting very old data on the first run.
	endTime := time.Now()
	startTime := endTime.Add(c.config.Interval * -2)

	query := fmt.Sprintf(selectMaxMinRequestId, sqlAuditTableNameForObVersion4)
	err = c.db.QueryRowContext(ctx, query, startTime.UnixNano()/1000, endTime.UnixNano()/1000).Scan(&max, &min)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return max, min, nil
}

func (c *Collector) querySqlAuditRaw(ctx context.Context, minRequestId int64, maxRequestId int64) (*sql.Rows, error) {
	query := fmt.Sprintf(sqlAuditRawQuery, sqlAuditTableNameForObVersion4)
	args := []interface{}{minRequestId, maxRequestId, c.config.BatchSize}
	return c.db.QueryContext(ctx, query, args...)
}

func (c *Collector) parseSqlAuditResults(ctx context.Context, rows *sql.Rows) (map[SqlAuditKey]SqlAuditResult, int64, error) {
	aggregatedResults := make(map[SqlAuditKey]SqlAuditResult)
	var innerMaxRequestId int64 = 0
	var rowCount int64 = 0

	for rows.Next() {
		rowCount++
		rawResult := &SqlAuditRawResult{}
		// The number of columns must match the `sqlAuditRawQuery`
		err := rows.Scan(
			&rawResult.SvrIp, &rawResult.SqlId, &rawResult.TenantId, &rawResult.TenantName,
			&rawResult.UserId, &rawResult.UserName, &rawResult.DbId, &rawResult.DbName,
			&rawResult.PlanId, &rawResult.TraceId, &rawResult.Sid, &rawResult.ClientIp,
			&rawResult.ClientPort, &rawResult.UserClientIp, &rawResult.RequestTime, &rawResult.RequestId,
			&rawResult.ElapsedTime, &rawResult.ExecuteTime, &rawResult.TotalWaitTimeMicro, &rawResult.WaitTimeMicro,
			&rawResult.GetPlanTime, &rawResult.AffectedRows, &rawResult.ReturnRows, &rawResult.PartitionCount,
			&rawResult.RetCode, &rawResult.FailCount, &rawResult.RetCode4012Count, &rawResult.RetCode4013Count,
			&rawResult.RetCode5001Count, &rawResult.RetCode5024Count, &rawResult.RetCode5167Count,
			&rawResult.RetCode5217Count, &rawResult.RetCode6002Count, &rawResult.Event, &rawResult.P1,
			&rawResult.P2, &rawResult.P3, &rawResult.P1Text, &rawResult.P2Text, &rawResult.P3Text,
			&rawResult.Event0WaitTime, &rawResult.Event1WaitTime, &rawResult.Event2WaitTime, &rawResult.Event3WaitTime,
			&rawResult.TotalWaits, &rawResult.RpcCount, &rawResult.PlanType, &rawResult.PlanTypeLocalCount,
			&rawResult.PlanTypeRemoteCount, &rawResult.PlanTypeDistCount, &rawResult.IsInnerSql, &rawResult.IsExecutorRpc,
			&rawResult.IsHitPlan, &rawResult.ConsistencyLevel, &rawResult.InnerSqlCount, &rawResult.ExecutorRpcCount,
			&rawResult.MissPlanCount, &rawResult.ConsistencyLevelStrong, &rawResult.ConsistencyLevelWeak,
			&rawResult.NetTime, &rawResult.NetWaitTime, &rawResult.QueueTime, &rawResult.DecodeTime,
			&rawResult.ApplicationWaitTime, &rawResult.ConcurrencyWaitTime, &rawResult.UserIoWaitTime,
			&rawResult.ScheduleTime, &rawResult.RowCacheHit, &rawResult.BloomFilterCacheHit, &rawResult.BlockCacheHit,
			&rawResult.DiskReads, &rawResult.RetryCount, &rawResult.TableScan, &rawResult.MemstoreReadRowCount,
			&rawResult.SSStoreReadRowCount, &rawResult.QuerySql,
			// Extended columns for OB 4.0+
			&rawResult.BlockIndexCacheHit, &rawResult.ExpectedWorkerCount, &rawResult.UsedWorkerCount,
			&rawResult.MemoryUsed, &rawResult.TransactionHash, &rawResult.EffectiveTenantId,
			&rawResult.ParamsValue, &rawResult.FltTraceId, &rawResult.FormatSqlId, &rawResult.PlanHash,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		innerMaxRequestId = rawResult.RequestId

		// Basic validation
		if len(rawResult.SqlId) == 0 || len(rawResult.TenantName) == 0 {
			continue
		}

		// Key for aggregation
		auditKey := SqlAuditKey{
			TenantId: rawResult.TenantId,
			UserId:   rawResult.UserId,
			DbId:     rawResult.DbId,
			SqlId:    rawResult.SqlId,
		}

		convertedResult := convertToSqlAuditResult(rawResult)

		// Aggregate
		if existing, found := aggregatedResults[auditKey]; found {
			aggregatedResults[auditKey] = aggregateGroupBySqlId(&existing, &convertedResult)
		} else {
			aggregatedResults[auditKey] = convertedResult
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return aggregatedResults, innerMaxRequestId, nil
}

// --- Data Structures (from ocp-agent) ---

type SqlAuditKey struct {
	TenantId uint64
	UserId   int64
	DbId     uint64
	SqlId    string
}

type SqlAuditResult struct {
	CollectTime           int64
	SqlId                 string
	FormatSqlId           string
	TenantId              uint64
	EffectiveTenantId     uint64
	TenantName            string
	UserId                int64
	UserName              string
	DbId                  uint64
	DbName                string
	QuerySql              string
	PlanId                int64
	TraceId               string
	MaxRequestId          int64
	MinRequestId          int64
	MaxRequestTime        int64
	MinRequestTime        int64
	MaxCpuTime            int64
	MaxElapsedTime        int64
	Executions            int64
	AffectedRows          int64
	ReturnRows            int64
	PartitionCount        int64
	FailCount             int64
	RetCode4012Count      int64
	RetCode4013Count      int64
	RetCode5001Count      int64
	RetCode5024Count      int64
	RetCode5167Count      int64
	RetCode5217Count      int64
	RetCode6002Count      int64
	Event0WaitTime        int64
	Event1WaitTime        int64
	Event2WaitTime        int64
	Event3WaitTime        int64
	TotalWaitTimeMicro    int64
	TotalWaits            int64
	RpcCount              int64
	PlanTypeLocalCount    int64
	PlanTypeRemoteCount   int64
	PlanTypeDistCount     int64
	InnerSqlCount         int64
	ExecutorRpcCount      int64
	MissPlanCount         int64
	ElapsedTime           int64
	NetTime               int64
	NetWaitTime           int64
	QueueTime             int64
	DecodeTime            int64
	GetPlanTime           int64
	ExecuteTime           int64
	CpuTime               int64
	ApplicationWaitTime   int64
	ConcurrencyWaitTime   int64
	UserIoWaitTime        int64
	ScheduleTime          int64
	RowCacheHit           int64
	BloomFilterCacheHit   int64
	BlockCacheHit         int64
	BlockIndexCacheHit    int64
	DiskReads             int64
	RetryCount            int64
	TableScan             int64
	ConsistencyLevelStrong int64
	ConsistencyLevelWeak  int64
	MemstoreReadRowCount  int64
	SSStoreReadRowCount   int64
	ExpectedWorkerCount   int64
	UsedWorkerCount       int64
	MemoryUsed            int64
}

type SqlAuditRawResult struct {
	SvrIp                  string
	SqlId                  string
	TenantId               uint64
	EffectiveTenantId      uint64
	TenantName             string
	UserId                 int64
	UserName               string
	DbId                   uint64
	DbName                 string
	QuerySql               string
	PlanId                 int64
	TraceId                string
	Sid                    uint64
	ClientIp               string
	ClientPort             int64
	UserClientIp           string
	RequestTime            int64
	RequestId              int64
	ElapsedTime            int64
	ExecuteTime            int64
	WaitTimeMicro          int64
	TotalWaitTimeMicro     int64
	GetPlanTime            int64
	AffectedRows           int64
	ReturnRows             int64
	PartitionCount         int64
	FailCount              int64
	RetCode                int64
	RetCode4012Count       int64
	RetCode4013Count       int64
	RetCode5001Count       int64
	RetCode5024Count       int64
	RetCode5167Count       int64
	RetCode5217Count       int64
	RetCode6002Count       int64
	Event                  string
	P1                     uint64
	P2                     uint64
	P3                     uint64
	P1Text                 string
	P2Text                 string
	P3Text                 string
	Event0WaitTime         int64
	Event1WaitTime         int64
	Event2WaitTime         int64
	Event3WaitTime         int64
	TotalWaits             int64
	RpcCount               int64
	PlanType               int64
	IsInnerSql             int64
	IsExecutorRpc          int64
	IsHitPlan              int64
	PlanTypeLocalCount     int64
	PlanTypeRemoteCount    int64
	PlanTypeDistCount      int64
	InnerSqlCount          int64
	ExecutorRpcCount       int64
	MissPlanCount          int64
	NetTime                int64
	NetWaitTime            int64
	QueueTime              int64
	DecodeTime             int64
	ApplicationWaitTime    int64
	ConcurrencyWaitTime    int64
	UserIoWaitTime         int64
	ScheduleTime           int64
	RowCacheHit            int64
	BloomFilterCacheHit    int64
	BlockCacheHit          int64
	BlockIndexCacheHit     int64
	DiskReads              int64
	RetryCount             int64
	TableScan              int64
	ConsistencyLevel       int64
	ConsistencyLevelStrong int64
	ConsistencyLevelWeak   int64
	MemstoreReadRowCount   int64
	SSStoreReadRowCount    int64
	ExpectedWorkerCount    int64
	UsedWorkerCount        int64
	MemoryUsed             int64
	TransactionHash        string
	ParamsValue            string
	FltTraceId             string
	FormatSqlId            string
	PlanHash               int64
}

// --- Constants ---

const (
	sqlAuditTableNameForObVersion4 = "V$OB_SQL_AUDIT"
)

const (
	selectMaxMinRequestId = "SELECT IFNULL(MAX(request_id), 0), IFNULL(MIN(request_id), 0) FROM %s WHERE (request_time + elapsed_time) >= ? AND (request_time + elapsed_time) < ?"

	// This query is for OceanBase 4.0+
	sqlAuditRawQuery = `
    SELECT
        svr_ip, sql_id, tenant_id, tenant_name, user_id, user_name, db_id, db_name,
        plan_id, trace_id, sid, client_ip, client_port, user_client_ip, request_time, request_id,
        elapsed_time, execute_time, total_wait_time_micro, wait_time_micro, get_plan_time,
        affected_rows, return_rows, partition_cnt, ret_code,
        (CASE WHEN ret_code = 0 THEN 0 ELSE 1 END) as fail_count,
        (CASE WHEN ret_code = -4012 THEN 1 ELSE 0 END) as ret_code_4012_count,
        (CASE WHEN ret_code = -4013 THEN 1 ELSE 0 END) as ret_code_4013_count,
        (CASE WHEN ret_code = -5001 THEN 1 ELSE 0 END) as ret_code_5001_count,
        (CASE WHEN ret_code = -5024 THEN 1 ELSE 0 END) as ret_code_5024_count,
        (CASE WHEN ret_code = -5167 THEN 1 ELSE 0 END) as ret_code_5167_count,
        (CASE WHEN ret_code = -5217 THEN 1 ELSE 0 END) as ret_code_5217_count,
        (CASE WHEN ret_code = -6002 THEN 1 ELSE 0 END) as ret_code_6002_count,
        event, p1, p2, p3, p1text, p2text, p3text,
        (CASE event WHEN 'system internal wait' THEN wait_time_micro ELSE 0 END) as event_0_wait_time,
        (CASE event WHEN 'mysql response wait client' THEN wait_time_micro ELSE 0 END) as event_1_wait_time,
        (CASE event WHEN 'sync rpc' THEN wait_time_micro ELSE 0 END) as event_2_wait_time,
        (CASE event WHEN 'db file data read' THEN wait_time_micro ELSE 0 END) as event_3_wait_time,
        total_waits, rpc_count, plan_type,
        (CASE WHEN plan_type=1 THEN 1 ELSE 0 END) as plan_type_local_count,
        (CASE WHEN plan_type=2 THEN 1 ELSE 0 END) as plan_type_remote_count,
        (CASE WHEN plan_type=3 THEN 1 ELSE 0 END) as plan_type_dist_count,
        is_inner_sql, is_executor_rpc, is_hit_plan, consistency_level,
        (CASE WHEN is_inner_sql=1 THEN 1 ELSE 0 END) as inner_sql_count,
        (CASE WHEN is_executor_rpc = 1 THEN 1 ELSE 0 END) as executor_rpc_count,
        (CASE WHEN is_hit_plan=1 THEN 0 ELSE 1 END) as miss_plan_count,
        (CASE consistency_level WHEN 3 THEN 1 ELSE 0 END) as consistency_level_strong,
        (CASE consistency_level WHEN 2 THEN 1 ELSE 0 END) as consistency_level_weak,
        net_time, net_wait_time, queue_time, decode_time, application_wait_time,
        concurrency_wait_time, user_io_wait_time, schedule_time, row_cache_hit,
        bloom_filter_cache_hit, block_cache_hit, disk_reads, retry_cnt, table_scan,
        memstore_read_row_count, ssstore_read_row_count, query_sql,
		0 as block_index_cache_hit, expected_worker_count, used_worker_count, request_memory_used,
		tx_id as transaction_hash, effective_tenant_id, params_value, flt_trace_id, format_sql_id, plan_hash
    FROM %s
    WHERE request_id >= ? AND request_id <= ?
    LIMIT ?`
)
