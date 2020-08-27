package envtest

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Run Controller", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Stub", func() {
		It("Should never fail", func() {
			Expect(1).To(Equal(1))
		})
	})

	Context("Nuxeo Cluster", func() {
		It("Should create successfully", func() {
			By("Create Nuxeo CR")
			nuxeo := &v1alpha1.Nuxeo{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-nuxeo",
					Namespace: "default",
				},
				Spec: v1alpha1.NuxeoSpec{
					NodeSets: []v1alpha1.NodeSet{{
						Name:     "test-cluster",
						Replicas: 1,
					}},
					// the test tools require these even though nil is fine in-cluster
					Volumes:        []corev1.Volume{},
					Containers:     []corev1.Container{},
					InitContainers: []corev1.Container{},
				},
			}
			Expect(k8sClient.Create(context.Background(), nuxeo)).Should(Succeed())

			defer func() {
				Expect(k8sClient.Delete(context.Background(), nuxeo)).Should(Succeed())
			}()

			By("Expect Deployment to be created by the operator")
			Eventually(func() error {
				dep := &appsv1.Deployment{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      "test-nuxeo-test-cluster",
					Namespace: "default",
				}, dep)
			}, time.Second*5, time.Second*1).Should(BeNil())
		})
	})
})
