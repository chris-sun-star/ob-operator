package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"

	"github.com/marcboeker/go-duckdb"
)

// DuckDBManager handles operations with the DuckDB database.
type DuckDBManager struct {
	db *sql.DB
}

// NewDuckDBManager creates a new DuckDBManager instance.
func NewDuckDBManager(dbPath string) (*DuckDBManager, error) {
	connector, err := duckdb.NewConnector(dbPath, nil)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	mgr := &DuckDBManager{db: db}
	if err := mgr.createTable(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// Close closes the DuckDB connection.
func (m *DuckDBManager) Close() error {
	return m.db.Close()
}

func (m *DuckDBManager) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sql_audit_stats (
		CollectTime BIGINT,
		SvrIp VARCHAR,
		TenantId UBIGINT,
		UserId BIGINT,
		DbId UBIGINT,
		SqlId VARCHAR,
		MaxRequestId BIGINT,
		MinRequestTime BIGINT,
		MaxRequestTime BIGINT,
		Executions BIGINT,
		AffectedRows BIGINT,
		ReturnRows BIGINT,
		FailCount BIGINT,
		ElapsedTime BIGINT,
		MaxElapsedTime BIGINT,
		CpuTime BIGINT,
		MaxCpuTime BIGINT,
		NetTime BIGINT,
		NetWaitTime BIGINT,
		QueueTime BIGINT,
		DecodeTime BIGINT,
		GetPlanTime BIGINT,
		ExecuteTime BIGINT,
		ApplicationWaitTime BIGINT,
		ConcurrencyWaitTime BIGINT,
		UserIoWaitTime BIGINT,
		ScheduleTime BIGINT,
		RpcCount BIGINT,
		MissPlanCount BIGINT,
		MemstoreReadRowCount BIGINT,
		SSStoreReadRowCount BIGINT
	);`

	_, err := m.db.Exec(query)
	return err
}

// InsertBatch inserts a batch of SqlAuditResult into DuckDB using the Appender API.
func (m *DuckDBManager) InsertBatch(results []SqlAuditResult) error {
	if len(results) == 0 {
		return nil
	}

	conn, err := m.db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		dc, ok := driverConn.(driver.Conn)
		if !ok {
			return fmt.Errorf("expected driver.Conn but got %T", driverConn)
		}

		appender, err := duckdb.NewAppenderFromConn(dc, "", "sql_audit_stats")
		if err != nil {
			return err
		}
		defer appender.Close()

		for _, res := range results {
			err := appender.AppendRow(
				res.CollectTime, res.SvrIp, res.TenantId, res.UserId, res.DbId, res.SqlId,
				res.MaxRequestId, res.MinRequestTime, res.MaxRequestTime, res.Executions,
				res.AffectedRows, res.ReturnRows, res.FailCount, res.ElapsedTime, res.MaxElapsedTime,
				res.CpuTime, res.MaxCpuTime, res.NetTime, res.NetWaitTime, res.QueueTime, res.DecodeTime,
				res.GetPlanTime, res.ExecuteTime, res.ApplicationWaitTime, res.ConcurrencyWaitTime,
				res.UserIoWaitTime, res.ScheduleTime, res.RpcCount, res.MissPlanCount,
				res.MemstoreReadRowCount, res.SSStoreReadRowCount,
			)
			if err != nil {
				return fmt.Errorf("error appending row: %w", err)
			}
		}

		if err := appender.Flush(); err != nil {
			return fmt.Errorf("error flushing appender: %w", err)
		}
		return nil
	})

	if err == nil {
		log.Printf("Successfully inserted %d records into DuckDB.", len(results))
	}
	return err
}
