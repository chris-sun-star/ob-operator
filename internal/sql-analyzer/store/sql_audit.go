/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	duckdb "github.com/marcboeker/go-duckdb"
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/const/parquet"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"

	logger "github.com/sirupsen/logrus"
)

type SqlAuditStore struct {
	ctx  context.Context
	db   *sql.DB
	path string
}

func NewSqlAuditStore(c context.Context, path string) (*SqlAuditStore, error) {
	// Use an in-memory DuckDB database for operations.
	db, err := sql.Open("duckdb", "") // In-memory
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory duckdb: %w", err)
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}

	store := &SqlAuditStore{db: db, path: path, ctx: c}
	store.StartCleanupWorker()
	return store, nil
}

func (s *SqlAuditStore) GetLastRequestIDs() (map[string]uint64, error) {
	files, err := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob parquet files: %w", err)
	}

	if len(files) == 0 {
		return make(map[string]uint64), nil
	}

	sort.Slice(files, func(i, j int) bool {
		timeI, errI := parseTimeFromFileName(files[i])
		timeJ, errJ := parseTimeFromFileName(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.Before(timeJ)
	})

	mostRecentFile := files[len(files)-1]

	// Query only the most recent file.
	query := fmt.Sprintf("SELECT svr_ip, MAX(max_request_id) FROM read_parquet('%s') GROUP BY svr_ip", mostRecentFile)

	rows, err := s.db.Query(query)
	if err != nil {
		logger.Printf("Error querying latest parquet file %s: %v", mostRecentFile, err)
		return make(map[string]uint64), nil
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

func (s *SqlAuditStore) InsertBatch(resultsSlices [][]model.SqlAudit) error {
	if len(resultsSlices) == 0 {
		return nil
	}

	conn, err := s.db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	tempTableName := "sql_audit_batch_" + uuid.New().String()[:8] // Use a unique temp table name
	if _, err := conn.ExecContext(context.Background(), fmt.Sprintf(sqlconst.CreateSqlAuditTempTableTemplate, tempTableName)); err != nil {
		return fmt.Errorf("failed to create temp table: %w", err)
	}

	// Use the appender to load data into the temp table.
	err = conn.Raw(func(driverConn any) error {
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

		for _, results := range resultsSlices {
			for _, r := range results {
				err := appender.AppendRow(
					r.SvrIP, r.SvrPort, r.TenantId, r.TenantName, r.UserId, r.UserName, r.DbId, r.DBName, r.SqlId, r.PlanId,
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
					return errors.Wrapf(err, "Failed to append row for SvrIP %s, SvrPort %d. MinRequestId: %d, MaxRequestID: %d", r.SvrIP, r.SvrPort, r.MinRequestId, r.MaxRequestId)
				}
			}
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "Failed to append data")
	}

	// Determine the target Parquet file based on the current timestamp.
	currentTime := time.Now().Format(parquet.FileTimeFormat)
	targetParquetFile := filepath.Join(s.path, fmt.Sprintf("%s-%s.parquet", currentTime, uuid.New().String()[:8]))

	// Now, copy the data from the temp table to the new parquet file.
	copySql := fmt.Sprintf(
		"COPY %s TO '%s' (FORMAT PARQUET)",
		tempTableName, targetParquetFile,
	)
	_, err = conn.ExecContext(s.ctx, copySql)
	return err
}

func (s *SqlAuditStore) Compact() error {
	conn, err := s.db.Conn(s.ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get connection")
	}
	defer conn.Close()

	smallFiles, err := filepath.Glob(filepath.Join(s.path, parquet.SmallFilePattern))
	if err != nil {
		return fmt.Errorf("failed to glob small parquet files: %w", err)
	}

	if len(smallFiles) <= 1 {
		return nil // Nothing to compact
	}

	sort.Slice(smallFiles, func(i, j int) bool {
		timeI, errI := parseTimeFromFileName(smallFiles[i])
		timeJ, errJ := parseTimeFromFileName(smallFiles[j])
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.Before(timeJ)
	})

	filesToCompact := smallFiles[:len(smallFiles)-1]
	lastFileInBatch := filesToCompact[len(filesToCompact)-1]
	timestamp, err := parseTimeFromFileName(lastFileInBatch)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp from file %s: %w", lastFileInBatch, err)
	}

	// Create a temporary table to hold the data from the small files.
	tempTableName := "compaction_table_" + uuid.New().String()[:8]
	createTempTableSql := fmt.Sprintf("CREATE TEMP TABLE %s AS SELECT * FROM read_parquet(['%s'])", tempTableName, strings.Join(filesToCompact, "','"))
	if _, err := conn.ExecContext(context.Background(), createTempTableSql); err != nil {
		return fmt.Errorf("failed to create compaction table from small files: %w", err)
	}

	// Define the compacted file path and a temporary path for atomic operation.
	compactedFile := filepath.Join(s.path, "compacted-"+timestamp.Format(parquet.FileTimeFormat)+".parquet")
	tempCompactedFile := compactedFile + ".tmp"

	// Copy the data from the temporary table to the temporary compacted file.
	copySql := fmt.Sprintf("COPY %s TO '%s' (FORMAT PARQUET)", tempTableName, tempCompactedFile)
	if _, err := conn.ExecContext(context.Background(), copySql); err != nil {
		return fmt.Errorf("failed to copy to temporary compacted file: %w", err)
	}

	// Atomically rename the temporary compacted file to the final name.
	if err := os.Rename(tempCompactedFile, compactedFile); err != nil {
		return fmt.Errorf("failed to rename temporary compacted file: %w", err)
	}

	// Delete the original files that were compacted.
	for _, file := range filesToCompact {
		if err := os.Remove(file); err != nil {
			logger.Printf("Failed to delete old file %s: %v", file, err)
		}
	}

	return nil
}

// QueryOptions holds all the parameters for a dynamic query.
type QueryOptions struct {
	SelectExpressions []string
	Filters           map[string]any
	GroupByColumns    []string
	OrderBy           string
	SortOrder         string
	Limit             int
	Offset            int
}

func (s *SqlAuditStore) CountSqlAudits(opts *QueryOptions) (int64, error) {
	var args []any
	var whereClauses []string

	for key, value := range opts.Filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s ?", key))
		args = append(args, value)
	}

	fromClause := fmt.Sprintf("FROM read_parquet('%s/*.parquet')", s.path)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	groupByClause := ""
	if len(opts.GroupByColumns) > 0 {
		groupByClause = "GROUP BY " + strings.Join(opts.GroupByColumns, ", ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT 1 %s %s %s)", fromClause, whereClause, groupByClause)
	var totalCount int64
	err := s.db.QueryRowContext(s.ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to query total count: %w", err)
	}
	return totalCount, nil
}

func (s *SqlAuditStore) QuerySqlAudits(opts *QueryOptions) ([]map[string]any, error) {
	var args []any
	var whereClauses []string

	for key, value := range opts.Filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s ?", key))
		args = append(args, value)
	}

	fromClause := fmt.Sprintf("FROM read_parquet('%s/*.parquet')", s.path)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	groupByClause := ""
	if len(opts.GroupByColumns) > 0 {
		groupByClause = "GROUP BY " + strings.Join(opts.GroupByColumns, ", ")
	}

	selectClause := "SELECT " + strings.Join(opts.SelectExpressions, ", ")

	var orderByClause string
	if opts.OrderBy != "" {
		safeOrderBy := strings.ReplaceAll(opts.OrderBy, ";", "")
		safeSortOrder := "ASC"
		if strings.ToUpper(opts.SortOrder) == "DESC" {
			safeSortOrder = "DESC"
		}
		orderByClause = fmt.Sprintf("ORDER BY %s %s", safeOrderBy, safeSortOrder)
	}

	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	dataQuery := fmt.Sprintf("%s %s %s %s %s %s", selectClause, fromClause, whereClause, groupByClause, orderByClause, limitClause)

	rows, err := s.db.QueryContext(s.ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audits: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		m := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*any)
			m[colName] = *val
		}
		results = append(results, m)
	}

	return results, nil
}

// Close closes the database connection.
func (s *SqlAuditStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// DeleteOldData deletes data from parquet files older than the retention period.
func (s *SqlAuditStore) DeleteOldData(retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	files, err := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	if err != nil {
		return fmt.Errorf("failed to glob parquet files: %w", err)
	}

	for _, filePath := range files {
		fileTime, err := parseTimeFromFileName(filePath)
		if err != nil {
			logger.Printf("Skipping file %s with invalid date format: %v", filePath, err)
			continue
		}

		if fileTime.Before(cutoffDate) {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete old file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func (s *SqlAuditStore) StartCleanupWorker() {
	// Start the cleanup routine for old data
	retentionStr := os.Getenv("DATA_RETENTION_DAYS")
	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil {
		logger.Fatalf("Invalid or missing DATA_RETENTION_DAYS environment variable: %v", err)
	}

	go func() {
		// Run cleanup once at startup
		logger.Println("Running initial cleanup of old data...")
		if err := s.DeleteOldData(retentionDays); err != nil {
			logger.Printf("Error during initial data cleanup: %v", err)
		}

		// Then run periodically
		cleanupTicker := time.NewTicker(24 * time.Hour)
		defer cleanupTicker.Stop()
		for {
			select {
			case <-cleanupTicker.C:
				logger.Println("Running periodic cleanup of old data...")
				if err := s.DeleteOldData(retentionDays); err != nil {
					logger.Printf("Error during periodic data cleanup: %v", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

func parseTimeFromFileName(fileName string) (time.Time, error) {
	baseName := filepath.Base(fileName)
	var dateStr string
	if strings.HasPrefix(baseName, "compacted-") {
		dateStr = strings.TrimSuffix(strings.TrimPrefix(baseName, "compacted-"), ".parquet")
	} else {
		parts := strings.Split(baseName, "-")
		if len(parts) > 6 {
			dateStr = strings.Join(parts[0:6], "-")
		} else {
			return time.Time{}, fmt.Errorf("invalid small file name format")
		}
	}
	return time.Parse(parquet.FileTimeFormat, dateStr)
}
