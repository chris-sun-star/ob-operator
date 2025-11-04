package sqlanalyzer

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PlanStore handles operations with the DuckDB database for SQL plans.
type PlanStore struct {
	db   *sql.DB
	path string // path to the duckdb file
}

// NewPlanStore creates a new PlanStore.
func NewPlanStore(path string) (*PlanStore, error) {
	log.Printf("Using plan store at %s", path)
	db, err := sql.Open("duckdb", path)
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
		PRIMARY KEY (TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, ID)
	)`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create sql_plan table: %w", err)
	}

	return &PlanStore{db: db, path: path}, nil
}

// LoadExistingPlans retrieves the identifiers of all plans currently in the database.
func (s *PlanStore) LoadExistingPlans() (map[string]bool, error) {
	query := "SELECT TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID FROM sql_plan"
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
		if err := rows.Scan(&tenantID, &svrIP, &svrPort, &planID); err != nil {
			return nil, fmt.Errorf("failed to scan existing plan: %w", err)
		}
		key := fmt.Sprintf("%d-%s-%d-%d", tenantID, svrIP, svrPort, planID)
		existingPlans[key] = true
	}
	return existingPlans, nil
}

const (
	SQLPlanColumnCount = 39
)

// Store inserts a single SQLPlan data into the database.
func (s *PlanStore) Store(plan SQLPlan) error {
	parsedGmtCreate, err := time.Parse("2006-01-02 15:04:05.000000", plan.GmtCreate)
	if err != nil {
		log.Printf("Error parsing GmtCreate \"%s\": %v. Using zero time.", plan.GmtCreate, err)
		parsedGmtCreate = time.Time{}
	}

	stmt := fmt.Sprintf("INSERT OR IGNORE INTO sql_plan (TENANT_ID, SVR_IP, SVR_PORT, PLAN_ID, SQL_ID, DB_ID, PLAN_HASH, GMT_CREATE, OPERATOR, OBJECT_NODE, OBJECT_ID, OBJECT_OWNER, OBJECT_NAME, OBJECT_ALIAS, OBJECT_TYPE, OPTIMIZER, ID, PARENT_ID, DEPTH, POSITION, COST, REAL_COST, CARDINALITY, REAL_CARDINALITY, IO_COST, CPU_COST, BYTES, ROWSET, OTHER_TAG, PARTITION_START, OTHER, ACCESS_PREDICATES, FILTER_PREDICATES, STARTUP_PREDICATES, PROJECTION, SPECIAL_PREDICATES, QBLOCK_NAME, REMARKS, OTHER_XML) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

	valueArgs := []interface{}{plan.TenantID, plan.SvrIP, plan.SvrPort, plan.PlanID, plan.SQLID, plan.DbID, fmt.Sprintf("%d", plan.PlanHash), parsedGmtCreate,
		plan.Operator, plan.ObjectNode, plan.ObjectID, plan.ObjectOwner, plan.ObjectName, plan.ObjectAlias,
		plan.ObjectType, plan.Optimizer, plan.ID, plan.ParentID, plan.Depth, plan.Position, plan.Cost, plan.RealCost,
		plan.Cardinality, plan.RealCardinality, plan.IoCost, plan.CpuCost, plan.Bytes, plan.Rowset, plan.OtherTag,
		plan.PartitionStart, plan.Other, plan.AccessPredicates, plan.FilterPredicates, plan.StartupPredicates,
		plan.Projection, plan.SpecialPredicates, plan.QblockName, plan.Remarks, plan.OtherXML}


	if _, err := s.db.Exec(stmt, valueArgs...); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection.
func (s *PlanStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
