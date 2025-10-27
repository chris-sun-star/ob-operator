package sqldatacollector

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PlanStore handles operations with the DuckDB database for SQL plans.
type PlanStore struct {
	db   *sql.DB
	path string // path to the duckdb file
}

// NewPlanStore creates a new PlanStore.
func NewPlanStore(path string) (*PlanStore, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}

	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb file at %s: %w", path, err)
	}

	// Get and log DuckDB version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get DuckDB version: %w", err)
	}
	log.Printf("DuckDB version: %s", version)

	// Create table if not exists
	createTableSQL := `CREATE TABLE IF NOT EXISTS sql_plan (
		TENANT_ID          UBIGINT,
		SVR_IP             VARCHAR,
		SVR_PORT           BIGINT,
		PLAN_ID            BIGINT,
		SQL_ID             VARCHAR,
		DB_ID              BIGINT,
		PLAN_HASH          UBIGINT,
		GMT_CREATE         TIMESTAMP,
		OPERATOR           VARCHAR,
		OBJECT_NODE        VARCHAR,
		OBJECT_ID          BIGINT,
		OBJECT_OWNER       VARCHAR,
		OBJECT_NAME        VARCHAR,
		OBJECT_ALIAS       VARCHAR,
		OBJECT_TYPE        VARCHAR,
		OPTIMIZER          VARCHAR,
		ID                 BIGINT,
		PARENT_ID          BIGINT,
		DEPTH              BIGINT,
		POSITION           BIGINT,
		COST               BIGINT,
		REAL_COST          BIGINT,
		CARDINALITY        BIGINT,
		REAL_CARDINALITY   BIGINT,
		IO_COST            BIGINT,
		CPU_COST           BIGINT,
		BYTES              BIGINT,
		ROWSET             BIGINT,
		OTHER_TAG          VARCHAR,
		PARTITION_START    VARCHAR,
		OTHER              VARCHAR,
		ACCESS_PREDICATES  VARCHAR,
		FILTER_PREDICATES  VARCHAR,
		STARTUP_PREDICATES VARCHAR,
		PROJECTION         VARCHAR,
		SPECIAL_PREDICATES VARCHAR,
		QBLOCK_NAME        VARCHAR,
		REMARKS            VARCHAR,
		OTHER_XML          VARCHAR,
		PRIMARY KEY (TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, PLAN_HASH)
	)`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create sql_plan table: %w", err)
	}

	return &PlanStore{db: db, path: path}, nil
}

// LoadExistingPlans retrieves the identifiers of all plans currently in the database.
func (s *PlanStore) LoadExistingPlans() (map[string]bool, error) {
	query := "SELECT TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, PLAN_HASH FROM sql_plan"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing plans: %w", err)
	}
	defer rows.Close()

	existingPlans := make(map[string]bool)
	for rows.Next() {
		var tenantID uint64
		var svrIP string
		var svrPort int64
		var planID int64
		var planHash uint64
		if err := rows.Scan(&tenantID, &svrIP, &svrPort, &planID, &planHash); err != nil {
			return nil, fmt.Errorf("failed to scan existing plan: %w", err)
		}
		key := fmt.Sprintf("%d-%s-%d-%d-%d", tenantID, svrIP, svrPort, planID, planHash)
		existingPlans[key] = true
	}
	return existingPlans, nil
}

const (
	maxBatchCount = 100
	SQLPlanColumnCount = 39
)

// Store inserts a batch of SQLPlan data into the database.
func (s *PlanStore) Store(plans []SQLPlan) error {
	if len(plans) == 0 {
		return nil
	}

	for i := 0; i < len(plans); i += maxBatchCount {
		end := i + maxBatchCount
		if end > len(plans) {
			end = len(plans)
		}
		batch := plans[i:end]

		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*39)
		for _, p := range batch {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, p.TenantID, p.SvrIP, p.SvrPort, p.PlanID, p.SQLID, p.DbID, p.PlanHash, p.GmtCreate,
				p.Operator, p.ObjectNode, p.ObjectID, p.ObjectOwner, p.ObjectName, p.ObjectAlias,
				p.ObjectType, p.Optimizer, p.ID, p.ParentID, p.Depth, p.Position, p.Cost, p.RealCost,
				p.Cardinality, p.RealCardinality, p.IoCost, p.CpuCost, p.Bytes, p.Rowset, p.OtherTag,
				p.PartitionStart, p.Other, p.AccessPredicates, p.FilterPredicates, p.StartupPredicates,
				p.Projection, p.SpecialPredicates, p.QblockName, p.Remarks, p.OtherXML)
		}

		stmt := fmt.Sprintf("INSERT OR REPLACE INTO sql_plan VALUES %s", strings.Join(valueStrings, ","))

		if _, err := s.db.Exec(stmt, valueArgs...); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database connection.
func (s *PlanStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
