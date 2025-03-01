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

package oceanbase

var UpgradeEssentialParameters = [...]string{"server_permanent_offline_time", "enable_rebalance", "enable_rereplication"}
var ReservedParameters = [...]string{"cpu_count", "datafile_size", "log_disk_size", "enable_syslog_recycle", "max_syslog_file_count"}

const (
	SqlPort = 2881
	RpcPort = 2882
)

const (
	SqlPortName = "sql"
	RpcPortName = "rpc"
)

const (
	ContainerName                  = "observer"
	InstallPath                    = "/home/admin/oceanbase"
	DataPath                       = "/home/admin/data-file"
	ClogPath                       = "/home/admin/data-log"
	LogPath                        = "/home/admin/log"
	UpgradeHealthCheckerScriptPath = "/home/admin/oceanbase/etc/upgrade_health_checker.py"
	UpgradeCheckerScriptPath       = "/home/admin/oceanbase/etc/upgrade_checker.py"
	UpgradePreScriptPath           = "/home/admin/oceanbase/etc/upgrade_pre.py"
	UpgradePostScriptPath          = "/home/admin/oceanbase/etc/upgrade_post.py"
	BackupPath                     = "/ob-backup"
	DataVolumeSuffix               = "data-file"
	ClogVolumeSuffix               = "data-log"
	LogVolumeSuffix                = "log"
	BackupVolumeSuffix             = "backup"
)

const (
	RootUser      = "root"
	ProxyUser     = "proxyro"
	OperatorUser  = "operator"
	StandbyROUser = "standbyro"
)

const (
	FullType     = "Full"
	ReadonlyType = "Readonly"
	LogonlyType  = "Logonly"
)

const (
	SysTenant       = "sys"
	SysTenantPool   = "sys_pool"
	DefaultDatabase = "oceanbase"
	DefaultRegion   = "default"
)

const (
	OBServerVersionKey     = "observer-version"
	EssentialParametersKey = "essential-parameters"
)

const (
	AllPrivilege    = "all"
	SelectPrivilege = "select"
)

const (
	OceanbaseAllScope = "oceanbase.*"
)

const (
	ClusterNameParam = "cluster"
	ClusterIdParam   = "cluster_id"
)

const (
	CmdVersion                 = "rpm -q --queryformat '%{VERSION}-%{RELEASE}' oceanbase-ce | sed 's/\\.[^.]*$//'"
	CmdUpgradeValidateTemplate = "/home/admin/oceanbase/bin/oceanbase-helper upgrade validate -s %s"
)
