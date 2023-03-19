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

package core

import (
	"context"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	restoreconst "github.com/oceanbase/ob-operator/pkg/controllers/restore/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RestoreReconciler reconciles a restore object
type RestoreReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme

	Recorder record.EventRecorder
}

type RestoreCtrl struct {
	Resource *resource.Resource
	Restore  cloudv1.Restore
}

type RestoreCtrlOperator interface {
	RestoreCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *RestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.Restore{}
	err := r.CRClient.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// Object not found, return.
			// Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object, requeue the request.
		return reconcile.Result{}, err
	}
	// custom logic
	restoreCtrl := NewRestoreCtrl(r.CRClient, r.Recorder, *instance)
	return restoreCtrl.RestoreCoordinator()
}

func NewRestoreCtrl(client client.Client, recorder record.EventRecorder, restore cloudv1.Restore) RestoreCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &RestoreCtrl{
		Resource: ctrlResource,
		Restore:  restore,
	}
}

func (ctrl *RestoreCtrl) RestoreCoordinator() (ctrl.Result, error) {
	// Restore control-plan
	err := ctrl.RestoreEffector()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (ctrl *RestoreCtrl) RestoreEffector() error {
	if "" == ctrl.Restore.Status.Status {
		klog.Infoln("no restore task, create restore task")
		return ctrl.CreateRestoreTask()
	} else {
		return ctrl.UpdateRestoreStatusFromDB()
	}
}

func (ctrl *RestoreCtrl) CreateRestoreTask() error {
	restoreSpec := ctrl.Restore.Spec
	parameters := restoreSpec.Parameters
	for _, param := range parameters {
		err := ctrl.SetParameter(param)
		if err != nil {
			klog.Errorln(err, "fail to set parameter ", param.Name, param.Value)
			return err
		}
	}
	err, isConcurrencyZero := ctrl.isConcurrencyZero()
	if err != nil {
		klog.Errorln(err, "fail to check restore concurrency")
		return err
	}
	if isConcurrencyZero {
		err = ctrl.SetParameter(cloudv1.Parameter{Name: restoreconst.RestoreConcurrency, Value: strconv.Itoa(restoreconst.RestoreConcurrencyDefault)})
		if err != nil {
			klog.Errorln(err, "fail to set default restore concurrency", restoreconst.RestoreConcurrency)
			return err
		}
	}
	pools, err := ctrl.PrepareForRestore()
	if err != nil {
		klog.Errorln(err, "fail to prepare resource")
		return errors.Wrap(err, "prepare for restore")
	}

	err = ctrl.DoRestore(pools)
	if err != nil {
		klog.Errorln(err, "fail to trigger restore")
		return errors.Wrap(err, "trigger restore")
	}
	err = ctrl.UpdateRestoreStatusFromDB()
	if err != nil {
		klog.Errorln(err, "update restore status failed, set status to init")
		initRestoreStatus := &cloudv1.RestoreStatus{
			Status: restoreconst.RestorePending,
		}
		return ctrl.UpdateRestoreStatus(initRestoreStatus)
	}
	return nil
}

func (ctrl *RestoreCtrl) GetSqlOperator() (*sql.SqlOperator, error) {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.Restore.Namespace, ctrl.Restore.Spec.Source.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	secretName := converter.GenerateSecretNameForDBUser(ctrl.Restore.Spec.Source.ClusterName, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.Restore.Namespace, secretName)
	user := "root"
	password := ""
	if err == nil {
		user = "admin"
		password = string(secret.(corev1.Secret).Data["password"])
	}

	p := &sql.DBConnectProperties{
		IP:       clusterIP,
		Port:     observerconst.MysqlPort,
		User:     user,
		Password: password,
		Database: "oceanbase",
		Timeout:  10,
	}
	so := sql.NewSqlOperator(p)
	if so.TestOK() {
		return so, nil
	}
	return nil, errors.New("failed to get sql operator")
}

func (ctrl *RestoreCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
