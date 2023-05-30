package controller

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CronJob controller", func() {

	const (
		OBClusterName      = "test-obcluster"
		OBClusterNamespace = "oceanbase"
	)

	const (
		createTimeout      = 30
		waitRunningTimeout = 300
		interval           = 1
	)

	Context("Create OBCluster", func() {
		It("Should successfully create OBCluster instance and ends with Status running", func() {
			By("By creating a new OBCluster")
			ctx := context.Background()
			obClusterName := "test-1-1"
			obcluster := newMinimalOBCluster(obClusterName, 1, 1)
			Expect(k8sClient.Create(ctx, obcluster)).Should(Succeed())

			obclusterLookupKey := types.NamespacedName{Name: obClusterName, Namespace: DefaultNamespace}
			createdOBCluster := &v1alpha1.OBCluster{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return true
			}, createTimeout, interval).Should(BeTrue())
			Expect(createdOBCluster.Spec.ClusterName).Should(Equal(obClusterName))
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return createdOBCluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
		})
	})
})
