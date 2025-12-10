package business

import (
	"context"

	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetSqlDetailInfo(ctx context.Context, cm *oceanbase.ConnectionManager, auditStore *store.SqlAuditStore, planStore *store.PlanStore, req model.SqlDetailRequest) (*model.SqlDetailResponse, error) {
	resp, err := auditStore.QuerySqlDetailInfo(planStore, req)
	if err != nil {
		return nil, err
	}

	// If we have tables and connection manager, query indexes
	if resp != nil && len(resp.Tables) > 0 && cm != nil {
		opMgr, err := cm.GetSysReadonlyConnection()
		if err != nil {
			logger.Warnf("Failed to get sys connection for index query: %v", err)
			// Do not fail the request, just return without indexes
			return resp, nil
		}
		defer opMgr.Close()

		// We need tenantID of the USER tenant, not sys tenant.
		// resp.Plans has TenantID. It should be the same for all plans of the same SQL usually, or at least we pick one.
		var tenantID uint64
		if len(resp.Plans) > 0 {
			tenantID = resp.Plans[0].TenantID
		} else {
			logger.Warn("No plans found, cannot determine tenantID for index query")
			return resp, nil
		}

		for _, table := range resp.Tables {
			indexes, err := oceanbase.QueryTableIndexes(ctx, opMgr, tenantID, table.DatabaseName, table.TableName)
			if err != nil {
				logger.Warnf("Failed to query indexes for table %s.%s: %v", table.DatabaseName, table.TableName, err)
				continue
			}
			resp.Indexes = append(resp.Indexes, indexes...)
		}
	}

	return resp, nil
}
