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

const (
	CableUrlProfix          = "http://"
	CablePort               = 19001
	CableNicInfoUrl         = "/api/system/info"
	CableStartUrl           = "/api/ob/start"
	CableStatusUrl          = "/api/ob/status"
	CableReadinessUrl       = "/api/ob/readiness"
	CableReadinessUpdateUrl = "/api/ob/readinessUpdate"
	CableVersionUrl         = "/api/ob/version"
	CableUpgradeRouteUrl    = "/api/ob/upgradeRoute"
	CableRecoverConfigUrl   = "/api/ob/recoverConfig"

	MonagentReadinessUrl = "/metrics/stat"
	MonagentUpdateUrl    = "/api/v1/module/config/update"
)
