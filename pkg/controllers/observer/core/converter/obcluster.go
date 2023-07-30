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

package converter

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
)

func GetClusterSpecFromOBTopology(topology []cloudv1.Cluster) cloudv1.Cluster {
	var cluster cloudv1.Cluster
	for _, cluster = range topology {
		if cluster.Cluster == myconfig.ClusterName {
			break
		}
	}
	return cluster
}

func GetClusterStatusFromOBTopologyStatus(clusterStatus []cloudv1.ClusterStatus) cloudv1.ClusterStatus {
	var cluster cloudv1.ClusterStatus
	for _, cluster = range clusterStatus {
		if cluster.Cluster == myconfig.ClusterName {
			break
		}
	}
	return cluster
}

func TenantListToTenants(tenants cloudv1.TenantList) []cloudv1.Tenant {
	res := make([]cloudv1.Tenant, 0)
	if len(tenants.Items) > 0 {
		res = tenants.Items
	}
	return res
}

func BackupListToBackups(backups cloudv1.BackupList) []cloudv1.Backup {
	res := make([]cloudv1.Backup, 0)
	if len(backups.Items) > 0 {
		res = backups.Items
	}
	return res
}

func RestoreListToRestores(restores cloudv1.RestoreList) []cloudv1.Restore {
	res := make([]cloudv1.Restore, 0)
	if len(restores.Items) > 0 {
		res = restores.Items
	}
	return res
}
