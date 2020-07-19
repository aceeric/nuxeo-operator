package nuxeo

import (
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileNuxeoConf inspects the Nuxeo CR to see if it contains an inline nuxeo.conf, or, clustering is
// enabled, or, if the passed backing service-generated nuxeo.conf entries contains anything. If any of these
// are true, then the function creates a ConfigMap struct to hold all nuxeo.conf entries. The content is placed
// into the ConfigMap identified by the key 'nuxeo.conf'. The function then reconciles this with the cluster.
// The caller must have defined a Volume and VolumeMount elsewhere to  reference the ConfigMap. (See the
// handleConfig function for details.) If the Nuxeo CR indicates that an inline nuxeo.conf should not exist,
// then the function makes sure a ConfigMap does not exist in the cluster. The ConfigMap is given a hard-coded
// name: nuxeo cluster name + "-" + node set name + "-nuxeo-conf". E.g.: 'my-nuxeo-cluster-nuxeo-conf'.
func reconcileNuxeoConf(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, nodeSet v1alpha1.NodeSet, backingNuxeoConf string,
	reqLogger logr.Logger) error {
	if shouldReconNuxeoConf(nodeSet, backingNuxeoConf){
		expected := r.defaultNuxeoConfCM(instance, nodeSet.Name, nodeSet.NuxeoConfig.NuxeoConf.Value,
			nodeSet.ClusterEnabled, backingNuxeoConf)
		return addOrUpdateConfigMap(r, instance, expected, reqLogger)
	} else {
		cmName := nuxeoConfCMName(instance, nodeSet.Name)
		return removeConfigMapIfPresent(r, instance, cmName, reqLogger)
	}
}

// Returns true if the Operator should reconcile a nuxeo.conf ConfigMap or Secret to hold nuxeo.conf settings
func shouldReconNuxeoConf(nodeSet v1alpha1.NodeSet, backingNuxeoConf string) bool {
	return nodeSet.NuxeoConfig.NuxeoConf.Value != "" || nodeSet.ClusterEnabled || backingNuxeoConf != ""
}

// defaultNuxeoConfCM generates a ConfigMap struct in a standard internally-defined form to hold the passed
// inline nuxeo conf string data, and/or clustering config settings. The generated struct is configured to be
// owned by the passed 'nux'. A ref to the generated struct is returned.
func (r *ReconcileNuxeo) defaultNuxeoConfCM(nux *v1alpha1.Nuxeo, nodeSetName string,
	inlineNuxeoConf string, clusterEnabled bool, bindingNuxeoConf string) *corev1.ConfigMap {
	cmName := nuxeoConfCMName(nux, nodeSetName)
	clusterNuxeoConf := ""
	if clusterEnabled {
		// configureClustering() creates POD_UID. configureClustering will also ensure that a binary
		// storage is configured. The binary storage will create env var NUXEO_BINARY_STORE.
		// See storage.go
		clusterNuxeoConf =
			"repository.binary.store=${env:NUXEO_BINARY_STORE}\n" +
			"nuxeo.cluster.enabled=true\n" +
			"nuxeo.cluster.nodeid=${env:POD_UID}\n"
	}
	allNuxeoConf := joinCompact("\n", inlineNuxeoConf, clusterNuxeoConf, bindingNuxeoConf)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: nux.Namespace,
		},
		Data: map[string]string{"nuxeo.conf": allNuxeoConf},
	}
	_ = controllerutil.SetControllerReference(nux, cm, r.scheme)
	return cm
}

// standardizes the generation of name for the operator-managed nuxeo.conf ConfigMap
func nuxeoConfCMName(nux *v1alpha1.Nuxeo, nodeSetName string) string {
	return nux.Name + "-" + nodeSetName + "-nuxeo-conf"
}

// Joins together strings like Go strings.Join, but removes leading and trailing whitespace (including
// newlines) from individual components to remove interior whitespace, providing a tidier representation.
func joinCompact(separator string, items ...string) string{
	ret := ""
	for _, str := range items {
		if s := strings.TrimSpace(str); len(s) != 0 {
			// terminate each chunk with a separator. Last chunk also gets a newline since this will
			// be mounted to the filesystem it's natural for the last line in a file to be newline-terminated
			ret += s + separator
		}
	}
	return ret
}
