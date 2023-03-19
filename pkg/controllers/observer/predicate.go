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

package observer

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type statefulAppPredicate struct {
}

func (p statefulAppPredicate) Create(e event.CreateEvent) bool {
	return true
}

func (p statefulAppPredicate) Delete(e event.DeleteEvent) bool {
	return true
}

func (p statefulAppPredicate) Update(e event.UpdateEvent) bool {
	return true
}

func (p statefulAppPredicate) Generic(e event.GenericEvent) bool {
	return true
}

type tenantPredicate struct {
}

func (p tenantPredicate) Create(e event.CreateEvent) bool {
	return true
}

func (p tenantPredicate) Delete(e event.DeleteEvent) bool {
	return true
}

func (p tenantPredicate) Update(e event.UpdateEvent) bool {
	return true
}

func (p tenantPredicate) Generic(e event.GenericEvent) bool {
	return true
}
