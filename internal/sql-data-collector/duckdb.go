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

	// Create the table if it doesn't exist.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sql_audit (
			svr_ip VARCHAR,
			tenant_id BIGINT,
			tenant_name VARCHAR,
			user_name VARCHAR,
			database_name VARCHAR,
			sql_id VARCHAR,
			query_sql TEXT,
			plan_id BIGINT,
			affected_rows BIGINT,
			return_rows BIGINT,
			ret_code BIGINT,
			request_id BIGINT,
			request_time BIGINT,
			elapsed_time BIGINT,
			execute_time BIGINT,
			queue_time BIGINT,
			PRIMARY KEY (svr_ip, request_id)
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DuckDBManager{db: db}, nil
}

// GetLastRequestIDs retrieves the last request ID for each server from the database.
func (m *DuckDBManager) GetLastRequestIDs() (map[string]uint64, error) {
	rows, err := m.db.Query("SELECT svr_ip, MAX(request_id) FROM sql_audit GROUP BY svr_ip")
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

	stmt, err := txn.Prepare(`
		INSERT OR IGNORE INTO sql_audit (svr_ip, tenant_id, tenant_name, user_name, database_name, sql_id, query_sql, plan_id, affected_rows, return_rows, ret_code, request_id, request_time, elapsed_time, execute_time, queue_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range results {
		_, err := stmt.Exec(r.SvrIP, r.TenantID, r.TenantName, r.UserName, r.DatabaseName, r.SQLID, r.QuerySQL, r.PlanID, r.AffectedRows, r.ReturnRows, r.RetCode, r.RequestID, r.RequestTime, r.ElapsedTime, r.ExecuteTime, r.QueueTime)
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
