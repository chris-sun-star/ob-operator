package business

import (
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetSqlDetailInfo(auditStore *store.SqlAuditStore, planStore *store.PlanStore, req model.SqlDetailRequest) (*model.SqlDetailResponse, error) {
	return auditStore.QuerySqlDetailInfo(planStore, req)
}
