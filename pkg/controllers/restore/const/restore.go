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

package restoreconst

const (
	RestoreConcurrencyDefault = 10
	RestoreConcurrencyZero    = 0
	RestoreConcurrency        = "restore_concurrency"

	Locality    = "locality"
	PrimaryZone = "primary_zone"
	KmsEncrypt  = "kms_encrypt"

	ResourceUnitName = "unit_restore"
	ResourcePoolName = "pool_restore"
)

const (
	RestorePending = "RESTORE_PENDING"
	RestoreRunning = "RESTORE_RUNNING"
	RestoreSuccess = "RESTORE_SUCCESS"
	RestoreFail    = "RESTORE_FAIL"
)
