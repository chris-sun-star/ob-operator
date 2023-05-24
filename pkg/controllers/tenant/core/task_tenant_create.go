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
	"fmt"
	"sort"
	"strings"
	"strconv"

	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
	util "github.com/oceanbase/ob-operator/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func (ctrl *TenantCtrl) GetTenantByName() ([]model.Tenant, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
	}
	tenant := sqlOperator.GetTenantByName(ctrl.Tenant.Name)
	return tenant, nil
}

func (ctrl *TenantCtrl) TenantExist(tenantName string) (bool, int, error) {
	klog.Infof("Check Whether The Tenant '%s' Exists", tenantName)
	tenant, err := ctrl.GetTenantByName()
	if err != nil {
		return false, 0, err
	}
	if len(tenant) == 0 {
		return false, 0, nil
	} else {
		return true, int(tenant[0].TenantID), nil
	}
}

func (ctrl *TenantCtrl) PoolExist(poolName string) (bool, int, error) {
	klog.Infof("Check Whether The Resource Pool '%s' Exists", poolName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return false, 0, errors.Wrap(err, "Get Sql Operator Error When Checking Whether The Resource Pool Exists")
	}
	pool := sqlOperator.GetPoolByName(poolName)
	if len(pool) == 0 {
		return false, 0, nil
	} else {
		return true, int(pool[0].ResourcePoolID), nil
	}
}

func (ctrl *TenantCtrl) UnitExist(name string) (error, bool) {
	klog.Infof("Check Whether The Resource Unit '%s' Exists", name)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Checking Whether The Resource Unit Exists"), false
	}
	unitConfig := sqlOperator.GetUnitConfigByName(name)
	if len(unitConfig) == 0 {
		return nil, false
	} else {
		return nil, true
	}
}

func (ctrl *TenantCtrl) GenerateUnitName(name, zoneName string) string {
	unitName := fmt.Sprintf("%s_unit_%s", name, zoneName)
	return unitName
}

func (ctrl *TenantCtrl) GeneratePoolName(name, zoneName string) string {
	poolName := fmt.Sprintf("%s_pool_%s", name, zoneName)
	return poolName
}

func (ctrl *TenantCtrl) CheckResourceEnough(tenantName string, zone v1.TenantReplica) error {
	klog.Infof("Check Tenant '%s' Zone '%s' Reousrce ", tenantName, zone.ZoneName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Checking Reousrce")
	}
	resource := sqlOperator.GetResource(zone)
	if len(resource) == 0 {
		return errors.Errorf("Tenant '%s' cannot get resource", tenantName)
	}
	unitList := sqlOperator.GetUnitList()
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone.ZoneName {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	poolList := sqlOperator.GetPoolList()
	unitConfigList := sqlOperator.GetUnitConfigV3List()
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						resource[0].CPUTotal -= unitConifg.MaxCPU
						resource[0].MemTotal -= unitConifg.MaxMemory
					}
				}
			}
		}
	}
	if zone.ResourceUnits.MaxCPU.AsApproximateFloat64() > resource[0].CPUTotal {
		return errors.New(fmt.Sprintf("Tenant '%s' Zone '%s' CPU Is Not Enough: Need %f, Only %f", tenantName, zone.ZoneName, zone.ResourceUnits.MaxCPU.AsApproximateFloat64(), resource[0].CPUTotal))
	}
	maxMem := zone.ResourceUnits.MemorySize.Value()
	if err != nil {
		return err
	}
	if maxMem > resource[0].MemTotal {
		return errors.New(fmt.Sprintf("Tenant '%s' Zone '%s' Memory Is Not Enough: Need %s, Only %s", tenantName, zone.ZoneName, util.FormatSize(int(maxMem)), util.FormatSize(int(resource[0].MemTotal))))
	}
	return nil
}

func (ctrl *TenantCtrl) GetOBVersion() (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Get OB Version")
	}
	version := sqlOperator.GetVersion()
	if len(version) > 0 && len(version[0].Version) > 0 {
		return version[0].Version, nil
	}
	return "", errors.Errorf("Tenant '%s' get ob version from db failed", ctrl.Tenant.Name)
}

func (ctrl *TenantCtrl) CreateUnit(unitName string, resourceUnit v1.ResourceUnit) error {
	version, err := ctrl.GetOBVersion()
	if err != nil {
		return err
	}
	switch string(version[0]) {
	case tenantconst.Version3:
		return ctrl.CreateUnitV3(unitName, resourceUnit)
	case tenantconst.Version4:
		return ctrl.CreateUnitV4(unitName, resourceUnit)
	}
	return errors.New("no match version for create unit")
}

func (ctrl *TenantCtrl) CreateUnitV4(unitName string, resourceUnit v1.ResourceUnit) error {
	klog.Infof("Create Tenant '%s' Resource Unit '%s' ", ctrl.Tenant.Name, unitName)
	if resourceUnit.MemorySize.Value() == 0 {
		klog.Errorf("Tenant '%s'  resource unit '%s' memorySize cannot be empty", ctrl.Tenant.Name, unitName)
		return errors.Errorf("Tenant '%s'  resource unit '%s' memorySize cannot be empty", ctrl.Tenant.Name, unitName)
	}
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Unit")
	}
	var option string
	if resourceUnit.MinCPU.Value() != 0 {
		option = fmt.Sprint(option, ", min_cpu ", strconv.FormatFloat(resourceUnit.MinCPU.AsApproximateFloat64(), 'f', -1, 64))
	}
	if resourceUnit.LogDiskSize.Value() != 0 {
		option = fmt.Sprint(option, ", log_disk_size ", resourceUnit.LogDiskSize.Value())
	}
	if resourceUnit.MaxIops != 0 {
		option = fmt.Sprint(option, ", max_iops ", resourceUnit.MaxIops)
	}
	if resourceUnit.MinIops != 0 {
		option = fmt.Sprint(option, ", min_iops ", resourceUnit.MinIops)
	}
	if resourceUnit.IopsWeight != 0 {
		option = fmt.Sprint(option, ", iops_weight ", resourceUnit.IopsWeight)
	}
	return sqlOperator.CreateUnitV4(unitName, resourceUnit, option)
}

func (ctrl *TenantCtrl) CreateUnitV3(unitName string, resourceUnit v1.ResourceUnit) error {
	klog.Infof("Create Tenant '%s' Resource Unit '%s' ", ctrl.Tenant.Name, unitName)
	if resourceUnit.MinCPU.Value() == 0 || resourceUnit.MemorySize.Value() == 0 {
		klog.Errorf("Tenant '%s'  resource unit '%s' minCPU & maxMemory & minMemory cannot be empty", ctrl.Tenant.Name, unitName)
		return errors.Errorf("Tenant '%s'  resource unit '%s' minCPU & maxMemory & minMemory cannot be empty", ctrl.Tenant.Name, unitName)
	}
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Unit")
	}
	if resourceUnit.MaxDiskSize.Value() == 0 {
		resourceUnit.MaxDiskSize = resource.MustParse(tenantconst.MaxDiskSize)
	}
	if resourceUnit.MaxIops == 0 {
		resourceUnit.MaxIops = tenantconst.MaxIops
	}
	if resourceUnit.MinIops == 0 {
		resourceUnit.MinIops = tenantconst.MinIops
	}
	if resourceUnit.MaxSessionNum == 0 {
		resourceUnit.MaxSessionNum = tenantconst.MaxSessionNum
	}
	return sqlOperator.CreateUnitV3(unitName, resourceUnit)
}

func (ctrl *TenantCtrl) CreatePool(poolName, unitName string, zone v1.TenantReplica) error {
	klog.Infof("Create Tenant '%s' Resource Pool %s ", ctrl.Tenant.Name, poolName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	return sqlOperator.CreatePool(poolName, unitName, zone)
}

func (ctrl *TenantCtrl) CheckAndCreateUnitAndPool(tenantName string, zone v1.TenantReplica) error {
	unitName := ctrl.GenerateUnitName(tenantName, zone.ZoneName)
	poolName := ctrl.GeneratePoolName(tenantName, zone.ZoneName)
	poolExist, _, err := ctrl.PoolExist(poolName)
	if err != nil {
		klog.Errorf("Check Tenant '%s' Whether The Resource Pool '%s' Exists Error: %s", tenantName, poolName, err)
		return err
	}

	err, unitExist := ctrl.UnitExist(unitName)
	if err != nil {
		klog.Errorf("Check Tenant '%s' Whether The Resource Unit '%s' Exists Error: %s", tenantName, unitName, err)
		return err
	}

	if !unitExist {
		version, err := ctrl.GetOBVersion()
		if err != nil {
			return err
		}
		if string(version[0]) == tenantconst.Version3 {
			err := ctrl.CheckResourceEnough(tenantName, zone)
			if err != nil {
				return err
			}
		}
		err = ctrl.CreateUnit(unitName, zone.ResourceUnits)
		if err != nil {
			klog.Errorf("Create Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
			return err
		}
	}
	if !poolExist {
		err = ctrl.CreatePool(poolName, unitName, zone)
		if err != nil {
			klog.Errorf("Create Tenant '%s' Pool '%s' Error: %s", tenantName, poolName, err)
			return err
		}
	}
	return nil
}

func (ctrl *TenantCtrl) CreateTenant(tenantName string, zones []v1.TenantReplica) error {
	klog.Infof("Create Tenant '%s'", tenantName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	zoneList := ctrl.GenerateSpecZoneList(zones)
	primaryZone := GenerateSpecPrimaryZone(zones)
	poolList := ctrl.GenerateSpecPoolList(tenantName, zones)
	variableList := ctrl.GenerateVariableList(ctrl.Tenant.Spec.ConnectWhiteList)
	charset := tenantconst.Charset
	locality := ctrl.GenerateLocality(zones)
	if locality != "" {
		locality = fmt.Sprintf(", LOCALITY='%s'", locality)
	}
	collate := ctrl.Tenant.Spec.Collate
	if collate != "" {
		collate = fmt.Sprintf(", COLLATE = %s", collate)
	}
	if ctrl.Tenant.Spec.Charset != "" {
		charset = ctrl.Tenant.Spec.Charset
	}
	return sqlOperator.CreateTenant(tenantName, charset, strings.Join(zoneList, "','"), primaryZone, strings.Join(poolList, "','"), locality, collate, variableList)
}

func (ctrl *TenantCtrl) GenerateSpecZoneList(zones []v1.TenantReplica) []string {
	var zoneList []string
	for _, zone := range zones {
		zoneList = append(zoneList, zone.ZoneName)
	}
	return zoneList
}

func (ctrl *TenantCtrl) GenerateStatusZoneList(zones []v1.TenantReplicaStatus) []string {
	var zoneList []string
	for _, zone := range zones {
		zoneList = append(zoneList, zone.ZoneName)
	}
	return zoneList
}

func (ctrl *TenantCtrl) GenerateSpecPoolList(tenantName string, zones []v1.TenantReplica) []string {
	var poolList []string
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(tenantName, zone.ZoneName)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func (ctrl *TenantCtrl) GenerateStatusPoolList(tenantName string, zones []v1.TenantReplicaStatus) []string {
	var poolList []string
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(tenantName, zone.ZoneName)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func GenerateSpecPrimaryZone(zones []v1.TenantReplica) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, zone := range zones {
		zones := zoneMap[zone.Priority]
		zones = append(zones, zone.ZoneName)
		zoneMap[zone.Priority] = zones

	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		for _, zone := range zones {
			primaryZone = fmt.Sprint(primaryZone, zone, ",")
		}
		primaryZone = primaryZone[0 : len(primaryZone)-1]
		primaryZone = fmt.Sprint(primaryZone, ";")
	}
	primaryZone = primaryZone[0 : len(primaryZone)-1]
	primaryZone = fmt.Sprint(primaryZone, ";")
	return primaryZone
}

func GenerateStatusPrimaryZone(zones []v1.TenantReplicaStatus) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, zone := range zones {
		zones := zoneMap[zone.Priority]
		zones = append(zones, zone.ZoneName)
		zoneMap[zone.Priority] = zones
	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		for _, zone := range zones {
			primaryZone = fmt.Sprint(primaryZone, zone, ",")
		}
		primaryZone = primaryZone[0 : len(primaryZone)-1]
		primaryZone = fmt.Sprint(primaryZone, ";")
	}
	primaryZone = primaryZone[0 : len(primaryZone)-1]
	primaryZone = fmt.Sprint(primaryZone, ";")
	return primaryZone
}

func (ctrl *TenantCtrl) GenerateLocality(zones []v1.TenantReplica) string {
	specLocalityMap := GenerateSpecLocalityMap(zones)
	localityList := ctrl.GenerateLocalityList(specLocalityMap)
	return strings.Join(localityList, ",")
}

func (ctrl *TenantCtrl) GenerateVariableList(variable string) string {
	if variable == "" {
		return fmt.Sprintf("SET VARIABLES %s = %s", tenantconst.OBTcpInvitedNodes, tenantconst.DefaultOBTcpInvitedNodes)
	} else {
		return fmt.Sprintf("SET VARIABLES %s = '%s'", tenantconst.OBTcpInvitedNodes, variable)
	}
}
