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

package resource

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

func (m *OBClusterManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBCluster.Name), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBClusterManager) generateZoneName(zone string) string {
	return fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, zone)
}

func (m *OBClusterManager) generateWaitOBZoneStatusFunc(status string, timeoutSeconds int) func() error {
	f := func() error {
		for i := 1; i < timeoutSeconds; i++ {
			obcluster, err := m.getOBCluster()
			if err != nil {
				return errors.Wrap(err, "get obcluster failed")
			}
			allMatched := true
			for _, obzoneStatus := range obcluster.Status.OBZoneStatus {
				if obzoneStatus.Status != status {
					m.Logger.Info("zone status still not matched", "zone", obzoneStatus.Zone, "status", status)
					allMatched = false
					break
				}
			}
			if allMatched {
				return nil
			}
			time.Sleep(time.Second)
		}
		return errors.New("zone status still not matched when timeout")
	}
	return f
}

func (m *OBClusterManager) CreateOBZone() error {
	m.Logger.Info("create obzones")
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBCluster.APIVersion,
		Kind:       m.OBCluster.Kind,
		Name:       m.OBCluster.Name,
		UID:        m.OBCluster.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	for _, zone := range m.OBCluster.Spec.Topology {
		zoneName := m.generateZoneName(zone.Zone)
		labels := make(map[string]string)
		labels["reference-uid"] = string(m.OBCluster.GetUID())
		labels["reference-cluster"] = m.OBCluster.Name
		obzone := &v1alpha1.OBZone{
			ObjectMeta: metav1.ObjectMeta{
				Name:            zoneName,
				Namespace:       m.OBCluster.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          labels,
			},
			Spec: v1alpha1.OBZoneSpec{
				ClusterName:      m.OBCluster.Spec.ClusterName,
				ClusterId:        m.OBCluster.Spec.ClusterId,
				OBServerTemplate: m.OBCluster.Spec.OBServerTemplate,
				MonitorTemplate:  m.OBCluster.Spec.MonitorTemplate,
				BackupVolume:     m.OBCluster.Spec.BackupVolume,
				Topology:         zone,
			},
		}
		m.Logger.Info("create obzone", "zone", zoneName)
		err := m.Client.Create(m.Ctx, obzone)
		if err != nil {
			m.Logger.Error(err, "create obzone failed", "zone", zone.Zone)
			return errors.Wrap(err, "create obzone")
		}
	}
	return nil
}

func (m *OBClusterManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones failed")
		return nil, errors.Wrap(err, "list obzones")
	}
	m.Logger.Info("successfully get obzone list", "obzone list", obzoneList)
	if len(obzoneList.Items) <= 0 {
		return nil, errors.Wrap(err, "no obzone belongs to this cluster")
	}
	address := obzoneList.Items[0].Status.OBServerStatus[0].Server
	p := connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "", "oceanbase")
	return operation.GetOceanbaseOperationManager(p)
}

func (m *OBClusterManager) Bootstrap() error {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones failed")
		return errors.Wrap(err, "list obzones")
	}
	m.Logger.Info("successfully get obzone list", "obzone list", obzoneList)
	if len(obzoneList.Items) <= 0 {
		return errors.Wrap(err, "no obzone belongs to this cluster")
	}
	address := obzoneList.Items[0].Status.OBServerStatus[0].Server
	p := connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "", "")
	manager, err := operation.GetOceanbaseOperationManager(p)
	if err != nil {
		m.Logger.Error(err, "get oceanbase sql operator failed")
		return errors.Wrap(err, "get oceanbase sql operator")
	}
	m.Logger.Info("successfully get oceanbase sql operator")

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	for _, zone := range obzoneList.Items {
		serverInfo := &model.ServerInfo{
			Ip:   zone.Status.OBServerStatus[0].Server,
			Port: 2882,
		}
		bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
			Region: "default",
			Zone:   zone.Spec.Topology.Zone,
			Server: serverInfo,
		})
	}

	err = manager.Bootstrap(bootstrapServers)
	if err != nil {
		m.Logger.Error(err, "bootstrap failed")
	}
	return err
}

func (m *OBClusterManager) CreateService() error {
	return nil
}

// move to util package
func (m *OBClusterManager) readPassword(secretName string) (string, error) {
	m.Logger.Info("begin get password", "secret", secretName)
	secret := &corev1.Secret{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(secretName), secret)
	if err != nil {
		m.Logger.Error(err, "Get password from secret failed", "secret", secretName)
	}
	return string(secret.Data["password"]), err
}

func (m *OBClusterManager) CreateUsers() error {
	err := m.createUser("operator", m.OBCluster.Spec.UserSecrets.Operator, "all")
	if err != nil {
		return errors.Wrap(err, "Create operator user")
	}
	err = m.createUser("monitor", m.OBCluster.Spec.UserSecrets.Monitor, "select")
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser("proxyro", m.OBCluster.Spec.UserSecrets.ProxyRO, "select")
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser("root", m.OBCluster.Spec.UserSecrets.Root, "all")
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	return nil
}

func (m *OBClusterManager) createUser(userName, secretName, privilege string) error {
	m.Logger.Info("begin create user", "username", userName)
	password, err := m.readPassword(secretName)
	if err != nil {
		return errors.Wrap(err, "Get password from secret failed")
	}
	m.Logger.Info("finish get password", "username", userName, "password", password)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	m.Logger.Info("finish get operationmanager", "username", userName)
	err = oceanbaseOperationManager.CreateUser(userName)
	if err != nil {
		m.Logger.Error(err, "Create user")
		return errors.Wrap(err, "CreateUser")
	}
	m.Logger.Info("finish create user", "username", userName)
	err = oceanbaseOperationManager.SetUserPassword(userName, password)
	if err != nil {
		m.Logger.Error(err, "Set user password")
		return errors.Wrap(err, "Set user password")
	}
	m.Logger.Info("finish set user password", "username", userName)
	object := "*.*"
	err = oceanbaseOperationManager.GrantPrivilege(privilege, object, userName)
	if err != nil {
		m.Logger.Error(err, "Grant privilege")
		return errors.Wrap(err, "GrantPrivilege")
	}
	m.Logger.Info("finish grant user privilege", "username", userName)
	return nil
}

func (m *OBClusterManager) CreateOBParameter() error {
	return nil
}
