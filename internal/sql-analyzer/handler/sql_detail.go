
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

// @ID GetSqlDetailInfo
// @Summary Get SQL detail info
// @Description Get SQL detail info
// @Tags SQL
// @Accept application/json
// @Produce application/json
// @Param body body model.SqlDetailRequest true "sql detail request"
// @Success 200 {object} model.SqlDetailResponse
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/v1/stats/sql_detail [POST]
func GetSqlDetailInfo(c *gin.Context) (*model.SqlDetailResponse, error) {
	var req model.SqlDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	auditStore, err := store.NewSqlAuditStore(c.Request.Context(), "/data/sql_audit")
	if err != nil {
		return nil, err
	}
	defer auditStore.Close()

	return business.GetSqlDetailInfo(auditStore, req)
}
