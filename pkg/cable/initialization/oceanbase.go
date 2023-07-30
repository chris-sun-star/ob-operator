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

package initialization

import (
	"github.com/oceanbase/ob-operator/pkg/cable/task/observer"
	"github.com/oceanbase/ob-operator/pkg/util/system"
	log "github.com/sirupsen/logrus"
)

func RecoverIfPossible() {
	configFileExists, err := system.FileExists("/home/admin/oceanbase/store/etc/observer.conf.bin")
	if err != nil {
		log.Error("failed to check oceanbase config file, skip recover process")
	} else if !configFileExists {
		log.Info("check oceanbase config file not exists, skip recover process")
	} else {
		log.Info("check oceanbase config file exists, start recover process")
		// recover by copy the backup config file to etc and start check loop
		observer.RecoverConfig()
		observer.Recover()
	}
}
