package business

import (
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetSqlDetailInfo(store *store.SqlAuditStore, req model.SqlDetailRequest) (*model.SqlDetailResponse, error) {
	return store.QuerySqlDetailInfo(req)
}
