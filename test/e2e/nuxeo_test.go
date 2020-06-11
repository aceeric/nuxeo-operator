package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestNuxeo(t *testing.T) {
	nuxeoList := &v1alpha1.NuxeoList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, nuxeoList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	t.Run("nuxeo-group", func(t *testing.T) {
		t.Run("Cluster", NuxeoCluster)
	})
}

func NuxeoCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout,
		RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for nuxeo operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "nuxeo-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	if err = nuxeoScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func nuxeoScaleTest(t *testing.T, f *framework.Framework, ctx *framework.Context) error {
	const nuxeoName = "nuxeo"
	const clusterName = "test-cluster"
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	t.Logf("Creating Nuxeo cluster in %v namespace\n", namespace)
	exampleNuxeo := &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nuxeoName,
			Namespace: namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			//NuxeoImage:      "image-registry.openshift-image-registry.svc.cluster.local:5000/images/nuxeo-operator:0.2.0",
			ImagePullPolicy: corev1.PullAlways,
			Access: v1alpha1.NuxeoAccess{
				Hostname: "z",
			},
			NodeSets: []v1alpha1.NodeSet{{
				Name:        clusterName,
				Replicas:    1,
				Interactive: true,
				PodTemplate: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":         "nuxeo",
							"nuxeoCr":     nuxeoName,
							"interactive": "true",
						},
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: util.NuxeoServiceAccountName,
						Containers: []corev1.Container{{
							Image:           "image-registry.openshift-image-registry.svc.cluster.local:5000/images/nuxeo:10.10",
							ImagePullPolicy: corev1.PullAlways,
							Name:            "nuxeo",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8080,
							}},
							VolumeMounts: []corev1.VolumeMount{},
						}},
						Volumes: []corev1.Volume{},
					},
				},
			}},
		},
	}
	// create the object and add a cleanup function for the new object
	err = f.Client.Create(context.TODO(), exampleNuxeo, &framework.CleanupOptions{TestContext: ctx,
		Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for nuxeo to reach replica count
	deploymentName := nuxeoName + "-" + clusterName
	t.Log("Waiting for Nuxeo to reach 1 replica")
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, deploymentName, 1, retryInterval, timeout)
	if err != nil {
		return err
	}
	err = f.Client.Get(context.TODO(), types.NamespacedName{Name: nuxeoName, Namespace: namespace}, exampleNuxeo)
	if err != nil {
		return err
	}
	exampleNuxeo.Spec.NodeSets[0].Replicas = 2
	t.Log("scaling Nuxeo to 2 replicas")
	err = f.Client.Update(context.TODO(), exampleNuxeo)
	if err != nil {
		return err
	}
	// wait for nuxeo to reach replica count
	t.Log("Waiting for Nuxeo to reach 2 replicas")
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, deploymentName, 2, retryInterval, timeout)
}
