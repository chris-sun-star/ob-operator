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
	"errors"
	"fmt"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	tenantBackupconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	util "github.com/oceanbase/ob-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TenantBackupReconciler reconciles a TenantBackup object
type TenantBackupReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme

	Recorder record.EventRecorder
}

type TenantBackupCtrl struct {
	Resource     *resource.Resource
	TenantBackup cloudv1.TenantBackup
}

type TenantBackupCtrlOperator interface {
	TenantBackupCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *TenantBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.TenantBackup{}
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
	tenantBackupCtrl := NewTenantBackupCtrl(r.CRClient, r.Recorder, *instance)

	// Fetch the OBCluster CR instance
	obNamespace := types.NamespacedName{
		Namespace: instance.Spec.SourceCluster.ClusterNamespace,
		Name:      instance.Spec.SourceCluster.ClusterName,
	}
	obInstance := &cloudv1.OBCluster{}
	err = r.CRClient.Get(ctx, obNamespace, obInstance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			klog.Infof("OBCluster %s not found, namespace %s", instance.Spec.SourceCluster.ClusterName, instance.Spec.SourceCluster.ClusterNamespace)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if obInstance.Status.Status != observerconst.ClusterReady {
		klog.Infoln("OBCluster  %s is not ready, namespace %s", instance.Spec.SourceCluster.ClusterName, instance.Spec.SourceCluster.ClusterNamespace)
		return reconcile.Result{}, nil
	}
	// Handle deleted tenant backup
	tenantBackupFinalizerName := fmt.Sprintf("cloud.oceanbase.com.finalizers.%s", instance.Name)
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(instance.ObjectMeta.Finalizers, tenantBackupFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, tenantBackupFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(instance.ObjectMeta.Finalizers, tenantBackupFinalizerName) {
			err := r.TenantBackupDelete(r.CRClient, r.Recorder, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.ObjectMeta.Finalizers = util.RemoveString(instance.ObjectMeta.Finalizers, tenantBackupFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	return tenantBackupCtrl.TenantBackupCoordinator()
}

func (r *TenantBackupReconciler) TenantBackupDelete(client client.Client, recorder record.EventRecorder, tenantBackup *cloudv1.TenantBackup) error {
	ctrlResource := resource.NewResource(client, recorder)
	ctrl := &TenantBackupCtrl{
		TenantBackup: *tenantBackup,
		Resource:     ctrlResource,
	}
    err := ctrl.CancelAllBackupTasks()
    if err != nil {
		klog.Errorf("cancel all backup tasks failed, error '%s'", err)
		return err
    }
	err = ctrl.CancelArchiveLog(tenantBackupconst.TenantAll)
	if err != nil {
		klog.Errorf("tenant '%s' cancel archivelog failed, error '%s'", tenantBackupconst.TenantAll, err)
		return err
	}
	return nil
}

func NewTenantBackupCtrl(client client.Client, recorder record.EventRecorder, tenantBackup cloudv1.TenantBackup) TenantBackupCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &TenantBackupCtrl{
		Resource:     ctrlResource,
		TenantBackup: tenantBackup,
	}
}

func (ctrl *TenantBackupCtrl) TenantBackupCoordinator() (ctrl.Result, error) {
	// TenantBackup control-plan
	err := ctrl.TenantBackupEffector()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (ctrl *TenantBackupCtrl) TenantBackupEffector() error {
	specTenantList := ctrl.TenantBackup.Spec.Tenants
	for _, tenant := range specTenantList {
		err := ctrl.SingleTenantBackupEffector(tenant)
		if err != nil {
			klog.Errorf("tenant '%s' backup failed, error '%s'", tenant.Name, err)
			continue
		}
	}
	cancelTenantList := ctrl.GetCancelTenantList()
	for _, tenant := range cancelTenantList {
		err := ctrl.SingleTenantCancelEffector(tenant)
		if err != nil {
			klog.Errorf("tenant '%s' cancel failed, error '%s'", tenant, err)
			continue
		}
	}
	return nil
}

func (ctrl *TenantBackupCtrl) GetCancelTenantList() []string {
	tenantList := make([]string, 0)
	for _, statusTenant := range ctrl.TenantBackup.Status.TenantBackupSet {
		if statusTenant.TenantName == "" {
			continue
		}
		exist := false
		for _, specTenant := range ctrl.TenantBackup.Spec.Tenants {
			if statusTenant.TenantName == specTenant.Name {
				exist = true
			}
		}
		if !exist {
			tenantList = append(tenantList, statusTenant.TenantName)
		}
	}
	return tenantList
}

func (ctrl *TenantBackupCtrl) SingleTenantBackupEffector(tenant cloudv1.TenantConfigSpec) error {
	exist, backupTypeList := ctrl.CheckTenantBackupExist(tenant)
	if exist {
		backupOnce, finished := ctrl.CheckTenantBackupOnce(tenant, backupTypeList)
		if backupOnce && finished {
			return nil
		}
	}
	return ctrl.SingleTenantBackup(tenant)
}

func (ctrl *TenantBackupCtrl) SingleTenantBackup(tenant cloudv1.TenantConfigSpec) error {
	err := ctrl.CheckAndSetLogArchiveDest(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndStartArchive(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetBackupDest(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetDeletePolicy(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndDoBackup(tenant)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *TenantBackupCtrl) SingleTenantCancelEffector(tenant string) error {
	err := ctrl.CancelArchiveLog(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' cancel archivelog failed, error '%s'", tenant, err)
		return err
	}
	err = ctrl.DeleteSingleTenantStatus(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' delete status failed, error '%s'", tenant, err)
		return err
	}
	return nil
}

func (ctrl *TenantBackupCtrl) GetSqlOperator() (*sql.SqlOperator, error) {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.TenantBackup.Namespace, ctrl.TenantBackup.Spec.SourceCluster.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	secretName := converter.GenerateSecretNameForDBUser(ctrl.TenantBackup.Spec.SourceCluster.ClusterName, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.TenantBackup.Namespace, secretName)
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

func (ctrl *TenantBackupCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
