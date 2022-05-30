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

package obproxy

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/oceanbase/ob-operator/pkg/util/shell"
	"github.com/oceanbase/ob-operator/pkg/util/str"
	"github.com/oceanbase/ob-operator/pkg/util/system"
)

const (
	PROCESS_OBPROXY                = "obproxy"
	OBPROXY_DEFAULT_PORT           = 2883
	OBPROXY_DEFAULT_EXPORTER_PORT  = 2884
	OBPROXY_START_COMMAND_TEMPLATE = "cd /home/admin/obproxy; /home/admin/obproxy/bin/obproxy -p 2883 -n ${NAME} -c ${OB_CLUSTER} -r'${RS_LIST}' -o prometheus_sync_interval=1,prometheus_listen_port=2884,enable_metadb_used=false,skip_proxy_sys_private_check=true,log_dir_size_threshold=10G,enable_proxy_scramble=true,enable_strict_kernel_release=false"
)

var ProcessStarted bool

var Liveness bool
var Readiness bool

var CheckStatusOnce sync.Once

type StartObproxyProcessParam struct {
	Name      string `json:"name" binding:"required"`
	ObCluster string `json:"obCluster" binding:"required"`
	RsList    string `json:"rsList" binding:"required"`
}

func StartObproxyProcess(param *StartObproxyProcessParam) {
	cmd := str.ReplaceAll(OBPROXY_START_COMMAND_TEMPLATE, paramReplacer(param.Name, param.ObCluster, param.RsList))
	_, err := shell.NewCommand(cmd).WithContext(context.TODO()).WithUser(shell.AdminUser).Execute()
	if err != nil {
		log.Println("cmd exec error", err)
	}
}

func paramReplacer(name, obCluster, rsList string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${OB_CLUSTER}", obCluster, "${RS_LIST}", rsList)
}

func CheckStatusLoop() {
	pm := &system.ProcessManager{}
	for {
		time.Sleep(2 * time.Second)
		processRunning := pm.ProcessIsRunningByName(PROCESS_OBPROXY)
		if !processRunning {
			log.Printf("process %s is not running", PROCESS_OBPROXY)
			// obproxy is down, exit to restart a container
			system.Exit()
		}
	}
}

func StopProcess() {
	pm := &system.ProcessManager{}
	err := pm.TerminateProcessByName(PROCESS_OBPROXY)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(2 * time.Second)
	err = pm.KillProcessByName(PROCESS_OBPROXY)
	if err != nil {
		log.Println(err)
	}
	ProcessStarted = false
}
