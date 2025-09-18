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
	// The schema is based on the SqlAuditResult struct.
	// Using appropriate DuckDB data types.
	query := `
	CREATE TABLE IF NOT EXISTS sql_audit_stats (
		CollectTime BIGINT,
		SqlId VARCHAR,
		FormatSqlId VARCHAR,
		TenantId UBIGINT,
		EffectiveTenantId UBIGINT,
		TenantName VARCHAR,
		UserId BIGINT,
		UserName VARCHAR,
		DbId UBIGINT,
		DbName VARCHAR,
		QuerySql VARCHAR,
		PlanId BIGINT,
		TraceId VARCHAR,
		MaxRequestId BIGINT,
		MinRequestId BIGINT,
		MaxRequestTime BIGINT,
		MinRequestTime BIGINT,
		MaxCpuTime BIGINT,
		MaxElapsedTime BIGINT,
		Executions BIGINT,
		AffectedRows BIGINT,
		ReturnRows BIGINT,
		PartitionCount BIGINT,
		FailCount BIGINT,
		RetCode4012Count BIGINT,
		RetCode4013Count BIGINT,
		RetCode5001Count BIGINT,
		RetCode5024Count BIGINT,
		RetCode5167Count BIGINT,
		RetCode5217Count BIGINT,
		RetCode6002Count BIGINT,
		Event0WaitTime BIGINT,
		Event1WaitTime BIGINT,
		Event2WaitTime BIGINT,
		Event3WaitTime BIGINT,
		TotalWaitTimeMicro BIGINT,
		TotalWaits BIGINT,
		RpcCount BIGINT,
		PlanTypeLocalCount BIGINT,
		PlanTypeRemoteCount BIGINT,
		PlanTypeDistCount BIGINT,
		InnerSqlCount BIGINT,
		ExecutorRpcCount BIGINT,
		MissPlanCount BIGINT,
		ElapsedTime BIGINT,
		NetTime BIGINT,
		NetWaitTime BIGINT,
		QueueTime BIGINT,
		DecodeTime BIGINT,
		GetPlanTime BIGINT,
		ExecuteTime BIGINT,
		CpuTime BIGINT,
		ApplicationWaitTime BIGINT,
		ConcurrencyWaitTime BIGINT,
		UserIoWaitTime BIGINT,
		ScheduleTime BIGINT,
		RowCacheHit BIGINT,
		BloomFilterCacheHit BIGINT,
		BlockCacheHit BIGINT,
		BlockIndexCacheHit BIGINT,
		DiskReads BIGINT,
		RetryCount BIGINT,
		TableScan BIGINT,
		ConsistencyLevelStrong BIGINT,
		ConsistencyLevelWeak BIGINT,
		MemstoreReadRowCount BIGINT,
		SSStoreReadRowCount BIGINT,
		ExpectedWorkerCount BIGINT,
		UsedWorkerCount BIGINT,
		MemoryUsed BIGINT
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

	// Use conn.Raw to get access to the underlying driver.Conn
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
				res.CollectTime, res.SqlId, res.FormatSqlId, res.TenantId, res.EffectiveTenantId,
				res.TenantName, res.UserId, res.UserName, res.DbId, res.DbName, res.QuerySql,
				res.PlanId, res.TraceId, res.MaxRequestId, res.MinRequestId, res.MaxRequestTime,
				res.MinRequestTime, res.MaxCpuTime, res.MaxElapsedTime, res.Executions, res.AffectedRows,
				res.ReturnRows, res.PartitionCount, res.FailCount, res.RetCode4012Count, res.RetCode4013Count,
				res.RetCode5001Count, res.RetCode5024Count, res.RetCode5167Count, res.RetCode5217Count,
				res.RetCode6002Count, res.Event0WaitTime, res.Event1WaitTime, res.Event2WaitTime,
				res.Event3WaitTime, res.TotalWaitTimeMicro, res.TotalWaits, res.RpcCount,
				res.PlanTypeLocalCount, res.PlanTypeRemoteCount, res.PlanTypeDistCount, res.InnerSqlCount,
				res.ExecutorRpcCount, res.MissPlanCount, res.ElapsedTime, res.NetTime, res.NetWaitTime,
				res.QueueTime, res.DecodeTime, res.GetPlanTime, res.ExecuteTime, res.CpuTime,
				res.ApplicationWaitTime, res.ConcurrencyWaitTime, res.UserIoWaitTime, res.ScheduleTime,
				res.RowCacheHit, res.BloomFilterCacheHit, res.BlockCacheHit, res.BlockIndexCacheHit,
				res.DiskReads, res.RetryCount, res.TableScan, res.ConsistencyLevelStrong,
				res.ConsistencyLevelWeak, res.MemstoreReadRowCount, res.SSStoreReadRowCount,
				res.ExpectedWorkerCount, res.UsedWorkerCount, res.MemoryUsed,
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