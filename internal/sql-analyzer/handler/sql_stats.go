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

package handler

import (
	"github.com/gin-gonic/gin"
	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

// @Summary Query SQL statistics
// @Description Query SQL statistics data within a given time range and with optional filters.
// @Tags sql-analyzer
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant Name"
// @Param request body apimodel.QuerySqlStatsRequest true "Query parameters"
// @Success 200 {object} apimodel.APIResponse{data=apimodel.SqlStatsResponse} "A list of SQL audit statistics"
// @Failure 400 {object} apimodel.APIResponse "Error: Invalid request"
// @Failure 500 {object} apimodel.APIResponse "Error: Internal server error"
// @Router /api/v1/tenants/{tenant_name}/sql-stats [post]
func QuerySqlStats(c *gin.Context) (*apimodel.SqlStatsResponse, error) {
	tenantName := c.Param("tenant_name")
	var req apimodel.QuerySqlStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// TODO: Implement the logic to query data from DuckDB
	// For now, just return an empty response.

	resp := &apimodel.SqlStatsResponse{
		Items:      []apimodel.SqlStatsItem{},
		TotalCount: 0,
	}
	_ = tenantName // to avoid unused variable error for now
	return resp, nil
}
