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
	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
	"github.com/pkg/errors"

	"k8s.io/klog"
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

func (op *SqlOperator) GetGvTenantByName(name string) []model.GvTenant {
	sql := ReplaceAll(GetGvTenantSQL, SetNameReplacer(name))
	res := make([]model.GvTenant, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.GvTenant{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.GvTenant
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

func (op *SqlOperator) GetTenantList() []model.Tenant {
	res := make([]model.Tenant, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Tenant{}).Raw(GetTenantListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Tenant
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

func (op *SqlOperator) GetPoolByName(name string) []model.Pool {
	sql := ReplaceAll(GetPoolSQL, SetNameReplacer(name))
	res := make([]model.Pool, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Pool{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Pool
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

func (op *SqlOperator) GetPoolList() []model.Pool {
	res := make([]model.Pool, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Pool{}).Raw(GetPoolListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Pool
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

func (op *SqlOperator) GetUnitList() []model.Unit {
	res := make([]model.Unit, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Unit{}).Raw(GetUnitListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Unit
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

func (op *SqlOperator) GetUnitConfigList() []model.UnitConfig {
	res := make([]model.UnitConfig, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.UnitConfig{}).Raw(GetUnitConfigListSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.UnitConfig
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

func (op *SqlOperator) GetUnitConfigByName(name string) []model.UnitConfig {
	sql := ReplaceAll(GetUnitConfigSQL, SetNameReplacer(name))
	res := make([]model.UnitConfig, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.UnitConfig{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.UnitConfig
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

func (op *SqlOperator) GetResource(zone v1.TenantReplica) []model.Resource {
	sql := ReplaceAll(GetResourceSQLTemplate, GetResourceSQLReplacer(zone.ZoneName))
	res := make([]model.Resource, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Resource{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Resource
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

func (op *SqlOperator) GetCharset() []model.Charset {
	res := make([]model.Charset, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.Charset{}).Raw(GetCharsetSQL).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.Charset
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

func (op *SqlOperator) GetVariable(name string, tenantID int) []model.SysVariableStat {
	res := make([]model.SysVariableStat, 0)
	sql := ReplaceAll(GetVariableSQLTemplate, GetVariableSQLReplacer(name, tenantID))
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.SysVariableStat{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.SysVariableStat
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

func (op *SqlOperator) GetInprogressJob(name string) []model.RsJob {
	sql := ReplaceAll(GetInprogressJobSQLTemplate, SetNameReplacer(name))
	res := make([]model.RsJob, 0)
	client, err := GetDBClient(op.ConnectProperties)
	if err == nil {
		defer client.Close()
		rows, err := client.Model(&model.RsJob{}).Raw(sql).Rows()
		if err == nil {
			defer rows.Close()
			var rowData model.RsJob
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

func (op *SqlOperator) CreateUnit(name string, resourceUnit v1.ResourceUnit) error {
	sql := ReplaceAll(CreateUnitSQLTemplate, CreateUnitSQLReplacer(name, resourceUnit))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreatePool(poolName, unitName string, zone v1.TenantReplica) error {
	sql := ReplaceAll(CreatePoolSQLTemplate, CreatePoolSQLReplacer(poolName, unitName, zone))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) CreateTenant(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, logonlyReplicaNum, variableList string) error {
	sql := ReplaceAll(CreateTenantSQLTemplate, CreateTenantSQLReplacer(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, logonlyReplicaNum, variableList))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenantVariable(tenantName, name, value string) error {
	sql := ReplaceAll(SetTenantVariableSQLTemplate, SetTenantVariableSQLReplacer(tenantName, name, value))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetUnitConfig(name string, resourceUnit v1.ResourceUnit) error {
	sql := ReplaceAll(SetUnitConfigSQLTemplate, SetUnitConfigSQLReplacer(name, resourceUnit))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetPoolUnitNum(name string, unitNum int) error {
	sql := ReplaceAll(SetPoolUnitNumSQLTemplate, SetPoolUnitNumSQLReplacer(name, unitNum))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenantLocality(name, locality string) error {
	sql := ReplaceAll(SetTenantLocalitySQLTemplate, SetTenantLocalitySQLReplacer(name, locality))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) SetTenant(name, zoneList, primaryZone, poolList, charset, locality, logonlyReplicaNum string) error {
	sql := ReplaceAll(SetTenantSQLTemplate, SetTenantSQLReplacer(name, zoneList, primaryZone, poolList, charset, locality, logonlyReplicaNum))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeleteUnit(name string) error {
	sql := ReplaceAll(DeleteUnitSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeletePool(name string) error {
	sql := ReplaceAll(DeletePoolSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}

func (op *SqlOperator) DeleteTenant(name string) error {
	sql := ReplaceAll(DeleteTenantSQLTemplate, SetNameReplacer(name))
	return op.ExecSQL(sql)
}
