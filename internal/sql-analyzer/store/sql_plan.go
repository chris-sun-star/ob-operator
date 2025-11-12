package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	logger "github.com/sirupsen/logrus"
)

type PlanStore struct {
	ctx  context.Context
	db   *sql.DB
}

func (s *PlanStore) initSqlPlanTable() error {
	// Create table if not exists
	_, err := s.db.Exec(sqlconst.CreateSqlPlanTable)
	return err
}

func NewPlanStore(c context.Context, path string, readOnly bool) (*PlanStore, error) {
	logger.Printf("Using plan store at %s", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}
	db, err := sql.Open("duckdb", filepath.Join(path, "sql_plan.duckdb"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open duckdb at path %s", path)
	}
	s := &PlanStore{db: db, ctx: c}
	err = s.initSqlPlanTable()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *PlanStore) LoadExistingPlans() ([]model.SqlPlanIdentifier, error) {
	rows, err := s.db.Query(sqlconst.ListSqlPlanIdentifier)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query existing plans")
	}
	defer rows.Close()

	existingPlans := make([]model.SqlPlanIdentifier, 0)
	for rows.Next() {
		var tenantID uint64
		var svrIP string
		var svrPort int64
		var planID int64
		if err := rows.Scan(&tenantID, &svrIP, &svrPort, &planID); err != nil {
			return nil, errors.Wrap(err, "failed to scan existing plan")
		}
		existingPlans = append(existingPlans, model.SqlPlanIdentifier{
			TenantID: tenantID,
			SvrIP:    svrIP,
			SvrPort:  svrPort,
			PlanID:   planID,
		})
	}
	return existingPlans, nil
}

func (s *PlanStore) Store(plan model.SqlPlan) error {
	valueArgs := []interface{}{plan.TenantID, plan.SvrIP, plan.SvrPort, plan.PlanID, plan.SqlID, plan.DbID, fmt.Sprintf("%d", plan.PlanHash), plan.GmtCreate,
		plan.Operator, plan.ObjectNode, plan.ObjectID, plan.ObjectOwner, plan.ObjectName, plan.ObjectAlias,
		plan.ObjectType, plan.Optimizer, plan.ID, plan.ParentID, plan.Depth, plan.Position, plan.Cost, plan.RealCost,
		plan.Cardinality, plan.RealCardinality, plan.IoCost, plan.CpuCost, plan.Bytes, plan.Rowset, plan.OtherTag,
		plan.PartitionStart, plan.Other, plan.AccessPredicates, plan.FilterPredicates, plan.StartupPredicates,
		plan.Projection, plan.SpecialPredicates, plan.QblockName, plan.Remarks, plan.OtherXML}

	if _, err := s.db.Exec(sqlconst.StoreSqlPlanStatement, valueArgs...); err != nil {
		return err
	}

	return nil
}

func (s *PlanStore) PlanExists(ident model.SqlPlanIdentifier) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sql_plan WHERE TENANT_ID = ? AND SVR_IP = ? AND SVR_PORT = ? AND PLAN_ID = ?`
	err := s.db.QueryRow(query, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "failed to query plan existence")
	}
	return count > 0, nil
}

// Close closes the database connection.
func (s *PlanStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
