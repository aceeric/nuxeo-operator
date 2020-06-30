package nuxeo

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileNuxeoConf inspects the Nuxeo CR to see if it contains an inline nuxeo.conf, or, clustering is
// enabled. If true, then the function creates a ConfigMap struct to hold the inlined nuxeo.conf and/or the
// clustering config. The content is placed into the ConfigMap identified by the key 'nuxeo.conf'. The function
// then reconciles this with the cluster. The caller must have defined a Volume and VolumeMount elsewhere to
// reference the ConfigMap. (See the handleConfig function for details.) If the Nuxeo CR indicates that an
// inline nuxeo.conf should not exist, then the function makes sure a ConfigMap does not exist in the cluster.
// The ConfigMap is given a hard-coded name: nuxeo cluster name + "-" + node set name + "-nuxeo-conf".
// E.g.: 'my-nuxeo-cluster-nuxeo-conf'.
func reconcileNuxeoConf(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet,
	reqLogger logr.Logger) (reconcile.Result, error) {
	if nodeSet.NuxeoConfig.NuxeoConf.Value != "" || nodeSet.ClusterEnabled {
		expected := r.defaultNuxeoConfCM(instance, nodeSet.Name, nodeSet.NuxeoConfig.NuxeoConf.Value, nodeSet.ClusterEnabled)
		return addOrUpdateConfigMap(r, instance, expected, reqLogger)
	} else {
		cmName := instance.Name + "-" + nodeSet.Name + "-nuxeo-conf"
		return removeConfigMapIfPresent(r, instance, cmName, reqLogger)
	}
}

// defaultNuxeoConfCM generates a ConfigMap struct in a standard internally-defined form to hold the passed
// inline nuxeo conf string data, and/or clustering config settings. The generated struct is configured to be
// owned by the passed 'nux'. A ref to the generated struct is returned.
func (r *ReconcileNuxeo) defaultNuxeoConfCM(nux *v1alpha1.Nuxeo, nodeSetName string,
	nuxeoConf string, clusterEnabled bool) *corev1.ConfigMap {
	cmName := nux.Name + "-" + nodeSetName + "-nuxeo-conf"
	if clusterEnabled {
		if nuxeoConf != "" {
			nuxeoConf += "\n"
		}
		// configureClustering() creates POD_UID. configureClustering will also ensure that a binary
		// storage is configured. The binary storage will create env var NUXEO_BINARY_STORE.
		// See storage.go
		nuxeoConf +=
			"repository.binary.store=${env:NUXEO_BINARY_STORE}\n" +
			"nuxeo.cluster.enabled=true\n" +
			"nuxeo.cluster.nodeid=${env:POD_UID}\n"
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: nux.Namespace,
		},
		Data: map[string]string{"nuxeo.conf": nuxeoConf},
	}
	_ = controllerutil.SetControllerReference(nux, cm, r.scheme)
	return cm
}
