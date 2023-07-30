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

package sql

import (
	"fmt"
	"strconv"
	"strings"
)

func ReplaceAll(template string, replacers ...*strings.Replacer) string {
	s := template
	for _, replacer := range replacers {
		s = replacer.Replace(s)
	}
	return s
}

func SetServerOfflineTimeSQLReplacer(offlineTime int) *strings.Replacer {
	return strings.NewReplacer("${OFFLINE_TIME}", strconv.Itoa(offlineTime))
}

func AddServerSQLReplacer(zoneName, serverIP string) *strings.Replacer {
	return strings.NewReplacer("${SERVER_IP}", serverIP, "${ZONE_NAME}", zoneName)
}

func DelServerSQLReplacer(serverIP string) *strings.Replacer {
	return strings.NewReplacer("${SERVER_IP}", serverIP)
}

func GetRSJobStatusSQLReplacer(serverIP string, port int) *strings.Replacer {
	return strings.NewReplacer("${DELETE_SERVER_IP}", serverIP, "${DELETE_SERVER_PORT}", fmt.Sprintf("%d", port))
}

func CreateUserSQLReplacer(user, password string) *strings.Replacer {
	return strings.NewReplacer("${USER}", user, "${PASSWORD}", password)
}

func GrantPrivilegeSQLReplacer(privilege, object, user string) *strings.Replacer {
	return strings.NewReplacer("${PRIVILEGE}", privilege, "${OBJECT}", object, "${USER}", user)
}

func SetParameterSQLReplacer(name, value string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${VALUE}", value)
}

func SetServerParameterSQLReplacer(server, name, value string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${VALUE}", value, "${SERVER}", server)
}

func GetParameterSQLReplacer(name string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name)
}

func ZoneNameReplacer(zoneName string) *strings.Replacer {
	return strings.NewReplacer("${ZONE_NAME}", zoneName)
}

func GetLeaderCountSQLReplacer(zoneName string) *strings.Replacer {
	return strings.NewReplacer("${ZONE_NAME}", zoneName)
}
