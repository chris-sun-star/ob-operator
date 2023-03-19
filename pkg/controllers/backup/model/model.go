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

package model

type AllBackupSet struct {
	TenantID   int64
	BSKey      int64
	BackupType string
	Status     string
}

type BackupArchiveLogStatus struct {
	TenantID int64
	Status   string
}

type BackupDestValue struct {
	ZoneName string
	SvrIP    string
	SvrPort  int64
	Value    string
}

type BackupSchedule struct {
	BackupType string
	Schedule   string
	NextTime   string
}

type Tenant struct {
	TenantID   int64
	TenantName string
}
