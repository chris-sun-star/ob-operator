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
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

// @Summary Query SQL statistics
// @Description Query SQL statistics data within a given time range and with optional filters.
// @Tags sql
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant Name"
// @Param request body model.QuerySqlStatsRequest true "Query parameters"
// @Success 200 {object} model.APIResponse{data=map[string]interface{}} "A map containing the query results"
// @Failure 400 {object} model.APIResponse "Error: Invalid request"
// @Failure 500 {object} model.APIResponse "Error: Internal server error"
// @Router /api/v1/tenants/{tenant_name}/sql-stats [post]
func QuerySqlStats(c *gin.Context) (map[string]interface{}, error) {
	tenantName := c.Param("tenant_name")
	var req model.QuerySqlStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// TODO: Implement the logic to query data from DuckDB
	// For now, just return the parsed request and tenant name as confirmation.

	data := gin.H{
		"message":      "Successfully parsed request for SQL statistics",
		"tenant_name":  tenantName,
		"request_body": req,
	}
	return data, nil
}
