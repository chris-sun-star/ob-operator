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

	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	bizconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/sql"
	sql_analyzer_model "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

const (
	SQLMetricConfigFileEnUS = "internal/assets/dashboard/sql_metric_en_US.yaml"
	SQLMetricConfigFileZhCN = "internal/assets/dashboard/sql_metric_zh_CN.yaml"
	SQLMetricScope          = "SQL_DIAGNOSIS"
)

var metricCategoryMap map[string]sql.MetricCategory

func init() {
	metricCategoryMap = make(map[string]sql.MetricCategory)
	metricConfigContent, err := bindata.Asset(SQLMetricConfigFileEnUS)
	if err != nil {
		panic(errors.Wrap(err, "load sql metric config failed"))
	}
	metricConfigMap := make(map[string][]struct {
		Name    string `yaml:"name"`
		Metrics []struct {
			Key      string `yaml:"key"`
			Category string `yaml:"category"`
		} `yaml:"metrics"`
	})
	err = yaml.Unmarshal(metricConfigContent, &metricConfigMap)
	if err != nil {
		panic(errors.Wrap(err, "parse sql metric config data failed"))
	}
	for _, category := range metricConfigMap[SQLMetricScope] {
		for _, metric := range category.Metrics {
			metricCategoryMap[metric.Key] = sql.MetricCategory(category.Name)
		}
	}
}

func ListSqlMetrics(language string) ([]sql.SqlMetricMetaCategory, error) {
	metricClasses := make([]sql.SqlMetricMetaCategory, 0)
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
	metricCategories := make([]sql.SqlMetricMetaCategory, 0)
	err = yaml.Unmarshal(metricConfigContent, &metricCategories)
	if err != nil {
		return metricClasses, err
	}
	logger.Debugf("sql metric configs: %v", metricCategories)
	return metricCategories, err
}

func ListSqlStats(ctx context.Context, filter *sql.SqlFilter) ([]sql.SqlInfo, error) {
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
	sqlInfos := make([]sql.SqlInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		sqlInfo := sql.SqlInfo{
			SqlMetaInfo: sql.SqlMetaInfo{
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
			ExecutionStatistics: []sql.SqlStatisticMetric{},
			LatencyStatistics:   []sql.SqlStatisticMetric{},
		}
		for _, stat := range item.Statistics {
			category, ok := metricCategoryMap[stat.Name]
			if !ok {
				logger.Warnf("metric %s has no category", stat.Name)
				continue
			}
			metric := sql.SqlStatisticMetric{
				Name:  stat.Name,
				Value: stat.Value,
			}
			switch category {
			case sql.Execution:
				sqlInfo.ExecutionStatistics = append(sqlInfo.ExecutionStatistics, metric)
			case sql.Latency:
				sqlInfo.LatencyStatistics = append(sqlInfo.LatencyStatistics, metric)
			case sql.Meta:
				// Do nothing, already populated in SqlMetaInfo
			}
		}
		sqlInfos = append(sqlInfos, sqlInfo)
	}

	return sqlInfos, nil
}
