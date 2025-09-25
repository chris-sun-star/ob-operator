package sqldatacollector

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// DuckDBManager handles operations with the DuckDB database.
type DuckDBManager struct {
	db *sql.DB
}

// NewDuckDBManager creates a new DuckDBManager and initializes the database table.
func NewDuckDBManager(path string) (*DuckDBManager, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS sql_audit (
            svr_ip VARCHAR, tenant_id BIGINT, tenant_name VARCHAR, user_id BIGINT, user_name VARCHAR,
            db_id BIGINT, db_name VARCHAR, sql_id VARCHAR, plan_id BIGINT,

            query_sql TEXT, client_ip VARCHAR, event VARCHAR,
            format_sql_id VARCHAR, effective_tenant_id BIGINT, trace_id VARCHAR, sid BIGINT,
            user_client_ip VARCHAR, tx_id VARCHAR,

            executions BIGINT, min_request_time BIGINT, max_request_time BIGINT,
            max_request_id BIGINT, min_request_id BIGINT,

            elapsed_time_sum BIGINT, elapsed_time_max BIGINT, elapsed_time_min BIGINT,
            execute_time_sum BIGINT, execute_time_max BIGINT, execute_time_min BIGINT,
            queue_time_sum BIGINT, queue_time_max BIGINT, queue_time_min BIGINT,
            get_plan_time_sum BIGINT, get_plan_time_max BIGINT, get_plan_time_min BIGINT,
            affected_rows_sum BIGINT, affected_rows_max BIGINT, affected_rows_min BIGINT,
            return_rows_sum BIGINT, return_rows_max BIGINT, return_rows_min BIGINT,
            partition_count_sum BIGINT, partition_count_max BIGINT, partition_count_min BIGINT,
            retry_count_sum BIGINT, retry_count_max BIGINT, retry_count_min BIGINT,
            disk_reads_sum BIGINT, disk_reads_max BIGINT, disk_reads_min BIGINT,
            rpc_count_sum BIGINT, rpc_count_max BIGINT, rpc_count_min BIGINT,
            memstore_read_row_count_sum BIGINT, memstore_read_row_count_max BIGINT, memstore_read_row_count_min BIGINT,
            ssstore_read_row_count_sum BIGINT, ssstore_read_row_count_max BIGINT, ssstore_read_row_count_min BIGINT,
            request_memory_used_sum BIGINT, request_memory_used_max BIGINT, request_memory_used_min BIGINT,
            wait_time_micro_sum BIGINT, wait_time_micro_max BIGINT, wait_time_micro_min BIGINT,
            total_wait_time_micro_sum BIGINT, total_wait_time_micro_max BIGINT, total_wait_time_micro_min BIGINT,
            net_time_sum BIGINT, net_time_max BIGINT, net_time_min BIGINT,
            net_wait_time_sum BIGINT, net_wait_time_max BIGINT, net_wait_time_min BIGINT,
            decode_time_sum BIGINT, decode_time_max BIGINT, decode_time_min BIGINT,
            application_wait_time_sum BIGINT, application_wait_time_max BIGINT, application_wait_time_min BIGINT,
            concurrency_wait_time_sum BIGINT, concurrency_wait_time_max BIGINT, concurrency_wait_time_min BIGINT,
            user_io_wait_time_sum BIGINT, user_io_wait_time_max BIGINT, user_io_wait_time_min BIGINT,
            schedule_time_sum BIGINT, schedule_time_max BIGINT, schedule_time_min BIGINT,
            row_cache_hit_sum BIGINT, row_cache_hit_max BIGINT, row_cache_hit_min BIGINT,
            bloom_filter_cache_hit_sum BIGINT, bloom_filter_cache_hit_max BIGINT, bloom_filter_cache_hit_min BIGINT,
            block_cache_hit_sum BIGINT, block_cache_hit_max BIGINT, block_cache_hit_min BIGINT,
            index_block_cache_hit_sum BIGINT, index_block_cache_hit_max BIGINT, index_block_cache_hit_min BIGINT,
            expected_worker_count_sum BIGINT, expected_worker_count_max BIGINT, expected_worker_count_min BIGINT,
            used_worker_count_sum BIGINT, used_worker_count_max BIGINT, used_worker_count_min BIGINT,
            table_scan_sum BIGINT, table_scan_max BIGINT, table_scan_min BIGINT,
            consistency_level_strong_count BIGINT,
            consistency_level_weak_count BIGINT,
            fail_count_sum BIGINT,
			ret_code_4012_count_sum BIGINT, ret_code_4013_count_sum BIGINT, ret_code_5001_count_sum BIGINT,
			ret_code_5024_count_sum BIGINT, ret_code_5167_count_sum BIGINT, ret_code_5217_count_sum BIGINT,
			ret_code_6002_count_sum BIGINT,
			event_0_wait_time_sum BIGINT, event_1_wait_time_sum BIGINT, event_2_wait_time_sum BIGINT,
			event_3_wait_time_sum BIGINT,
			plan_type_local_count BIGINT, plan_type_remote_count BIGINT, plan_type_distributed_count BIGINT,
			inner_sql_count BIGINT,
			miss_plan_count BIGINT,
			executor_rpc_count BIGINT,

            collect_time TIMESTAMPTZ,

PRIMARY KEY (svr_ip, tenant_id, tenant_name, user_id, user_name, db_id, db_name, sql_id, plan_id, max_request_id)
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DuckDBManager{db: db}, nil
}

// GetLastRequestIDs retrieves the last request ID for each server from the database.
func (m *DuckDBManager) GetLastRequestIDs() (map[string]uint64, error) {
	rows, err := m.db.Query("SELECT svr_ip, MAX(max_request_id) FROM sql_audit GROUP BY svr_ip")
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return make(map[string]uint64), nil
		}
		return nil, err
	}
	defer rows.Close()

	lastRequestIDs := make(map[string]uint64)
	for rows.Next() {
		var svrIP string
		var maxRequestID uint64
		if err := rows.Scan(&svrIP, &maxRequestID); err != nil {
			return nil, err
		}
		lastRequestIDs[svrIP] = maxRequestID
	}
	return lastRequestIDs, nil
}

// InsertBatch inserts a batch of SQL audit data into the database.
func (m *DuckDBManager) InsertBatch(results []SQLAudit) error {
	txn, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Dynamically generate the INSERT statement
	if len(results) == 0 {
		return nil
	}
	numColumns := 131 // Number of fields in SQLAudit + collect_time
	placeholders := strings.Repeat("?,", numColumns-1) + "?"
	sql := fmt.Sprintf("INSERT OR IGNORE INTO sql_audit VALUES (%s)", placeholders)

	stmt, err := txn.Prepare(sql)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	collectTime := time.Now()

	for _, r := range results {
		_, err := stmt.Exec(
			// Grouping Keys
			r.SvrIP, r.TenantId, r.TenantName, r.UserId, r.UserName, r.DbId, r.DBName, r.SqlId, r.PlanId,
			// Aggregated String/Identifier Values
			r.QuerySql, r.ClientIp, r.Event, r.FormatSqlId, r.EffectiveTenantId, r.TraceId, r.Sid, r.UserClientIp, r.TxId,
			// Aggregated Numeric Values
			r.Executions, r.MinRequestTime, r.MaxRequestTime, r.MaxRequestId, r.MinRequestId,
			r.ElapsedTimeSum, r.ElapsedTimeMax, r.ElapsedTimeMin,
			r.ExecuteTimeSum, r.ExecuteTimeMax, r.ExecuteTimeMin,
			r.QueueTimeSum, r.QueueTimeMax, r.QueueTimeMin,
			r.GetPlanTimeSum, r.GetPlanTimeMax, r.GetPlanTimeMin,
			r.AffectedRowsSum, r.AffectedRowsMax, r.AffectedRowsMin,
			r.ReturnRowsSum, r.ReturnRowsMax, r.ReturnRowsMin,
			r.PartitionCountSum, r.PartitionCountMax, r.PartitionCountMin,
			r.RetryCountSum, r.RetryCountMax, r.RetryCountMin,
			r.DiskReadsSum, r.DiskReadsMax, r.DiskReadsMin,
			r.RpcCountSum, r.RpcCountMax, r.RpcCountMin,
			r.MemstoreReadRowCountSum, r.MemstoreReadRowCountMax, r.MemstoreReadRowCountMin,
			r.SSStoreReadRowCountSum, r.SSStoreReadRowCountMax, r.SSStoreReadRowCountMin,
			r.RequestMemoryUsedSum, r.RequestMemoryUsedMax, r.RequestMemoryUsedMin,
			r.WaitTimeMicroSum, r.WaitTimeMicroMax, r.WaitTimeMicroMin,
			r.TotalWaitTimeMicroSum, r.TotalWaitTimeMicroMax, r.TotalWaitTimeMicroMin,
			r.NetTimeSum, r.NetTimeMax, r.NetTimeMin,
			r.NetWaitTimeSum, r.NetWaitTimeMax, r.NetWaitTimeMin,
			r.DecodeTimeSum, r.DecodeTimeMax, r.DecodeTimeMin,
			r.ApplicationWaitTimeSum, r.ApplicationWaitTimeMax, r.ApplicationWaitTimeMin,
			r.ConcurrencyWaitTimeSum, r.ConcurrencyWaitTimeMax, r.ConcurrencyWaitTimeMin,
			r.UserIoWaitTimeSum, r.UserIoWaitTimeMax, r.UserIoWaitTimeMin,
			r.ScheduleTimeSum, r.ScheduleTimeMax, r.ScheduleTimeMin,
			r.RowCacheHitSum, r.RowCacheHitMax, r.RowCacheHitMin,
			r.BloomFilterCacheHitSum, r.BloomFilterCacheHitMax, r.BloomFilterCacheHitMin,
			r.BlockCacheHitSum, r.BlockCacheHitMax, r.BlockCacheHitMin,
			r.IndexBlockCacheHitSum, r.IndexBlockCacheHitMax, r.IndexBlockCacheHitMin,
			r.ExpectedWorkerCountSum, r.ExpectedWorkerCountMax, r.ExpectedWorkerCountMin,
			r.UsedWorkerCountSum, r.UsedWorkerCountMax, r.UsedWorkerCountMin,
			r.TableScanSum, r.TableScanMax, r.TableScanMin,
			r.ConsistencyLevelStrongCount,
			r.ConsistencyLevelWeakCount,
			r.FailCountSum,
			r.RetCode4012CountSum, r.RetCode4013CountSum, r.RetCode5001CountSum, r.RetCode5024CountSum,
			r.RetCode5167CountSum, r.RetCode5217CountSum, r.RetCode6002CountSum,
			r.Event0WaitTimeSum, r.Event1WaitTimeSum, r.Event2WaitTimeSum, r.Event3WaitTimeSum,
			r.PlanTypeLocalCount, r.PlanTypeRemoteCount, r.PlanTypeDistributedCount,
			r.InnerSqlCount,
			r.MissPlanCount,
			r.ExecutorRpcCount,

			collectTime,
		)
		if err != nil {
			txn.Rollback()
			recordJSON, _ := json.Marshal(r)
			return fmt.Errorf("failed to execute statement for record %s: %w", string(recordJSON), err)
		}
	}

	return txn.Commit()
}

// Close closes the database connection.
func (m *DuckDBManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}
