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

package business

import (
	"fmt"
	"math/big"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	"github.com/sirupsen/logrus"
)

var fixedDimensions = map[string]any{
	"tenant_name":   struct{}{},
	"user_name":     struct{}{},
	"db_name":       struct{}{},
	"sql_id":        struct{}{},
	"plan_id":       struct{}{},
	"format_sql_id": struct{}{},
}

var dimensions = map[string]any{
	"svr_ip":              struct{}{},
	"svr_port":            struct{}{},
	"tenant_id":           struct{}{},
	"user_id":             struct{}{},
	"db_id":               struct{}{},
	"query_sql":           struct{}{},
	"client_ip":           struct{}{},
	"event":               struct{}{},
	"effective_tenant_id": struct{}{},
	"trace_id":            struct{}{},
	"sid":                 struct{}{},
	"user_client_ip":      struct{}{},
	"tx_id":               struct{}{},
	"sub_plan_count":      struct{}{},
	"last_fail_info":      struct{}{},
	"cause_type":          struct{}{},
}

// columnAggregations defines how to aggregate each metric column.
// These are all columns from "executions" onwards in model.SqlAudit.
var columnAggregations = map[string]string{
	"executions":                     "SUM",
	"min_request_time":               "MIN",
	"max_request_time":               "MAX",
	"max_request_id":                 "MAX",
	"min_request_id":                 "MIN",
	"elapsed_time":                   "AVG",
	"elapsed_time_sum":               "SUM",
	"elapsed_time_max":               "MAX",
	"elapsed_time_min":               "MIN",
	"execute_time":                   "AVG",
	"execute_time_sum":               "SUM",
	"execute_time_max":               "MAX",
	"execute_time_min":               "MIN",
	"queue_time":                     "AVG",
	"queue_time_sum":                 "SUM",
	"queue_time_max":                 "MAX",
	"queue_time_min":                 "MIN",
	"get_plan_time":                  "AVG",
	"get_plan_time_sum":              "SUM",
	"get_plan_time_max":              "MAX",
	"get_plan_time_min":              "MIN",
	"affected_rows":                  "AVG",
	"affected_rows_sum":              "SUM",
	"affected_rows_max":              "MAX",
	"affected_rows_min":              "MIN",
	"return_rows":                    "AVG",
	"return_rows_sum":                "SUM",
	"return_rows_max":                "MAX",
	"return_rows_min":                "MIN",
	"partition_count":                "AVG",
	"partition_count_sum":            "SUM",
	"partition_count_max":            "MAX",
	"partition_count_min":            "MIN",
	"retry_count":                    "AVG",
	"retry_count_sum":                "SUM",
	"retry_count_max":                "MAX",
	"retry_count_min":                "MIN",
	"disk_reads":                     "AVG",
	"disk_reads_sum":                 "SUM",
	"disk_reads_max":                 "MAX",
	"disk_reads_min":                 "MIN",
	"rpc_count":                      "AVG",
	"rpc_count_sum":                  "SUM",
	"rpc_count_max":                  "MAX",
	"rpc_count_min":                  "MIN",
	"memstore_read_row_count":        "AVG",
	"memstore_read_row_count_sum":    "SUM",
	"memstore_read_row_count_max":    "MAX",
	"memstore_read_row_count_min":    "MIN",
	"ssstore_read_row_count":         "AVG",
	"ssstore_read_row_count_sum":     "SUM",
	"ssstore_read_row_count_max":     "MAX",
	"ssstore_read_row_count_min":     "MIN",
	"request_memory_used":            "AVG",
	"request_memory_used_sum":        "SUM",
	"request_memory_used_max":        "MAX",
	"request_memory_used_min":        "MIN",
	"wait_time_micro":                "AVG",
	"wait_time_micro_sum":            "SUM",
	"wait_time_micro_max":            "MAX",
	"wait_time_micro_min":            "MIN",
	"total_wait_time_micro":          "AVG",
	"total_wait_time_micro_sum":      "SUM",
	"total_wait_time_micro_max":      "MAX",
	"total_wait_time_micro_min":      "MIN",
	"net_time":                       "AVG",
	"net_time_sum":                   "SUM",
	"net_time_max":                   "MAX",
	"net_time_min":                   "MIN",
	"net_wait_time":                  "AVG",
	"net_wait_time_sum":              "SUM",
	"net_wait_time_max":              "MAX",
	"net_wait_time_min":              "MIN",
	"decode_time":                    "AVG",
	"decode_time_sum":                "SUM",
	"decode_time_max":                "MAX",
	"decode_time_min":                "MIN",
	"application_wait_time":          "AVG",
	"application_wait_time_sum":      "SUM",
	"application_wait_time_max":      "MAX",
	"application_wait_time_min":      "MIN",
	"concurrency_wait_time":          "AVG",
	"concurrency_wait_time_sum":      "SUM",
	"concurrency_wait_time_max":      "MAX",
	"concurrency_wait_time_min":      "MIN",
	"user_io_wait_time":              "AVG",
	"user_io_wait_time_sum":          "SUM",
	"user_io_wait_time_max":          "MAX",
	"user_io_wait_time_min":          "MIN",
	"schedule_time":                  "AVG",
	"schedule_time_sum":              "SUM",
	"schedule_time_max":              "MAX",
	"schedule_time_min":              "MIN",
	"row_cache_hit":                  "AVG",
	"row_cache_hit_sum":              "SUM",
	"row_cache_hit_max":              "MAX",
	"row_cache_hit_min":              "MIN",
	"bloom_filter_cache_hit":         "AVG",
	"bloom_filter_cache_hit_sum":     "SUM",
	"bloom_filter_cache_hit_max":     "MAX",
	"bloom_filter_cache_hit_min":     "MIN",
	"block_cache_hit":                "AVG",
	"block_cache_hit_sum":            "SUM",
	"block_cache_hit_max":            "MAX",
	"block_cache_hit_min":            "MIN",
	"index_block_cache_hit":          "AVG",
	"index_block_cache_hit_sum":      "SUM",
	"index_block_cache_hit_max":      "MAX",
	"index_block_cache_hit_min":      "MIN",
	"expected_worker_count":          "AVG",
	"expected_worker_count_sum":      "SUM",
	"expected_worker_count_max":      "MAX",
	"expected_worker_count_min":      "MIN",
	"used_worker_count":              "AVG",
	"used_worker_count_sum":          "SUM",
	"used_worker_count_max":          "MAX",
	"used_worker_count_min":          "MIN",
	"table_scan":                     "AVG",
	"table_scan_sum":                 "SUM",
	"table_scan_max":                 "MAX",
	"table_scan_min":                 "MIN",
	"consistency_level_strong_count": "SUM",
	"consistency_level_weak_count":   "SUM",
	"cpu_time":                       "AVG",
	"cpu_time_sum":                   "SUM",
	"cpu_time_max":                   "MAX",
	"cpu_time_min":                   "MIN",
	"fail_count_sum":                 "SUM",
	"ret_code_4012_count_sum":        "SUM",
	"ret_code_4013_count_sum":        "SUM",
	"ret_code_5001_count_sum":        "SUM",
	"ret_code_5024_count_sum":        "SUM",
	"ret_code_5167_count_sum":        "SUM",
	"ret_code_5217_count_sum":        "SUM",
	"ret_code_6002_count_sum":        "SUM",
	"event_0_wait_time_sum":          "SUM",
	"event_1_wait_time_sum":          "SUM",
	"event_2_wait_time_sum":          "SUM",
	"event_3_wait_time_sum":          "SUM",
	"plan_type_local_count":          "SUM",
	"plan_type_remote_count":         "SUM",
	"plan_type_distributed_count":    "SUM",
	"inner_sql_count":                "SUM",
	"miss_plan_count":                "SUM",
	"executor_rpc_count":             "SUM",
}

type SqlStatsService struct {
	Store  *store.SqlAuditStore
	Logger *logrus.Logger
}

func NewSqlStatsService(store *store.SqlAuditStore, logger *logrus.Logger) *SqlStatsService {
	return &SqlStatsService{
		Store:  store,
		Logger: logger,
	}
}

func (s *SqlStatsService) QuerySqlStats(req *apimodel.QuerySqlStatsRequest) (*apimodel.SqlStatsResponse, error) {
	filters := s.buildFilters(req)
	s.Logger.Infof("QuerySqlStats filters: %+v", filters)

	selectExpressions, groupByColumns := s.buildQueryParts(req.OutputColumns)

	// Ensure all fixed dimensions are in the SELECT and GROUP BY clauses
	for dim := range fixedDimensions {
		if !contains(groupByColumns, dim) {
			groupByColumns = append(groupByColumns, dim)
		}
		if !contains(selectExpressions, dim) {
			selectExpressions = append(selectExpressions, dim)
		}
	}

	// Create query options for the store
	opts := &store.QueryOptions{
		SelectExpressions: selectExpressions,
		Filters:           filters,
		GroupByColumns:    groupByColumns,
		OrderBy:           req.SortByColumn,
		SortOrder:         req.SortOrder,
		Limit:             req.PageSize,
		Offset:            (req.PageNum - 1) * req.PageSize,
	}

	totalCount, err := s.Store.CountSqlAudits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count sql audits: %w", err)
	}
	s.Logger.Infof("QuerySqlStats totalCount: %d", totalCount)

	if totalCount == 0 {
		return &apimodel.SqlStatsResponse{
			Items:      []apimodel.SqlStatsItem{},
			TotalCount: 0,
		}, nil
	}

	results, err := s.Store.QuerySqlAudits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audits: %w", err)
	}

	items := s.transformResults(results)

	resp := &apimodel.SqlStatsResponse{
		Items:      items,
		TotalCount: totalCount,
	}

	return resp, nil
}

var timeMetrics = map[string]struct{}{
	"elapsed_time":            {},
	"execute_time":            {},
	"queue_time":              {},
	"get_plan_time":           {},
	"wait_time_micro":         {},
	"total_wait_time_micro":   {},
	"net_time":                {},
	"net_wait_time":           {},
	"decode_time":             {},
	"application_wait_time":   {},
	"concurrency_wait_time":   {},
	"user_io_wait_time":       {},
	"schedule_time":           {},
	"event_0_wait_time_sum":   {},
	"event_1_wait_time_sum":   {},
	"event_2_wait_time_sum":   {},
	"event_3_wait_time_sum":   {},
}

func (s *SqlStatsService) buildQueryParts(outputColumns []string) (selectExpressions []string, groupByColumns []string) {
	for _, col := range outputColumns {
		if agg, isMetric := columnAggregations[col]; isMetric {
			columnExpr := fmt.Sprintf("%s(%s)", agg, col)
			if agg == "AVG" {
				columnExpr = fmt.Sprintf("SUM(%s_sum) / SUM(executions)", col)
			}

			if _, isTime := timeMetrics[col]; isTime {
				columnExpr = fmt.Sprintf("(%s) / 1000", columnExpr)
			}

			selectExpressions = append(selectExpressions, fmt.Sprintf("%s as %s", columnExpr, col))
		} else if _, isFixedDimension := fixedDimensions[col]; isFixedDimension {
			// It's a dimension
			selectExpressions = append(selectExpressions, col)
			groupByColumns = append(groupByColumns, col)
		} else if _, isDimension := dimensions[col]; isDimension {
			selectExpressions = append(selectExpressions, fmt.Sprintf("MAX(%s) as %s", col, col))
		} else {
			// do nothing for unknown columns
		}
	}
	return
}

func (s *SqlStatsService) transformResults(results []map[string]any) []apimodel.SqlStatsItem {
	items := make([]apimodel.SqlStatsItem, len(results))
	for i, row := range results {
		item := apimodel.SqlStatsItem{
			Statistics: []apimodel.StatisticItem{},
		}
		for key, val := range row {
			// Populate fixed dimensions
			switch key {
			case "svr_ip":
				item.SvrIP, _ = val.(string)
			case "svr_port":
				item.SvrPort, _ = val.(int64)
			case "tenant_id":
				// DuckDB returns BIGINT as int64, need to convert to uint64
				if v, ok := val.(int64); ok {
					item.TenantId = uint64(v)
				}
			case "tenant_name":
				item.TenantName, _ = val.(string)
			case "user_id":
				item.UserId, _ = val.(int64)
			case "user_name":
				item.UserName, _ = val.(string)
			case "db_id":
				if v, ok := val.(int64); ok {
					item.DBId = uint64(v)
				}
			case "db_name":
				item.DBName, _ = val.(string)
			case "sql_id":
				item.SqlId, _ = val.(string)
			case "plan_id":
				item.PlanId, _ = val.(int64)
			case "query_sql":
				item.QuerySql, _ = val.(string)
			case "client_ip":
				item.ClientIp, _ = val.(string)
			case "event":
				item.Event, _ = val.(string)
			case "format_sql_id":
				item.FormatSqlId, _ = val.(string)
			case "effective_tenant_id":
				if v, ok := val.(int64); ok {
					item.EffectiveTenantId = uint64(v)
				}
			case "trace_id":
				item.TraceId, _ = val.(string)
			case "sid":
				if v, ok := val.(int64); ok {
					item.Sid = uint64(v)
				}
			case "user_client_ip":
				item.UserClientIp, _ = val.(string)
			case "tx_id":
				item.TxId, _ = val.(string)
			case "sub_plan_count":
				item.SubPlanCount, _ = val.(int64)
			case "last_fail_info":
				item.LastFailInfo, _ = val.(int64)
			case "cause_type":
				item.CauseType, _ = val.(int64)
			default:
				// If it's a requested metric, add it to the statistics slice
				var floatVal float64
				switch v := val.(type) {
				case int64:
					floatVal = float64(v)
				case float64:
					floatVal = v
				case uint64:
					floatVal = float64(v)
				case *big.Int:
					if v != nil {
						floatVal, _ = v.Float64()
					}
				default:
					// For now, default to 0.0 if type assertion fails
					floatVal = 0.0
				}
				item.Statistics = append(item.Statistics, apimodel.StatisticItem{
					Name:  key,
					Value: floatVal,
				})
			}
		}
		items[i] = item
	}
	return items
}

func (s *SqlStatsService) buildFilters(req *apimodel.QuerySqlStatsRequest) map[string]interface{} {
	filters := make(map[string]interface{})
	if req.StartTime > 0 {
		// The data in parquet is stored as microseconds, so we need to convert
		filters["max_request_time >="] = req.StartTime * 1000000
	}
	if req.EndTime > 0 {
		filters["min_request_time <="] = req.EndTime * 1000000
	}
	if req.UserName != "" {
		filters["user_name ="] = req.UserName
	}
	if req.DatabaseName != "" {
		filters["db_name ="] = req.DatabaseName
	}
	if req.QuerySqlKeyword != "" {
		filters["query_sql ILIKE"] = "%" + req.QuerySqlKeyword + "%"
	}
	if req.FilterInnerSql {
		filters["inner_sql_count ="] = 0
	}
	return filters
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
