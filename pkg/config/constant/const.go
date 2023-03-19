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

package constant

import (
	"time"
)

const (
	AppCable                = "cable"
	ProcessObserver         = "observer"
	CablePort               = 19001
	GracefulTime            = 10 * time.Second
	TickTime                = 5 * time.Second
	StartObserverRetryTimes = 3

	CPU_COUNT                            = 16
	MEMORY_LIMIT                         = 10
	MEMORY_LOW                           = 8
	MEMORY_SIMPLE                        = 64
	NIC                                  = "eth0"
	OBSERVER_MYSQL_PORT                  = 2881
	OBSERVER_RPC_PORT                    = 2882
	OBSERVER_START_COMMAND_WITHOUT_PARAM = "cd /home/admin/oceanbase; ulimit -s 10240; ulimit -c unlimited; LD_LIBRARY_PATH=/home/admin/oceanbase/lib:$LD_LIBRARY_PATH LD_PRELOAD='' /home/admin/oceanbase/bin/observer"
	OBSERVER_START_COMMAND_TEMPLATE      = "cd /home/admin/oceanbase; ulimit -s 10240; ulimit -c unlimited; LD_LIBRARY_PATH=/home/admin/oceanbase/lib:$LD_LIBRARY_PATH LD_PRELOAD='' /home/admin/oceanbase/bin/observer --appname ${OB_CLUSTER_NAME} --cluster_id ${OB_CLUSTER_ID} --zone ${ZONE_NAME} --devname ${DEV_NAME} -p 2881 -P 2882 -d /home/admin/oceanbase/store/ -l info -o 'rootservice_list=${RS_LIST},config_additional_dir=/home/admin/oceanbase/etc2;/home/admin/oceanbase/etc3;/home/admin/oceanbase/store/etc,${OPTION}'"

	OBSERVER_VERSION_COMMAND = "LD_LIBRARY_PATH=/home/admin/oceanbase/lib:$LD_LIBRARY_PATH /home/admin/oceanbase/bin/observer -V"
	CP_CONFIG_COMMAND        = "/usr/bin/cp -rf /home/admin/oceanbase/store/etc/*  /home/admin/oceanbase/etc/"
	MV_CONFIG_COMMAND        = "mv -f /home/admin/oceanbase/etc/observer.conf.bin /home/admin/oceanbase/etc/observer.config.bin"
)
