package sqldatacollector

import (
	"database/sql"
	"fmt"
	"strings"

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
            -- Grouping Keys
            svr_ip VARCHAR, tenant_id BIGINT, tenant_name VARCHAR, user_id BIGINT, user_name VARCHAR,
            db_id BIGINT, db_name VARCHAR, sql_id VARCHAR, query_sql TEXT, plan_id BIGINT,
            client_ip VARCHAR, event VARCHAR, plan_type BIGINT, consistency_level BIGINT,

            -- Aggregated Values
            executions BIGINT, min_request_time BIGINT, max_request_time BIGINT,

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
            fail_count_sum BIGINT,

            PRIMARY KEY (sql_id, svr_ip, min_request_time)
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DuckDBManager{db: db}, nil
}

// GetLastRequestIDs retrieves the last request ID for each server from the database.
func (m *DuckDBManager) GetLastRequestIDs() (map[string]uint64, error) {
	rows, err := m.db.Query("SELECT svr_ip, MAX(max_request_time) FROM sql_audit GROUP BY svr_ip")
	if err != nil {
		// If the table doesn't exist yet on the very first run, return an empty map.
		if strings.Contains(err.Error(), "does not exist") {
			return make(map[string]uint64), nil
		}
		return nil, err
	}
	defer rows.Close()

	lastRequestIDs := make(map[string]uint64)
	for rows.Next() {
		var svrIP string
		var maxRequestTime uint64
		if err := rows.Scan(&svrIP, &maxRequestTime); err != nil {
			return nil, err
		}
		lastRequestIDs[svrIP] = maxRequestTime
	}
	return lastRequestIDs, nil
}

// InsertBatch inserts a batch of SQL audit data into the database.
func (m *DuckDBManager) InsertBatch(results []SQLAudit) error {
	txn, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := txn.Prepare(`INSERT OR IGNORE INTO sql_audit VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range results {
		_, err := stmt.Exec(
			// Grouping Keys
			r.SvrIP, r.TenantId, r.TenantName, r.UserId, r.UserName, r.DbId, r.DBName, r.SqlId, r.QuerySql, r.PlanId, r.ClientIp, r.Event, r.PlanType, r.ConsistencyLevel,
			// Aggregated Values
			r.Executions, r.MinRequestTime, r.MaxRequestTime,
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
			r.FailCountSum,
		)
		if err != nil {
			txn.Rollback()
			return fmt.Errorf("failed to execute statement: %w", err)
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