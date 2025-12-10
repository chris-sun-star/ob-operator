/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package sql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	bizconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	dashboard_sql "github.com/oceanbase/ob-operator/internal/dashboard/model/sql"
	sql_analyzer_model "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
)

const (
	SQLMetricConfigFileEnUS = "internal/assets/dashboard/sql_metric_en_US.yaml"
	SQLMetricConfigFileZhCN = "internal/assets/dashboard/sql_metric_zh_CN.yaml"
	SQLMetricScope          = "SQL_DIAGNOSIS"
)

var metricCategoryMap map[string]dashboard_sql.MetricCategory

func init() {
	metricCategoryMap = make(map[string]dashboard_sql.MetricCategory)
	metricConfigContent, err := bindata.Asset(SQLMetricConfigFileEnUS)
	if err != nil {
		panic(errors.Wrap(err, "load sql metric config failed"))
	}
	metricConfigs := make([]dashboard_sql.SqlMetricMetaCategory, 0)
	err = yaml.Unmarshal(metricConfigContent, &metricConfigs)
	if err != nil {
		panic(errors.Wrap(err, "parse sql metric config data failed"))
	}
	for _, category := range metricConfigs {
		for _, metric := range category.Metrics {
			metricCategoryMap[metric.Key] = category.Category
		}
	}
}

func ListSqlMetrics(language string) ([]dashboard_sql.SqlMetricMetaCategory, error) {
	metricClasses := make([]dashboard_sql.SqlMetricMetaCategory, 0)
	configFile := SQLMetricConfigFileEnUS
	switch language {
	case bizconstant.LANGUAGE_EN_US:
		configFile = SQLMetricConfigFileEnUS
	case bizconstant.LANGUAGE_ZH_CN:
		configFile = SQLMetricConfigFileZhCN
	default:
		logger.Infof("Not supported language %s, return default", language)
	}

	metricConfigContent, err := bindata.Asset(configFile)
	if err != nil {
		return metricClasses, err
	}
	metricCategories := make([]dashboard_sql.SqlMetricMetaCategory, 0)
	err = yaml.Unmarshal(metricConfigContent, &metricCategories)
	if err != nil {
		return metricClasses, err
	}
	logger.Debugf("sql metric configs: %v", metricCategories)
	return metricCategories, err
}

func ListSqlStats(ctx context.Context, filter *dashboard_sql.SqlFilter) ([]dashboard_sql.SqlInfo, error) {
	podIP, err := k8s.GetSQLAnalyzerPodIP(ctx, filter.Namespace, filter.OBTenant)
	if err != nil {
		return nil, err
	}

	req := sql_analyzer_model.QuerySqlStatsRequest{
		StartTime:       filter.StartTime,
		EndTime:         filter.EndTime,
		UserName:        filter.User,
		DatabaseName:    filter.Database,
		FilterInnerSql:  !filter.IncludeInnerSql,
		QuerySqlKeyword: filter.Keyword,
		OutputColumns:   filter.OutputColumns,
		SortByColumn:    filter.SortByColumn,
		SortOrder:       filter.SortOrder,
		PageNum:         filter.PageNum,
		PageSize:        filter.PageSize,
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: filter.Namespace,
		Name:      filter.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	resp, err := QuerySqlStats(podIP, obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	// Convert resp to []model.SqlInfo
	sqlInfos := make([]dashboard_sql.SqlInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		sqlInfo := dashboard_sql.SqlInfo{
			SqlMetaInfo: dashboard_sql.SqlMetaInfo{
				SvrIP:      item.SvrIP,
				SvrPort:    item.SvrPort,
				TenantId:   item.TenantId,
				TenantName: item.TenantName,
				UserId:     item.UserId,
				UserName:   item.UserName,
				DBId:       item.DBId,
				DBName:     item.DBName,
				SqlId:      item.SqlId,
				PlanId:     item.PlanId,

				QuerySql:          item.QuerySql,
				ClientIp:          item.ClientIp,
				Event:             item.Event,
				FormatSqlId:       item.FormatSqlId,
				EffectiveTenantId: item.EffectiveTenantId,
				TraceId:           item.TraceId,
				Sid:               item.Sid,
				UserClientIp:      item.UserClientIp,
				TxId:              item.TxId,
				SubPlanCount:      item.SubPlanCount,
				LastFailInfo:      item.LastFailInfo,
				CauseType:         item.CauseType,
			},
			ExecutionStatistics: []dashboard_sql.SqlStatisticMetric{},
			LatencyStatistics:   []dashboard_sql.SqlStatisticMetric{},
		}
		for _, stat := range item.Statistics {
			category, ok := metricCategoryMap[stat.Name]
			if !ok {
				logger.Warnf("metric %s has no category", stat.Name)
				continue
			}
			metric := dashboard_sql.SqlStatisticMetric{
				Name:  stat.Name,
				Value: stat.Value,
			}
			switch category {
			case dashboard_sql.Execution:
				sqlInfo.ExecutionStatistics = append(sqlInfo.ExecutionStatistics, metric)
			case dashboard_sql.Latency:
				sqlInfo.LatencyStatistics = append(sqlInfo.LatencyStatistics, metric)
			case dashboard_sql.Meta:
				// Do nothing, already populated in SqlMetaInfo
			}
		}
		sqlInfos = append(sqlInfos, sqlInfo)
	}

	return sqlInfos, nil
}

func getSysTenantDB(ctx context.Context, namespace, clusterName string) (*sql.DB, error) {
	obCluster, err := clients.GetOBCluster(ctx, namespace, clusterName)
	if err != nil {
		return nil, err
	}

	k8sClient := client.GetClient()
	secretName := obCluster.Spec.UserSecrets.Root
	secret, err := k8sClient.ClientSet.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	password := string(secret.Data[clients.PasswordKey])
	ds := connector.NewOceanBaseDataSource(obCluster.Name, 2881, "root", "sys", password, "oceanbase")
	return sql.Open("mysql", ds.DataSourceName())
}

func QuerySqlDetailInfo(ctx context.Context, param *dashboard_sql.SqlDetailParam) (*dashboard_sql.SqlDetailedInfo, error) {
	podIP, err := k8s.GetSQLAnalyzerPodIP(ctx, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := sql_analyzer_model.SqlDetailRequest{
		StartTime:      param.StartTime,
		EndTime:        param.EndTime,
		SqlId:          param.SqlId,
		Interval:       param.Interval,
		LatencyColumns: param.LatencyColumns,
	}

	resp, err := QuerySqlDetail(podIP, obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	detailedInfo := &dashboard_sql.SqlDetailedInfo{
		ExecutionTrend: []response.MetricData{},
		LatencyTrend:   []response.MetricData{},
		DiagnoseInfo:   []dashboard_sql.SqlDiagnoseInfo{},
		Plans:          []dashboard_sql.PlanStatistic{},
		Indexies:       []dashboard_sql.IndexInfo{},
	}

	// Convert ExecutionTrend
	localTrend := response.MetricData{
		Metric: response.Metric{Name: "local_plan"},
		Values: []response.MetricValue{},
	}
	remoteTrend := response.MetricData{
		Metric: response.Metric{Name: "remote_plan"},
		Values: []response.MetricValue{},
	}
	distributedTrend := response.MetricData{
		Metric: response.Metric{Name: "distributed_plan"},
		Values: []response.MetricValue{},
	}

	for _, trend := range resp.ExecutionTrend {
		ts := float64(trend.Time)
		localTrend.Values = append(localTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Local})
		remoteTrend.Values = append(remoteTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Remote})
		distributedTrend.Values = append(distributedTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Distributed})
	}
	detailedInfo.ExecutionTrend = append(detailedInfo.ExecutionTrend, localTrend, remoteTrend, distributedTrend)

	// Convert LatencyTrend
	latencyTrends := make(map[string]*response.MetricData)
	for _, col := range param.LatencyColumns {
		latencyTrends[col] = &response.MetricData{
			Metric: response.Metric{Name: col},
			Values: []response.MetricValue{},
		}
	}

	for _, item := range resp.LatencyTrend {
		ts := float64(item.Time)
		for col, val := range item.Value {
			if trend, ok := latencyTrends[col]; ok {
				trend.Values = append(trend.Values, response.MetricValue{Timestamp: ts, Value: val})
			}
		}
	}

	for _, trend := range latencyTrends {
		detailedInfo.LatencyTrend = append(detailedInfo.LatencyTrend, *trend)
	}

	// Convert Plans
	for _, planStat := range resp.Plans {
		plan := dashboard_sql.PlanStatistic{
			PlanMeta: dashboard_sql.PlanMeta{
				PlanIdentity: dashboard_sql.PlanIdentity{
					TenantID: planStat.TenantID,
					SvrIP:    planStat.SvrIP,
					SvrPort:  planStat.SvrPort,
					PlanID:   planStat.PlanID,
				},
				PlanHash:      planStat.PlanHash,
				GeneratedTime: planStat.GeneratedTime,
			},
			IoCost:   planStat.IoCost,
			CpuCost:  planStat.CpuCost,
			Cost:     planStat.Cost,
			RealCost: planStat.RealCost,
		}
		detailedInfo.Plans = append(detailedInfo.Plans, plan)
	}

	// Convert Indexes
	if len(resp.Tables) > 0 {
		db, err := getSysTenantDB(ctx, param.Namespace, obtenant.Spec.ClusterName)
		if err != nil {
			logger.Warnf("Failed to connect to sys tenant: %v", err)
		} else {
			defer db.Close()

			for _, table := range resp.Tables {
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
				rows, err := db.QueryContext(ctx, query, obtenant.Status.TenantRecordInfo.TenantID, table.DatabaseName, table.TableName)
				if err != nil {
					logger.Warnf("Failed to query indexes for table %s.%s: %v", table.DatabaseName, table.TableName, err)
					continue
				}

				for rows.Next() {
					var indexName, indexType, uniqueness, status, columns string
					if err := rows.Scan(&indexName, &indexType, &uniqueness, &status, &columns); err != nil {
						logger.Warnf("Failed to scan index row: %v", err)
						continue
					}

					var category dashboard_sql.IndexCategory
					if strings.HasPrefix(indexName, "t_pk_obpk_") {
						category = dashboard_sql.IndexCategoryPrimaryKey
					} else if uniqueness == "UNIQUE" {
						category = dashboard_sql.IndexCategoryGlobalUnique
					} else {
						category = dashboard_sql.IndexCategoryGlobalNormal
					}

					var indexStatus dashboard_sql.IndexStatus
					switch status {
					case "VALID", "AVAILABLE":
						indexStatus = dashboard_sql.IndexStatusAvailable
					case "ERROR", "UNUSABLE":
						indexStatus = dashboard_sql.IndexStatusError
					default:
						indexStatus = dashboard_sql.IndexStatusAvailable // Default to available
					}

					detailedInfo.Indexies = append(detailedInfo.Indexies, dashboard_sql.IndexInfo{
						TableName: table.TableName,
						Category:  category,
						IndexName: indexName,
						Columns:   strings.Split(columns, ","),
						Status:    indexStatus,
					})
				}
				rows.Close()
			}
		}
	}

	return detailedInfo, nil
}

func ListRequestStatistics(c context.Context, param *dashboard_sql.SqlRequestStatisticParam) ([]dashboard_sql.RequestStatisticInfo, error) {
	podIP, err := k8s.GetSQLAnalyzerPodIP(c, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(c, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := sql_analyzer_model.RequestStatisticsRequest{
		StartTime:      param.StartTime,
		EndTime:        param.EndTime,
		UserName:       param.User,
		DatabaseName:   param.Database,
		FilterInnerSql: !param.IncludeInnerSql,
	}

	resp, err := QueryRequestStatistics(podIP, obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []dashboard_sql.RequestStatisticInfo{}, nil
	}

	var averageLatency float64
	if resp.TotalExecutions > 0 {
		averageLatency = resp.TotalLatency / resp.TotalExecutions
	}

	info := dashboard_sql.RequestStatisticInfo{
		Tenant:                 obtenant.Spec.TenantName,
		User:                   param.User,
		Database:               param.Database,
		PlanCategoryStatistics: []dashboard_sql.SqlStatisticMetric{}, // This field is not available from the sql-analyzer
		TotalExecutions:        resp.TotalExecutions,
		FailedExecutions:       resp.FailedExecutions,
		TotalLatency:           resp.TotalLatency,
		AverageLatency:         averageLatency,
		ExecutionTrend:         []response.MetricValue{},
		LatencyTrend:           []response.MetricValue{},
	}

	for _, trend := range resp.ExecutionTrend {
		t, err := time.Parse("2006-01-02", trend.Date)
		if err != nil {
			logger.Errorf("Failed to parse date string %s: %v", trend.Date, err)
			continue
		}
		timestamp := float64(t.Unix())
		info.ExecutionTrend = append(info.ExecutionTrend, response.MetricValue{
			Timestamp: timestamp,
			Value:     trend.Value,
		})
	}

	for _, trend := range resp.LatencyTrend {
		t, err := time.Parse("2006-01-02", trend.Date)
		if err != nil {
			logger.Errorf("Failed to parse date string %s: %v", trend.Date, err)
			continue
		}
		timestamp := float64(t.Unix())
		info.LatencyTrend = append(info.LatencyTrend, response.MetricValue{
			Timestamp: timestamp,
			Value:     trend.Value,
		})
	}

	return []dashboard_sql.RequestStatisticInfo{info}, nil
}

func QueryPlanDetailInfo(ctx context.Context, param *dashboard_sql.PlanDetailParam) (*dashboard_sql.PlanDetail, error) {
	podIP, err := k8s.GetSQLAnalyzerPodIP(ctx, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := model.SqlPlanIdentifier{
		TenantID: param.TenantID,
		SvrIP:    param.SvrIP,
		SvrPort:  param.SvrPort,
		PlanID:   param.PlanID,
	}

	plans, err := QueryPlanDetail(podIP, obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, nil
	}

	// Build plan tree
	planMap := make(map[int64]*dashboard_sql.PlanOperator)
	var root *dashboard_sql.PlanOperator

	for _, plan := range plans {
		planMap[plan.ID] = &dashboard_sql.PlanOperator{
			Operator:      plan.Operator,
			Name:          plan.ObjectName,
			EstimatedRows: int(plan.Cardinality),
			Cost:          plan.Cost,
		}
	}

	for _, plan := range plans {
		if plan.ParentID == 0 {
			root = planMap[plan.ID]
		} else {
			parent, ok := planMap[plan.ParentID]
			if ok {
				parent.ChildOperators = append(parent.ChildOperators, planMap[plan.ID])
			}
		}
	}

	planIdentity := dashboard_sql.PlanIdentity{
		SvrIP:    plans[0].SvrIP,
		SvrPort:  plans[0].SvrPort,
		TenantID: plans[0].TenantID,
		PlanID:   plans[0].PlanID,
	}

	return &dashboard_sql.PlanDetail{
		PlanMeta: dashboard_sql.PlanMeta{
			PlanIdentity: planIdentity,
			PlanHash:     plans[0].PlanHash,
		},
		PlanDetail: root,
	}, nil
}