package resource

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
)

func GetOceanbaseOperationManagerFromOBCluster(c client.Client, obcluster *v1alpha1.OBCluster) (*operation.OceanbaseOperationManager, error) {
	obzoneList := &v1alpha1.OBZoneList{}
	err := c.List(context.Background(), obzoneList, client.MatchingLabels{
		"reference-cluster": obcluster.Name,
	}, client.InNamespace(obcluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "Get obzone list")
	}
	if len(obzoneList.Items) <= 0 {
		return nil, errors.Wrap(err, "No obzone belongs to this cluster")
	}
	address := obzoneList.Items[0].Status.OBServerStatus[0].Server

	var p *connector.OceanbaseConnectProperties
	switch obcluster.Status.Status {
	case clusterstatus.New:
		p = connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "", "")
	case clusterstatus.Bootstrapped:
		p = connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "", "oceanbase")
	default:
		// TODO use user operator and read password from secret
		p = connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "", "oceanbase")
	}
	return operation.GetOceanbaseOperationManager(p)
}
