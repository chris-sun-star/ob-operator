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

package cable

import (
	"errors"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func CableStatusCheck(subsets []cloudv1.SubsetStatus) error {
	for _, subset := range subsets {
		podList := subset.Pods
		for _, pod := range podList {
			podIP := pod.PodIP
			err := CableStatusCheckExecuter(podIP)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func OBServerStart(obCluster cloudv1.OBCluster, subsets []cloudv1.SubsetStatus, rsList, version string) {
	for _, subset := range subsets {
		podList := subset.Pods
		for _, pod := range podList {
			obServerStartArgs := GenerateOBServerStartArgs(obCluster, subset.Name, rsList, version)
			podIP := pod.PodIP
			// check OBServer is already running, for OBServer Scale UP
			err := OBServerStatusCheckExecuter(obCluster.ClusterName, podIP)
			// nil is OBServer is already running
			if err != nil {
				OBServerStartExecuter(podIP, obServerStartArgs)
			}
		}
	}
}

func OBServerGetVersion(podIP string) (string, error) {
	responseData, err := OBServerGetVersionExecuter(podIP)
	if err != nil {
		return "", err
	}
	version := responseData["data"].(string)
	return version, nil
}

func OBServerGetUpgradeRoute(podIP, currentVer, targetVer string) ([]UpgradeRoute, error) {
	obUpgradeRouteArgs := GenerateOBUpgradeRouteArgs(currentVer, targetVer)
	responseData, err := OBServerGetUpgradeRouteExecuter(podIP, obUpgradeRouteArgs)
	if err != nil {
		return nil, err
	}
	response := responseData["data"]
	upgradeRoute, err := GetObUpgradeRouteFromResponse(response)
	if err != nil {
		return upgradeRoute, err
	}
	if len(upgradeRoute) == 0 {
		return upgradeRoute, errors.New("get upgrade route from response Failed")
	}
	return upgradeRoute, nil
}

func OBServerStatusCheck(clusterName string, subsets []cloudv1.SubsetStatus) error {
	for _, subset := range subsets {
		podList := subset.Pods
		for _, pod := range podList {
			podIP := pod.PodIP
			err := OBServerStatusCheckExecuter(clusterName, podIP)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CableReadinessUpdate(subsets []cloudv1.SubsetStatus) error {
	for _, subset := range subsets {
		podList := subset.Pods
		err := CableReadinessUpdateExecuter(podList[0].PodIP)
		if err != nil {
			return err
		}
	}
	return nil
}

func OBRecoverConfig(podIP string) error {
	return OBRecoverConfigExecuter(podIP)
}
