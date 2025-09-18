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
	BatchSize int64         `yaml:"batchSize"`
	Interval  time.Duration `yaml:"interval"`
}

// Collector manages the data collection from OceanBase.
type Collector struct {
	manager      *operation.OceanbaseOperationManager
	config       *Config
	requestIdMap sync.Map
}

// Tenant represents an OceanBase tenant.
type Tenant struct {
	ID   int64  `db:"tenant_id"`
	Name string `db:"tenant_name"`
}

// NewCollector creates a new Collector instance.
func NewCollector(config *Config, manager *operation.OceanbaseOperationManager) (*Collector, error) {
	return &Collector{
		manager: manager,
		config:  config,
	}, nil
}

// Close closes the database connection.
func (c *Collector) Close() error {
	return c.manager.Close()
}

// Collect performs a single collection cycle.
func (c *Collector) Collect(ctx context.Context) ([]SqlAuditResult, error) {
	log.Println("Starting SQL audit collection...")

	servers, err := c.listServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	tenants, err := c.listTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	log.Printf("Discovered %d servers and %d tenants.", len(servers), len(tenants))

	var allResults []SqlAuditResult
	var wg sync.WaitGroup
	resultChan := make(chan []SqlAuditResult, len(servers)*len(tenants))

	for _, server := range servers {
		for _, tenant := range tenants {
			wg.Add(1)
			go func(s model.OBServer, t Tenant) {
				defer wg.Done()
				log.Printf("Collecting for server %s, tenant %d (%s)", s.Ip, t.ID, t.Name)
				results, err := c.collectForTarget(ctx, s.Ip, t.ID)
				if err != nil {
					log.Printf("Failed to collect for server %s, tenant %d: %v", s.Ip, t.ID, err)
					return
				}
				if len(results) > 0 {
					resultChan <- results
				}
			}(server, tenant)
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

func (c *Collector) listServers(ctx context.Context) ([]model.OBServer, error) {
	return c.manager.ListServers(ctx)
}

func (c *Collector) listTenants(ctx context.Context) ([]Tenant, error) {
	var tenants []Tenant
	err := c.manager.QueryList(ctx, &tenants, "SELECT tenant_id, tenant_name FROM __all_tenant")
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (c *Collector) collectForTarget(ctx context.Context, svrIP string, tenantID int64) ([]SqlAuditResult, error) {
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

	rows, err := c.querySqlAuditRaw(ctx, svrIP, tenantID, startRequestId, maxRequestId)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audit raw: %w", err)
	}
	defer rows.Close()

	aggregatedResults, newMaxRequestId, err := c.parseSqlAuditResults(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sql audit results: %w", err)
	}

	if newMaxRequestId > 0 {
		c.requestIdMap.Store(key, newMaxRequestId)
	}

	finalResults := make([]SqlAuditResult, 0, len(aggregatedResults))
	for _, result := range aggregatedResults {
		finalResults = append(finalResults, result)
	}

	return finalResults, nil
}

func (c *Collector) getMaxMinRequestID(ctx context.Context, svrIP string, tenantID int64) (max int64, min int64, err error) {
	endTime := time.Now()
	startTime := endTime.Add(c.config.Interval * -2)

	query := fmt.Sprintf(selectMaxMinRequestId, globalSqlAuditTableName)
	err = c.manager.Connector.GetClient().QueryRowxContext(ctx, query, svrIP, tenantID, startTime.UnixNano()/1000, endTime.UnixNano()/1000).Scan(&max, &min)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return max, min, nil
}

func (c *Collector) querySqlAuditRaw(ctx context.Context, svrIP string, tenantID int64, minRequestId int64, maxRequestId int64) (*sql.Rows, error) {
	query := fmt.Sprintf(sqlAuditRawQuery, globalSqlAuditTableName)
	args := []interface{}{svrIP, tenantID, minRequestId, maxRequestId, c.config.BatchSize}
	return c.manager.Connector.GetClient().QueryContext(ctx, query, args...)
}

func (c *Collector) parseSqlAuditResults(ctx context.Context, rows *sql.Rows) (map[SqlAuditKey]SqlAuditResult, int64, error) {
	aggregatedResults := make(map[SqlAuditKey]SqlAuditResult)
	var innerMaxRequestId int64 = 0

	for rows.Next() {
		rawResult := &SqlAuditRawResult{}
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
			&rawResult.BlockIndexCacheHit, &rawResult.ExpectedWorkerCount, &rawResult.UsedWorkerCount,
			&rawResult.MemoryUsed, &rawResult.TransactionHash, &rawResult.EffectiveTenantId,
			&rawResult.ParamsValue, &rawResult.FltTraceId, &rawResult.FormatSqlId, &rawResult.PlanHash,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		innerMaxRequestId = rawResult.RequestId

		if len(rawResult.SqlId) == 0 || len(rawResult.TenantName) == 0 {
			continue
		}

		auditKey := SqlAuditKey{
			TenantId: rawResult.TenantId,
			UserId:   rawResult.UserId,
			DbId:     rawResult.DbId,
			SqlId:    rawResult.SqlId,
		}

		convertedResult := convertToSqlAuditResult(rawResult)

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

// --- Data Structures ---

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

// --- SQL Constants ---

const (
	globalSqlAuditTableName = "gv$ob_sql_audit"
)

const (
	selectMaxMinRequestId = `SELECT IFNULL(MAX(request_id), 0), IFNULL(MIN(request_id), 0) FROM %s WHERE svr_ip = ? AND tenant_id = ? AND (request_time + elapsed_time) >= ? AND (request_time + elapsed_time) < ?`

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
    WHERE svr_ip = ? AND tenant_id = ? AND request_id >= ? AND request_id <= ?
    LIMIT ?`
)
