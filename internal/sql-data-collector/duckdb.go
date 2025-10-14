package sqldatacollector

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	duckdb "github.com/marcboeker/go-duckdb"
	"github.com/google/uuid"
)

// DuckDBManager handles operations with the DuckDB database.
type DuckDBManager struct {
	db   *sql.DB
	path string // path to the directory storing daily parquet files
}

// NewDuckDBManager creates a new DuckDBManager.
// The path is the directory where the daily parquet files will be stored.
func NewDuckDBManager(path string) (*DuckDBManager, error) {
	// Use an in-memory DuckDB database for operations.
	db, err := sql.Open("duckdb", "") // In-memory
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory duckdb: %w", err)
	}

	// Get and log DuckDB version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get DuckDB version: %w", err)
	}
	log.Printf("DuckDB version: %s", version)

	// Ensure the data directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}

	return &DuckDBManager{db: db, path: path}, nil
}

// GetLastRequestIDs retrieves the last request ID for each server from the most recent daily parquet file.
func (m *DuckDBManager) GetLastRequestIDs() (map[string]uint64, error) {
	files, err := filepath.Glob(filepath.Join(m.path, "*.parquet"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob parquet files: %w", err)
	}

	if len(files) == 0 {
		return make(map[string]uint64), nil
	}

	// Find the most recent file by parsing the date from the filename.
	var latestFile string
	var latestDate time.Time
	for _, file := range files {
		fileName := filepath.Base(file)
		dateStr := strings.TrimSuffix(fileName, ".parquet")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Skip files with invalid date format in their name.
			log.Printf("Skipping file %s with invalid date format: %v", fileName, err)
			continue
		}
		if latestFile == "" || fileDate.After(latestDate) {
			latestDate = fileDate
			latestFile = file
		}
	}

	if latestFile == "" {
		return make(map[string]uint64), nil
	}

	query := fmt.Sprintf("SELECT svr_ip, MAX(max_request_id) FROM read_parquet('%s') GROUP BY svr_ip", latestFile)

	rows, err := m.db.Query(query)
	if err != nil {
		// If the file is empty or corrupted, it might error.
		log.Printf("Error querying latest parquet file %s: %v", latestFile, err)
		return make(map[string]uint64), nil // Return empty map to start fresh for this file
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
	if len(results) == 0 {
		return nil
	}

	conn, err := m.db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	// Determine the target Parquet file based on the current date.
	currentDate := time.Now().Format("2006-01-02")
	targetParquetFile := filepath.Join(m.path, fmt.Sprintf("%s.parquet", currentDate))

	// Create a temporary table for this batch.
	tempTableName := "sql_audit_batch_" + uuid.New().String()[:8] // Use a unique temp table name
	createTempTableSQL := `CREATE TEMP TABLE ` + tempTableName + ` (
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
        collect_date DATE
    )`
	if _, err := conn.ExecContext(context.Background(), createTempTableSQL); err != nil {
		return fmt.Errorf("failed to create temp table: %w", err)
	}

	// Use the appender to load data into the temp table.
	err = conn.Raw(func(driverConn interface{}) error {
		duckdbConn, ok := driverConn.(driver.Conn)
		if !ok {
			return fmt.Errorf("failed to get raw duckdb connection")
		}
		appender, err := duckdb.NewAppenderFromConn(duckdbConn, "", tempTableName)
		if err != nil {
			return fmt.Errorf("failed to create appender: %w", err)
		}
		defer appender.Close()

		collectTime := time.Now()

		for _, r := range results {
			err := appender.AppendRow(
				r.SvrIP, r.TenantId, r.TenantName, r.UserId, r.UserName, r.DbId, r.DBName, r.SqlId, r.PlanId,
				r.QuerySql, r.ClientIp, r.Event, r.FormatSqlId, r.EffectiveTenantId, r.TraceId, r.Sid, r.UserClientIp, r.TxId,
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
				collectTime,
			)
			if err != nil {
				return fmt.Errorf("failed to append row to temp table: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to append data: %w", err)
	}

	// If the target file exists, load its data into the temp table.
	if _, err := os.Stat(targetParquetFile); err == nil {
		loadSQL := fmt.Sprintf("INSERT INTO %s SELECT * FROM read_parquet('%s')", tempTableName, targetParquetFile)
		if _, err := conn.ExecContext(context.Background(), loadSQL); err != nil {
			return fmt.Errorf("failed to load existing parquet data: %w", err)
		}
	}

	// Now, copy all data from the temp table to the daily parquet file, overwriting it.
	copySQL := fmt.Sprintf(
		"COPY %s TO '%s' (FORMAT PARQUET)",
		tempTableName, targetParquetFile,
	)
	if _, err := conn.ExecContext(context.Background(), copySQL); err != nil {
		return fmt.Errorf("failed to copy to parquet: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (m *DuckDBManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

// DeleteOldData deletes data from daily parquet files older than the retention period.
func (m *DuckDBManager) DeleteOldData(retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Glob for daily parquet files.
	// Example: m.path/YYYY-MM-DD.parquet
	files, err := filepath.Glob(filepath.Join(m.path, "*.parquet"))
			if err != nil {
				return fmt.Errorf("failed to glob daily parquet files: %w", err)
			}
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		// Extract date from filename (e.g., "YYYY-MM-DD.parquet")
		dateStr := strings.TrimSuffix(fileName, ".parquet")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Skip files with invalid date format in their name.
			log.Printf("Skipping file %s with invalid date format: %v", fileName, err)
			continue
		}

		if fileDate.Before(cutoffDate) {
			if err := os.Remove(filePath); err != nil {
				// Log this error?
				return fmt.Errorf("failed to delete old daily file %s: %w", filePath, err)
			}
		}
	}

	return nil
}