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

package oceanbase

import (
	"context"
	"strings"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

type IndexRow struct {
	IndexName  string `db:"index_name"`
	IndexType  string `db:"index_type"`
	Uniqueness string `db:"uniqueness"`
	Status     string `db:"status"`
	ColumnName string `db:"column_name"`
}

func QueryTableIndexes(ctx context.Context, opMgr *operation.OceanbaseOperationManager, tenantID uint64, dbName, tableName string) ([]model.IndexInfo, error) {
	query := `
    SELECT
      I.index_name,
      I.index_type,
      I.uniqueness,
      I.status,
      GROUP_CONCAT(C.column_name ORDER BY column_position SEPARATOR ',') AS column_name
    FROM cdb_indexes I
    LEFT JOIN cdb_ind_columns C
      ON I.table_owner = C.table_owner
      AND I.table_name = C.table_name
      AND I.index_name = C.index_name
      AND I.con_id = C.con_id
    WHERE I.con_id = ?
      AND I.table_owner = ?
      AND I.table_name = ?
    GROUP BY I.index_name, I.index_type, I.uniqueness, I.status;
    `

	var rows []IndexRow
	err := opMgr.QueryList(ctx, &rows, query, tenantID, dbName, tableName)
	if err != nil {
		return nil, err
	}

	var indexes []model.IndexInfo
	for _, row := range rows {
		var category string
		if strings.HasPrefix(row.IndexName, "t_pk_obpk_") {
			category = "PRIMARY"
		} else {
			category = row.Uniqueness
		}

		indexes = append(indexes, model.IndexInfo{
			TableName: tableName,
			Category:  category,
			IndexName: row.IndexName,
			Columns:   strings.Split(row.ColumnName, ","),
			Status:    row.Status,
		})
	}
	return indexes, nil
}
