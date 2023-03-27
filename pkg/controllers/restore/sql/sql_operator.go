/*
Copyright (c) 2021 OceanBase
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
	"fmt"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	"github.com/pkg/errors"
	"k8s.io/klog"
	"strings"
)

type SqlOperator struct {
	ConnectProperties *DBConnectProperties
}

func NewSqlOperator(c *DBConnectProperties) *SqlOperator {
	return &SqlOperator{
		ConnectProperties: c,
	}
}

func (op *SqlOperator) TestOK() bool {
	err := op.ExecSQL("select 1")
	return err == nil
}

func (op *SqlOperator) ExecSQLs(SQLs []string) error {
	client, err := GetDBClient(op.ConnectProperties)
	if err != nil {
		return errors.Wrap(err, "Get DB Connection")
	} else {
		defer client.Close()
		for _, SQL := range SQLs {
			if SQL != "select 1" {
				klog.Infoln(SQL)
			}
			res := client.Exec(SQL)
			if res.Error != nil {
				errNum, errMsg := covertErrToMySQLError(res.Error)
				klog.Errorln(errNum, errMsg)
				return errors.New(errMsg)
			}
		}
	}
	return nil
}

func (op *SqlOperator) ExecSQL(SQL string) error {
	if SQL != "select 1" {
		klog.Infoln(SQL)
	}
	client, err := GetDBClient(op.ConnectProperties)
	if err != nil {
		return errors.Wrap(err, "Get DB Connection")
	} else {
		defer client.Close()
		res := client.Exec(SQL)
		if res.Error != nil {
			errNum, errMsg := covertErrToMySQLError(res.Error)
			klog.Errorln(errNum, errMsg)
			return errors.New(errMsg)
		}
	}
	return nil
}

func (op *SqlOperator) GetVersion() (string, error) {
	res := make([]model.OBVersion, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.OBVersion{}).Raw(GetObVersionSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.OBVersion
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	if len(res) == 0 {
		return "", errors.New("failed to get ob version")
	} else {
		return res[0].Version, nil

	}
}

func (op *SqlOperator) GetAllRestoreHistorySet() []model.RestoreStatus {
	res := make([]model.RestoreStatus, 0)
	version, err := op.GetVersion()
	if err != nil {
		klog.Errorf("get ob version got exception: %v", err)
		return res
	}
	sql := GetRestoreSetHistorySql
	if string(version[0]) != "3" {
		sql = GetRestoreSetHistorySql4
	}
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RestoreStatus{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RestoreStatus
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) GetAllRestoreCurrentSet() []model.RestoreStatus {
	res := make([]model.RestoreStatus, 0)
	version, err := op.GetVersion()
	if err != nil {
		klog.Errorf("get ob version got exception: %v", err)
		return res
	}
	sql := GetRestoreSetCurrentSql
	if string(version[0]) != "3" {
		sql = GetRestoreSetCurrentSql4
	}
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RestoreStatus{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RestoreStatus
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}

func (op *SqlOperator) SetParameter(name, value string) error {
	sql := ReplaceAll(SetParameterTemplate, SetParameterSQLReplacer(name, value))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DoRestore(dest_tenant, source_tenant, dest_path, time, backup_cluster_name, backup_cluster_id, pool_list, restoreOption string, secrets []string) error {
	sqls := make([]string, 0)
	if len(secrets) > 0 {
		setDecryptionSql := fmt.Sprintf(SetDecryptionTemplate, strings.Join(secrets, "', '"))
		sqls = append(sqls, setDecryptionSql)
	}
	doRestoreSql := ReplaceAll(DoRestoreSql, DoRestoreSQLReplacer(dest_tenant, source_tenant, dest_path, time, backup_cluster_name, backup_cluster_id, pool_list, restoreOption))
	sqls = append(sqls, doRestoreSql)
	return op.ExecSQLs(sqls)
}

func (op *SqlOperator) DoRestore4(dest_tenant, dest_path, save_point, backup_cluster_name, backup_cluster_id, pool_list, restoreOption string, secrets []string) error {
	sqls := make([]string, 0)
	if len(secrets) > 0 {
		setDecryptionSql := fmt.Sprintf(SetDecryptionTemplate, strings.Join(secrets, "', '"))
		sqls = append(sqls, setDecryptionSql)
	}
	doRestoreSql := ReplaceAll(DoRestoreSql4, DoRestoreSQLReplacer4(dest_tenant, dest_path, save_point, backup_cluster_name, backup_cluster_id, pool_list, restoreOption))
	sqls = append(sqls, doRestoreSql)
	return op.ExecSQLs(sqls)
}

func (op *SqlOperator) CreateResourceUnit(unit_name, max_cpu, max_memory, max_iops, max_disk_size, max_session_num, min_cpu, min_memory, min_iops string) error {
	sql := ReplaceAll(CreateResourceUnitSql, CreateResourceUnitSQLReplacer(unit_name, max_cpu, max_memory, max_iops, max_disk_size, max_session_num, min_cpu, min_memory, min_iops))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreateResourcePool(pool_name, unit_name, unit_num, zone_list string) error {
	sql := ReplaceAll(CreateResourcePoolSql, CreateResourcePoolSQLReplacer(pool_name, unit_name, unit_num, zone_list))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) GetRestoreConcurrency() []model.RestoreConcurrency {
	res := make([]model.RestoreConcurrency, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RestoreConcurrency{}).Raw(GetRestoreConcurrencySql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RestoreConcurrency
			for rows.Next() {
				err = client.ScanRows(rows, &rowData)
				if err == nil {
					res = append(res, rowData)
				}
			}
		}
	}
	return res
}
