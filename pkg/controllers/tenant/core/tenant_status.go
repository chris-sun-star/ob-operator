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

package core

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/pkg/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (ctrl *TenantCtrl) UpdateTenantStatus(tenantStatus string) error {
	tenant := ctrl.Tenant
	tenantExecuter := resource.NewTenantResource(ctrl.Resource)
	tenantTmp, err := tenantExecuter.Get(context.TODO(), tenant.Namespace, tenant.Name)
	if err != nil {
		return err
	}
	tenantCurrent := tenantTmp.(cloudv1.Tenant)
	tenantCurrentDeepCopy := tenantCurrent.DeepCopy()
	ctrl.Tenant = *tenantCurrentDeepCopy
	tenantNew, err := ctrl.BuildTenantStatus(*tenantCurrentDeepCopy, tenantStatus)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantCurrent.Status, tenantNew.Status)
	if !compareStatus {
		err = tenantExecuter.PatchStatus(context.TODO(), tenantNew, client.MergeFrom(tenantCurrent.DeepCopyObject().(client.Object)))
		if err != nil {
			return err
		}
	}
	ctrl.Tenant = tenantNew
	return nil
}

func (ctrl *TenantCtrl) UpdateTenantStatusOBTcpInvitedNodes(value string) error {
	tenant := ctrl.Tenant
	tenantExecuter := resource.NewTenantResource(ctrl.Resource)
	tenantTmp, err := tenantExecuter.Get(context.TODO(), tenant.Namespace, tenant.Name)
	if err != nil {
		return err
	}
	tenantCurrent := tenantTmp.(cloudv1.Tenant)
	tenantCurrentDeepCopy := tenantCurrent.DeepCopy()
	ctrl.Tenant = *tenantCurrentDeepCopy
	tenantNew, err := ctrl.BuildTenantStatusForVariables(*tenantCurrentDeepCopy, value)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantCurrent.Status, tenantNew.Status)
	if !compareStatus {
		err = tenantExecuter.PatchStatus(context.TODO(), tenantNew, client.MergeFrom(tenantCurrent.DeepCopyObject().(client.Object)))
		if err != nil {
			return err
		}
	}
	ctrl.Tenant = tenantNew
	return nil
}

func (ctrl *TenantCtrl) BuildTenantStatusForVariables(tenant cloudv1.Tenant, value string) (cloudv1.Tenant, error) {
	var tenantCurrentStatus cloudv1.TenantStatus
	tenantTopology, err := ctrl.BuildTenantTopology(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Status = tenant.Status.Status
	tenantCurrentStatus.Topology = tenantTopology
	tenantCurrentStatus.ConnectWhiteList = value
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Charset, err = ctrl.GetCharset()
	if err != nil {
		return tenant, err
	}
	tenant.Status = tenantCurrentStatus
	return tenant, nil
}

func (ctrl *TenantCtrl) BuildTenantStatus(tenant cloudv1.Tenant, tenantStatus string) (cloudv1.Tenant, error) {
	var tenantCurrentStatus cloudv1.TenantStatus
	tenantTopology, err := ctrl.BuildTenantTopology(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Status = tenantStatus
	tenantCurrentStatus.Topology = tenantTopology
	tenantCurrentStatus.ConnectWhiteList = tenant.Status.ConnectWhiteList

	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Charset, err = ctrl.GetCharset()
	if err != nil {
		return tenant, err
	}
	tenant.Status = tenantCurrentStatus
	return tenant, nil
}

func (ctrl *TenantCtrl) BuildTenantTopology(tenant cloudv1.Tenant) ([]cloudv1.TenantReplicaStatus, error) {
	var tenantTopologyStatusList []cloudv1.TenantReplicaStatus
	var err error
	var locality string
	var primaryZone string
	gvTenant, err := ctrl.GetTenantByName()
	if err != nil {
		return tenantTopologyStatusList, err
	}
	if len(gvTenant) == 0 {
		return tenantTopologyStatusList, errors.New(fmt.Sprint("Cannot Get Tenant For BuildTenantTopology: ", ctrl.Tenant.Name))
	}
	locality = gvTenant[0].Locality
	primaryZone = gvTenant[0].PrimaryZone
	typeMap := GenerateTypeMap(locality)
	tenantID := gvTenant[0].TenantID
	priorityMap := GeneratePriorityMap(primaryZone)
	unitNumMap, err := ctrl.GenerateStatusUnitNumMap(tenant.Spec.Topology)
	if err != nil {
		return tenantTopologyStatusList, err
	}
	zoneList, err := ctrl.GenerateStatusZone(tenantID)
	if err != nil {
		return tenantTopologyStatusList, err
	}
	for _, zone := range zoneList {
		var tenantCurrentStatus cloudv1.TenantReplicaStatus
		tenantCurrentStatus.ZoneName = zone
		tenantCurrentStatus.Type = typeMap[zone]
		tenantCurrentStatus.UnitNumber = unitNumMap[zone]
		tenantCurrentStatus.Priority = priorityMap[zone]
		tenantCurrentStatus.ResourceUnits, err = ctrl.BuildResourceUnitFromDB(zone, tenantID)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantCurrentStatus.UnitConfigs, err = ctrl.BuildUnitFromDB(zone, tenantID)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantTopologyStatusList = append(tenantTopologyStatusList, tenantCurrentStatus)
	}
	return tenantTopologyStatusList, nil
}

func (ctrl *TenantCtrl) GenerateStatusZone(tenantID int64) ([]string, error) {
	var zoneList []string
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return zoneList, errors.Wrap(err, "Get Sql Operator Error When Generating Zone For Tenant CR Status")
	}
	poolList := sqlOperator.GetPoolList()

	poolIdMap := make(map[int64]string, 0)
	for _, pool := range poolList {
		if pool.TenantID == tenantID {
			poolIdMap[pool.ResourcePoolID] = pool.Name
		}
	}
	zoneMap := make(map[string]string, 0)
	unitList := sqlOperator.GetUnitList()
	for _, unit := range unitList {
		if poolIdMap[unit.ResourcePoolID] != "" && zoneMap[unit.Zone] == "" {
			zoneMap[unit.Zone] = unit.Zone
		}
	}
	for k, _ := range zoneMap {
		zoneList = append(zoneList, k)
	}
	return zoneList, nil
}

func (ctrl *TenantCtrl) GetCharset() (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Charset")
	}
	charset := sqlOperator.GetCharset()
	return charset[0].Charset, nil
}

func GenerateTypeMap(locality string) map[string]cloudv1.TypeSpec {
	typeMap := make(map[string]cloudv1.TypeSpec, 0)
	typeList := strings.Split(locality, ", ")
	for _, type1 := range typeList {
		split1 := strings.Split(type1, "@")
		typeName := strings.Split(split1[0], "{")[0]
		typeReplica := type1[strings.Index(type1, "{")+1 : strings.Index(type1, "}")]
		replicaInt, _ := strconv.Atoi(typeReplica)
		typeMap[split1[1]] = cloudv1.TypeSpec{
			Name:    typeName,
			Replica: replicaInt,
		}
	}
	return typeMap
}

func GeneratePriorityMap(primaryZone string) map[string]int {
	priorityMap := make(map[string]int, 0)
	cutZones := strings.Split(primaryZone, ";")
	priority := len(cutZones)
	for _, cutZone := range cutZones {
		zoneList := strings.Split(cutZone, ",")
		for _, zone := range zoneList {
			priorityMap[zone] = priority
		}
		priority -= 1
	}
	return priorityMap
}

func (ctrl *TenantCtrl) GenerateStatusUnitNumMap(zones []cloudv1.TenantReplica) (map[string]int, error) {
	unitNumMap := make(map[string]int, 0)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList := sqlOperator.GetPoolList()
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(ctrl.Tenant.Name, zone.ZoneName)
		for _, pool := range poolList {
			if pool.Name == poolName {
				unitNumMap[zone.ZoneName] = int(pool.UnitCount)
			}
		}
	}
	return unitNumMap, nil
}

func (ctrl *TenantCtrl) BuildResourceUnitFromDB(zone string, tenantID int64) (cloudv1.ResourceUnit, error) {
	var resourceUnit cloudv1.ResourceUnit
	version, err := ctrl.GetOBVersion()
	if err != nil {
		return resourceUnit, err
	}
	switch string(version[0]) {
	case tenantconst.Version3:
		return ctrl.BuildResourceUnitV3FromDB(zone, tenantID)
	case tenantconst.Version4:
		return ctrl.BuildResourceUnitV4FromDB(zone, tenantID)
	}
	return resourceUnit, errors.New("no match version for build resource unit from db")
}

func (ctrl *TenantCtrl) BuildResourceUnitV3FromDB(zone string, tenantID int64) (cloudv1.ResourceUnit, error) {
	var resourceUnit cloudv1.ResourceUnit
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return resourceUnit, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	unitList := sqlOperator.GetUnitList()
	poolList := sqlOperator.GetPoolList()
	unitConfigList := sqlOperator.GetUnitConfigV3List()
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) && pool.TenantID == tenantID {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						resourceUnit.MaxCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						resourceUnit.MinCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						resourceUnit.MemorySize = *apiresource.NewQuantity(unitConifg.MaxMemory, apiresource.DecimalSI)
						resourceUnit.MaxDiskSize = *apiresource.NewQuantity(unitConifg.MaxDiskSize, apiresource.DecimalSI)
						resourceUnit.MaxIops = int(unitConifg.MaxIops)
						resourceUnit.MinIops = int(unitConifg.MinIops)
						resourceUnit.MaxSessionNum = int(unitConifg.MaxSessionNum)
					}
				}
			}
		}
	}
	return resourceUnit, nil
}

func (ctrl *TenantCtrl) BuildResourceUnitV4FromDB(zone string, tenantID int64) (cloudv1.ResourceUnit, error) {
	var resourceUnit cloudv1.ResourceUnit
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return resourceUnit, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	unitList := sqlOperator.GetUnitList()
	poolList := sqlOperator.GetPoolList()
	unitConfigList := sqlOperator.GetUnitConfigV4List()
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) && pool.TenantID == tenantID {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						resourceUnit.MaxCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						resourceUnit.MinCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						resourceUnit.MemorySize = *apiresource.NewQuantity(unitConifg.MemorySize, apiresource.DecimalSI)
						resourceUnit.LogDiskSize = *apiresource.NewQuantity(unitConifg.LogDiskSize, apiresource.DecimalSI)
						resourceUnit.MaxIops = int(unitConifg.MaxIops)
						resourceUnit.MinIops = int(unitConifg.MinIops)
						resourceUnit.IopsWeight = int(unitConifg.IopsWeight)
					}
				}
			}
		}
	}
	return resourceUnit, nil
}

func (ctrl *TenantCtrl) BuildUnitFromDB(zone string, tenantID int64) ([]cloudv1.Unit, error) {
	var unitList []cloudv1.Unit
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList := sqlOperator.GetPoolList()
	var resourcePoolIDList []int64
	for _, pool := range poolList {
		if pool.TenantID == tenantID {
			resourcePoolIDList = append(resourcePoolIDList, pool.ResourcePoolID)
		}
	}
	units := sqlOperator.GetUnitList()
	for _, unit := range units {
		for _, poolId := range resourcePoolIDList {
			if unit.Zone == zone && poolId == unit.ResourcePoolID {
				var res cloudv1.Unit
				res.UnitId = int(unit.UnitID)
				res.ServerIP = unit.SvrIP
				res.ServerPort = int(unit.SvrPort)
				res.Status = unit.Status
				var migrateServer cloudv1.MigrateServer
				migrateServer.ServerIP = unit.MigrateFromSvrIP
				migrateServer.ServerPort = int(unit.MigrateFromSvrPort)
				res.Migrate = migrateServer
				unitList = append(unitList, res)
			}
		}
	}
	return unitList, nil
}
