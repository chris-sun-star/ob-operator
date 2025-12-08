
package handler

import (
	"github.com/gin-gonic/gin"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	analyticmodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetPlanDetail(c *gin.Context) ([]analyticmodel.SqlPlan, error) {
	var req apimodel.PlanDetailParam
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	planStore, err := store.NewPlanStore(c.Request.Context(), "/data/sql_plan", true)
	if err != nil {
		return nil, err
	}
	defer planStore.Close()

	return business.GetPlanDetail(planStore, req)
}
