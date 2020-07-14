package nuxeo

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/jsonpath"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
)

// config map resource
//   note     - the inline handler and this need to cooperate on the contents of the configmap!
//              right now, reconcileNuxeoConf replaces it every time!
//   solution - CM should look like:
//              data
//                inline:
//                backingSvc:
//                nuxeo.conf
//              Each entity only touches their own then - the reconciler creates nuxeo.conf by concatenating
//               inline and backingSvc!
//              Therefore, reconcileNuxeoConf has to be called last after deployment is configured AND backing
//               services are configured,
//              reconcileNuxeoConf has to move inside node set reconciler
// volume and mounts in deployment
// secondary secret
// env vars
// volume and mounts for secondary secret
// returns nuxeo conf to caller!
func configureBackingServices(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, dep *appsv1.Deployment,
	reqLogger logr.Logger) (string, error) {
	nuxeoConf := ""
	for _, backingService := range instance.Spec.BackingServices {
		if err := configureBackingService(r, instance, backingService, dep, reqLogger); err != nil {
			return "", err
		}
		// accumulate each backing service's nuxeo.conf settings
		nuxeoConf = strings.TrimSpace(strings.Join([]string{nuxeoConf, backingService.NuxeoConf}, "\n"))
	}
	return nuxeoConf, nil
}

// configures one backing service. Iterates all resources and bindings, calls helpers to add environment variables
// and mounts into the nuxeo container in the passed deployment. May create a secondary secret if needed. Coalesces
// nuxeo.conf settings and reconciles those settings to the nuxeo.conf ConfigMap or Secret.
func configureBackingService(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, backingService v1alpha1.BackingService,
	dep *appsv1.Deployment, reqLogger logr.Logger) error {
	secondarySecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-backing-" + backingService.Name,
			Namespace: instance.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}
	for _, resource := range backingService.Resources {
		gvk := strings.ToLower(resource.Group + "." + resource.Version + "." + resource.Kind)
		// validating the projections here ensures that the switch statement below works
		if !projectionsAreValid(gvk, resource.Projections) {
			return goerrors.New("backing service resource " + resource.Name + " has invalid projections")
		}
		for i := 0; i < len(resource.Projections); i++ {
			var err error
			projection := resource.Projections[i]
			switch {
			case (isSecret(resource) || isConfigMap(resource)) && projection.Env != "":
				err = projectEnvFrom(resource, i, dep)
			case projection.Mount != "":
				err = projectMount(r, instance.Namespace, backingService.Name, resource, i, dep, &secondarySecret)
			case projection.Transform != (v1alpha1.CertTransform{}):
				err = projectTransform(backingService.Name, resource, i, dep, &secondarySecret)
			default:
				err = goerrors.New(fmt.Sprintf("no handler for projection at ordinal position %v in resource %s", i, resource.Name))
			}
			if err != nil {
				return err
			}
		}
	}
	return reconcileSecondary(r, instance, &secondarySecret, reqLogger)
}

// Adds an environment variable with a valueFrom that references the key in the passed resource, which must be a
// Secret or ConfigMap. Returns non-nil error if: passed resource is not a Secret or ConfigMap, or environment
// variable name is not unique in the nuxeo container. Otherwise nil error is returned and an environment variable
// is added to the nuxeo container in the passed deployment like:
//   env:
//   - name: ELASTIC_PASSWORD              # from projection.Env
//     valueFrom:
//       secretKeyRef:
//         key: elastic                    # from projection.Key
//         name: elastic-es-elastic-user   # from resource.Name
func projectEnvFrom(resource v1alpha1.BackingServiceResource, idx int, dep *appsv1.Deployment) error {
	projection := resource.Projections[idx]
	env := corev1.EnvVar{
		Name: projection.Env,
	}
	if isSecret(resource) {
		env.ValueFrom = &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Key:                  projection.Key,
			},
		}
	} else if isConfigMap(resource) {
		env.ValueFrom = &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Key:                  projection.Key,
			},
		}
	} else {
		return goerrors.New("illegal operation: projectEnvFrom called with resource other than ConfigMap or Secret")
	}
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else if util.GetEnv(nuxeoContainer, env.Name) != nil {
		return goerrors.New("invalid backing service projection - attempt to add duplicate environment var: " + env.Name)
	} else {
		nuxeoContainer.Env = append(nuxeoContainer.Env, env)
	}
	return nil
}

// READY TO TEST!
// secret/cm  mount       add volume and vol mnt for upstream secret or cm key
// other      mount       create/update secondary secret, add value as key, add
// If the passed resource is a Secret or ConfigMap, then a volume is created with a volume source of that
// resource. If the passed resource is any other kind of resource, then the resource value is copied to the secondary
// secret and a volume is created with the volume source of the secondary secret.
//
// Since all mounts are from secrets or CMs all mounts will use items:
//   items:
//   - key: SOME_KEY
//     path: as-some-file.txt
//
// And all volume mounts will be relative to /etc/nuxeo-operator/binding/<backing service name>
//
func projectMount(r *ReconcileNuxeo, namespace string, backingServiceName string, resource v1alpha1.BackingServiceResource, idx int,
	dep *appsv1.Deployment, secondarySecret *corev1.Secret) error {
	var nuxeoContainer *corev1.Container
	var err error
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}
	_ = nuxeoContainer
	var vol corev1.Volume
	if isSecret(resource) {
		// mount Secret
		vol = corev1.Volume{
			Name: resource.Kind + "-" + resource.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: resource.Name,
					Items: []corev1.KeyToPath{{
						Key:  resource.Projections[idx].Key,
						Path: resource.Projections[idx].Mount,
					}},
				},
			},
		}
	} else if isConfigMap(resource) {
		// mount ConfigMap
		vol = corev1.Volume{
			Name: resource.Kind + "-" + resource.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
					Items: []corev1.KeyToPath{{
						Key:  resource.Projections[idx].Key,
						Path: resource.Projections[idx].Mount,
					}},
				},
			},
		}
	} else {
		var val, newKey string
		val, err = getValueFromResource(r, resource, namespace, resource.Projections[idx].Path)
		if err != nil {
			return err
		}
		newKey, err = pathToKey(resource.Projections[idx].Path)
		if err != nil {
			return err
		}
		if _, ok := secondarySecret.Data[newKey]; !ok {
			return goerrors.New("secondary secret " + secondarySecret.Name + " already contains key " + newKey)
		}
		secondarySecret.Data[newKey] = []byte(val)

		vol = corev1.Volume{
			Name: resource.Kind + "-" + secondarySecret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secondarySecret.Name,
					Items: []corev1.KeyToPath{{
						Key:  newKey,
						Path: resource.Projections[idx].Mount,
					}},
				},
			},
		}
	}
	if err = addVolumeAndItems(dep, vol); err != nil {
		return err
	}
	volMnt := corev1.VolumeMount{
		Name:      vol.Name,
		ReadOnly:  true,
		MountPath: "/etc/nuxeo-operator/binding/" + backingServiceName,
	}
	return addVolMnt(nuxeoContainer, volMnt)
}

// secret/cm  transform   create/update secondary secret, add transformed value as key, add
// other      transform   " along with transformation
func projectTransform(backingServiceName string, resource v1alpha1.BackingServiceResource, idx int,
	dep *appsv1.Deployment, secondarySecret *corev1.Secret) error {
	return nil
}

// Converts a JSONPath expression to a valid Secret Key name
func pathToKey(jsonPath string) (string, error) {
	reg, err := regexp.Compile("[^-._a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(jsonPath, ""), nil
}

// Validates projections for the passed resource based on resource GVK. These are the currentlyt supported
// projections:
//
// Secret and ConfigMap resources 1) must specify .key, 2) must not specify .path, and 3) must specify one of: .mount,
// .env, or .transform. This means that only the projection key can be used to get a value from the secret/cm. It
// means that the resulting value can be projected as an environment variable, a mount, or transformed into
// a (secondary) secret.
//
// All other resources 1) must specify .path, 2) must not specify .key or .env, and 3) must specify one of .mount
// or .transform. This means that non-Secret and non-ConfigMap resources require JSONPath expressions to get the
// value. And it means that the resulting value - which will ALWAYS be in a secondary secret - can only be projected
// as a mount, or transformed.
func projectionsAreValid(gvk string, projections []v1alpha1.ResourceProjection) bool {
	for _, projection := range projections {
		if gvk == ".v1.secret" || gvk == ".v1.configmap" {
			if projection.Key == "" || projection.Path != "" ||
				(projection.Env == "" && projection.Mount == "" && projection.Transform == (v1alpha1.CertTransform{})) {
				return false
			}
		} else if projection.Key != "" || projection.Path == "" || projection.Env != "" ||
			(projection.Mount == "" && projection.Transform == (v1alpha1.CertTransform{})) {
			return false
		}
	}
	return true
}

// returns true of the passed resourced is a Secret, else false
func isSecret(resource v1alpha1.BackingServiceResource) bool {
	return strings.ToLower(resource.Group+"."+resource.Version+"."+resource.Kind) == ".v1.secret"
}

// returns true of the passed resourced is a ConfigMap, else false
func isConfigMap(resource v1alpha1.BackingServiceResource) bool {
	return strings.ToLower(resource.Group+"."+resource.Version+"."+resource.Kind) == ".v1.configmap"
}

// Reconciles the passed secondary secret with the cluster
func reconcileSecondary(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, secondarySecret *corev1.Secret,
	reqLogger logr.Logger) error {
	if len(secondarySecret.Data)+len(secondarySecret.StringData) != 0 {
		// secondary secret has content so it should exist in the cluster
		return addOrUpdate(r, secondarySecret.Name, instance.Namespace, secondarySecret, &corev1.Secret{},
			util.SecretCompare, reqLogger)
	} else {
		// secondary secret has no content so it not should exist in the cluster
		return removeIfPresent(r, instance, secondarySecret.Name, instance.Namespace, secondarySecret, reqLogger)
	}
}

// The nuxeo.conf CM contains three keys: inline, bindings, and nuxeo.conf. The inline nuxeo conf code populates
// the inline key. This code populates the binding key.
// TODO-ME WHO POPULATES THE NUXEO.CONF KEY!!!???
func reconcileNuxeoConfCM(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, bindings string, nodeSetName string,
	reqLogger logr.Logger) error {
	var nuxeoConfCM corev1.ConfigMap
	cmName := nuxeoConfCMName(instance, nodeSetName)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: instance.Namespace}, &nuxeoConfCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new nuxeo.conf ConfigMap " + cmName)
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: instance.Namespace,
			},
			Data: map[string]string{"bindings": bindings},
		}
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			reqLogger.Error(err, "Failed to create nuxeo.conf ConfigMap "+cmName)
			return err
		}
		return nil
	} else if err != nil {
		reqLogger.Error(err, "Error attempting to get nuxeo.conf ConfigMap "+cmName)
		return err
	}

	// cm exists  have binding content  action
	// ---------  --------------------  ----------
	//    NO              NO            should delete
	//    NO              YES           should create
	//    YES             NO            remove binding and update
	//    YES             YES           add binding and update
	return nil
}

// getJsonPathValue applies the passed JSONPath expression to the passed runtime object and returns the
// result of the parse. It's less friendly than the kubectl get -o=jsonpath= in that the passed JSON path
// has to be included in curly braces. A variety of errors are returned but an empty return value and nil
// error can also indicate that the provided JSON path didn't find anything in the passed resource.
func getJsonPathValue(obj runtime.Object, jsonPath string) (string, error) {
	if len(jsonPath) < 3 {
		return "", goerrors.New("invalid JSONPath expression: " + jsonPath)
	}
	if jsonPath[0:1]+jsonPath[len(jsonPath)-1:] != "{}" {
		return "", goerrors.New("JSONPath expression must be curly-brace enclosed: " + jsonPath)
	}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return "", err
	}
	j := jsonpath.New("jp")
	// parse the JSON path expression
	err = j.Parse(jsonPath)
	if err != nil {
		return "", err
	}
	result, err := j.FindResults(&unstructured)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	for ix := range result {
		if err := j.PrintResults(&buf, result[ix]); err != nil {
			return "", err
		}
	}
	return string(buf.Bytes()), nil
}

// gets the GVK+Name from the passed backing service resource, obtains the resource from the clustrer, and
// gets a value from the resource using the passed JSONPath expression. Any issue results in non-nil return
// code. As with getJsonPathValue, an empty return value and nil error can also indicate that the provided
// JSON path didn't find anything in the passed resource.
func getValueFromResource(r *ReconcileNuxeo, resource v1alpha1.BackingServiceResource, namespace string,
	jsonPath string) (string, error) {
	schemaGvk := schema.GroupVersionKind{
		Group:   resource.Group,
		Version: resource.Version,
		Kind:    resource.Kind,
	}
	obj, err := r.scheme.New(schemaGvk)
	if err != nil {
		return "", err
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: resource.Name, Namespace: namespace}, obj)
	if err != nil {
		return "", err
	}
	return getJsonPathValue(obj, jsonPath)
}