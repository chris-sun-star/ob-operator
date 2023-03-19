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

package observerconst

// OBCluster Status
const (
	ClusterReady = "Ready"

	TopologyPrepareing = "Prepareing"
	TopologyNotReady   = "Not Ready"
	TopologyReady      = "Ready"

	ResourcePrepareing      = "Resource Prepareing"
	ResourceReady           = "Resource Ready"
	OBServerPrepareing      = "OBServer Prepareing"
	OBServerReady           = "OBServer Ready"
	OBClusterBootstraping   = "OBCluster Bootstraping"
	OBClusterBootstrapReady = "OBCluster Bootstrap Ready"
	OBClusterReady          = "OBCluster Ready"

	OBClusterV3 = "3"
	OBClusterV4 = "4"
)
