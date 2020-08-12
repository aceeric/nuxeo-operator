package nuxeo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/controller/nuxeo/preconfigs"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// all mount protections are under this directory in the Nuxeo container
	backingMountBase = "/etc/nuxeo-operator/binding/"
)

// Configures all backing services, resulting in volumes, mounts, secondary secrets, environment variables,
// and nuxeo.conf entries as needed to configure Nuxeo to connect to a backing service. Caller must handle
// the nuxeo.conf returned from the function by storing it the Operator-owned nuxeo.conf ConfigMap. The returned
// nuxeo.conf is a concatenation of all backing service nuxeo.conf entries (so it could be an empty string.)
func configureBackingServices(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, dep *appsv1.Deployment) (string, error) {
	nuxeoConf := ""
	for idx, backingService := range instance.Spec.BackingServices {
		var err error
		if !backingSvcIsValid(backingService) {
			return "", errors.New("invalid backing service definition at ordinal position " + strconv.Itoa(idx))
		}
		// if configurer provided a preconfigured backing service use that as if it were actually in the CR
		if backingService.Preconfigured.Type != "" {
			if backingService, err = xlatBacking(backingService.Preconfigured); err != nil {
				return "", err
			}
		}
		if err = configureBackingService(r, instance, backingService, dep); err != nil {
			return "", err
		}
		// accumulate each backing service's nuxeo.conf settings
		nuxeoConf = joinCompact("\n", nuxeoConf, backingService.NuxeoConf)

		if backingService.Template != "" {
			// backing service requires a template
			if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
				return "", err
			} else {
				env := corev1.EnvVar{
					Name:  "NUXEO_TEMPLATES",
					Value: backingService.Template,
				}
				if err := util.MergeOrAddEnvVar(nuxeoContainer, env, ","); err != nil {
					return "", err
				}
			}
		}
	}
	return nuxeoConf, nil
}

// Configures one backing service. Iterates all resources and bindings, calls helpers to add environment variables
// and mounts into the nuxeo container, and volumes in the passed deployment. May create a secondary secret if
// needed.
func configureBackingService(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, backingService v1alpha1.BackingService,
	dep *appsv1.Deployment) error {
	// 0-1 secondary secret per backing service
	secondarySecret := defaultSecondarySecret(r, instance, backingService)
	for _, resource := range backingService.Resources {
		gvk := strings.ToLower(resource.Group + "." + resource.Version + "." + resource.Kind)
		// validating the projections here ensures that the switch statement below works
		if err := validateProjections(gvk, resource.Projections); err != nil {
			return err
		}
		for i := 0; i < len(resource.Projections); i++ {
			var err error
			projection := resource.Projections[i]
			switch {
			case projection.Env != "" && projection.Value:
				err = projectEnvVal(r, instance.Namespace, resource, projection, dep)
			case (isSecret(resource) || isConfigMap(resource)) && projection.Env != "":
				err = projectEnvFrom(resource, projection, dep)
			case projection.Mount != "":
				err = projectMount(r, instance.Namespace, backingService.Name, resource, projection, dep,
					&secondarySecret)
			case projection.Transform != (v1alpha1.CertTransform{}):
				err = projectTransform(r, instance.Namespace, backingService.Name, resource, projection, dep,
					&secondarySecret)
			default:
				err = errors.New(fmt.Sprintf("no handler for projection at ordinal position %v in resource %s",
					i, resource.Name))
			}
			if err != nil {
				return err
			}
		}
	}
	return reconcileSecondary(r, instance, &secondarySecret)
}

func projectEnvVal(r *ReconcileNuxeo, namespace string, resource v1alpha1.BackingServiceResource,
	projection v1alpha1.ResourceProjection, dep *appsv1.Deployment) error {
	env := corev1.EnvVar{
		Name: projection.Env,
	}
	val, _, err := getValueFromResource(r, resource, namespace, projection.From)
	if err != nil {
		return err
	} else if val == nil {
		return fmt.Errorf("resource %v does not have value for path %v", resource.Name, projection.From)
	}
	env.Value = string(val)
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
}

// Adds an environment variable with a valueFrom that references the key in the passed resource, which must be a
// Secret or ConfigMap. Returns non-nil error if: passed resource is not a Secret or ConfigMap, or environment
// variable name is not unique in the nuxeo container. Otherwise nil error is returned and an environment variable
// is added to the nuxeo container in the passed deployment like:
//   env:
//   - name: ELASTIC_PASSWORD              # from projection.Env
//     valueFrom:
//       secretKeyRef:
//         key: elastic                    # from projection.From
//         name: elastic-es-elastic-user   # from resource.Name
func projectEnvFrom(resource v1alpha1.BackingServiceResource, projection v1alpha1.ResourceProjection,
	dep *appsv1.Deployment) error {
	env := corev1.EnvVar{
		Name: projection.Env,
	}
	if isSecret(resource) {
		env.ValueFrom = &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Key:                  projection.From,
			},
		}
	} else if isConfigMap(resource) {
		env.ValueFrom = &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Key:                  projection.From,
			},
		}
	} else {
		return errors.New("illegal operation: projectEnvFrom called with resource other than ConfigMap or Secret")
	}
	if nuxeoContainer, err := util.GetNuxeoContainer(dep); err != nil {
		return err
	} else {
		return util.OnlyAddEnvVar(nuxeoContainer, env)
	}
}

// Handles mount projections for resources by creating/appending to a volume with a projection source like so:
//   volumes:
//   - name: backing-elastic
//     projected:
//       sources:
//       - secret:
//           name: tls-secret
//           items:
//           - from: ca.crt
//             path: ca.crt
// There will be one such volume and corresponding vol mount for each backing service specifying any mount
// projection in the Nuxeo CR like so:
//   backingServices:
//   - name: elastic # Nuxeo Operator creates volume "backing-elastic"
//     resources:
//     - version: v1
//       kind: secret
//       name: some-secret
//       projections:
//       - from: ca.crt
//         mount: ca.crt # becomes path in projection
// This function supports projecting certificates and similar values onto the filesystem so nuxeo.conf can reference
// them with explicit filesystem paths. The function also supports projecting values from non-ConfigMap and
// non-Secret cluster resources. In this case, values are copied from the upstream resources into the passed
// secondary secret and mounted from the secondary secret.
func projectMount(r *ReconcileNuxeo, namespace string, backingServiceName string,
	resource v1alpha1.BackingServiceResource, projection v1alpha1.ResourceProjection, dep *appsv1.Deployment,
	secondarySecret *corev1.Secret) error {

	var nuxeoContainer *corev1.Container
	var err error
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}
	vol := corev1.Volume{
		Name: strings.ToLower("backing-" + backingServiceName),
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				DefaultMode: util.Int32Ptr(420),
				Sources:     []corev1.VolumeProjection{},
			},
		},
	}
	var src corev1.VolumeProjection
	if isSecret(resource) {
		src = corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Items: []corev1.KeyToPath{{
					Key:  projection.From,
					Path: projection.Mount,
				}},
			},
		}
	} else if isConfigMap(resource) {
		src = corev1.VolumeProjection{
			ConfigMap: &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{Name: resource.Name},
				Items: []corev1.KeyToPath{{
					Key:  projection.From,
					Path: projection.Mount,
				}},
			},
		}
	} else {
		// configurer wants value from non-Secret/non-CM which isn't a supported Kubernetes projection type. So
		// copy the value into the secondary secret and use the secondary secret as the source of the mount. Caller
		// must reconcile secondary secret
		var val []byte
		var newKey string
		if val, _, err = getValueFromResource(r, resource, namespace, projection.From); err != nil || val == nil {
			return err
		}
		if newKey, err = pathToKey(projection.From); err != nil {
			return err
		}
		if _, ok := secondarySecret.Data[newKey]; ok {
			return errors.New("secondary secret " + secondarySecret.Name + " already contains key " + newKey)
		}
		// caller must reconcile
		secondarySecret.Data[newKey] = val
		src = corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{Name: secondarySecret.Name},
				Items: []corev1.KeyToPath{{
					Key:  newKey,
					Path: projection.Mount,
				}},
			},
		}
	}
	vol.VolumeSource.Projected.Sources = append(vol.VolumeSource.Projected.Sources, src)
	if err = addVolumeProjectionAndItems(dep, vol); err != nil {
		return err
	}
	volMnt := corev1.VolumeMount{
		Name:      vol.Name,
		ReadOnly:  true,
		MountPath: backingMountBase + backingServiceName,
	}
	return addVolMnt(nuxeoContainer, volMnt)
}

// Takes an upstream source cert and optionally key, and generates a trust store or key store which it stores
// in the passed secondary secret. The store is mounted as a file and can be referenced from nuxeo.conf to
// configure TLS backing service connections. Note that the keystore conversion generates a random password
// as part of the conversion. Therefore, this function includes code to make sure that this only happens once
// per upstream resource version. In other words, if an upstream resource contains a cert, and that resource
// has metadata.resourceVersion X, then when this function runs, if the upstream resource still has resourceVersion X,
// then don't re-generate the keystore - just use the keystore (and password) that were already created. This
// is accomplished by annotating the secondary secret - only for cases where this keystore conversion/password
// gen is performed.
func projectTransform(r *ReconcileNuxeo, namespace string, backingServiceName string,
	resource v1alpha1.BackingServiceResource, projection v1alpha1.ResourceProjection, dep *appsv1.Deployment,
	secondarySecret *corev1.Secret) error {
	var resVer string
	var err error
	var cert []byte
	var nuxeoContainer *corev1.Container

	storeKey := projection.Transform.Store
	passKey := projection.Transform.Password
	// some basic validation to protect against logic errors in the operator
	if _, ok := secondarySecret.Data[storeKey]; ok {
		return errors.New("key " + storeKey + " already defined in secret " + secondarySecret.Name)
	} else if _, ok := secondarySecret.Data[passKey]; ok {
		return errors.New("key " + passKey + " already defined in secret " + secondarySecret.Name)
	}

	// get the cert and possibly key from the cluster
	var privateKey []byte
	if cert, resVer, err = getValueFromResource(r, resource, namespace, projection.Transform.Cert); err != nil {
		return err
	}
	if projection.Transform.Type == v1alpha1.KeyStore {
		if privateKey, _, err = getValueFromResource(r, resource, namespace, projection.Transform.PrivateKey);
			err != nil {
			return err
		}
	}
	if secondarySecretIsCurrent(r, secondarySecret.Name, namespace, resource, resVer) {
		if err = loadSecondary(r, resource, namespace, secondarySecret, resVer, storeKey, passKey); err != nil {
			return err
		}
	} else if err = populateSecondary(resource, secondarySecret, storeKey, passKey, resVer, cert, privateKey,
		projection.Transform.Type); err != nil {
		return err
	}
	// generate deployment/pod structs to support the projection
	vol := corev1.Volume{
		Name: strings.ToLower("backing-" + backingServiceName),
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				DefaultMode: util.Int32Ptr(420),
				Sources: []corev1.VolumeProjection{{
					Secret: &corev1.SecretProjection{
						LocalObjectReference: corev1.LocalObjectReference{Name: secondarySecret.Name},
						Items: []corev1.KeyToPath{{
							Key:  storeKey,
							Path: storeKey, // for now the path is the key maybe in future provide an override
						}, {
							Key:  passKey,
							Path: passKey,
						}},
					},
				}},
			},
		},
	}
	if err = addVolumeProjectionAndItems(dep, vol); err != nil {
		return err
	}
	volMnt := corev1.VolumeMount{
		Name:      vol.Name,
		ReadOnly:  true,
		MountPath: backingMountBase + backingServiceName,
	}
	if nuxeoContainer, err = util.GetNuxeoContainer(dep); err != nil {
		return err
	}
	if err = addVolMnt(nuxeoContainer, volMnt); err != nil {
		return err
	}
	// store password
	env := corev1.EnvVar{
		Name: projection.Transform.PassEnv,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secondarySecret.Name},
				Key:                  passKey,
			},
		},
	}
	return util.OnlyAddEnvVar(nuxeoContainer, env)
}

// Gets the existing secondary secret from the cluster and populates the passed keys in the the passed in-mem
// secret struct so a subsequent reconcile of the secret has nothing to do. Also annotates the secret as would be
// done on initial creation.
func loadSecondary(r *ReconcileNuxeo, resource v1alpha1.BackingServiceResource, namespace string,
	secondarySecret *corev1.Secret, resVer string, keys ...string) error {
	obj := corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: secondarySecret.Name, Namespace: namespace},
		&obj); err != nil {
		// should not get here if the secondary secret does not exist
		return err
	}
	secondarySecret.Annotations[genAnnotationKey(resource)] = resVer
	for _, key := range keys {
		secondarySecret.Data[key] = obj.Data[key]
	}
	return nil
}

// Converts the passed cert and optionally key into a trust store or key store of type JKS. If no error,
// then the secondary secret struct is updated with the store and the password
//
// args
//  resource        - the upstream resource GVK+Name that contributed cert and key - used to annotate the
//                    secondary secret
//  secondarySecret - ref to the secondary secret that will hold the store and store pass
//                    generated by this fxn
//  storeKey        - the key in secondarySecret to hold the trust store
//  passKey         - the key in secondarySecret to hold the trust store password generated by this function
//  resVer          - the resourceVersion of the upstream resource also used to annotate the secondary secret
//  cert            - the PEM-encoded certificate from the upstream resource to add to the trust store
//  privateKey      - the PEM-encoded private key "
//  transformType   - indicates trust store or key store
func populateSecondary(resource v1alpha1.BackingServiceResource, secondarySecret *corev1.Secret, storeKey string,
	passKey string, resVer string, cert []byte, privateKey []byte, transformType v1alpha1.CertTransformType) error {

	secondarySecret.Annotations[genAnnotationKey(resource)] = resVer
	var store []byte
	var pass string
	var err error
	if transformType == v1alpha1.TrustStore {
		if store, pass, err = trustStoreFromPEM(cert); err != nil {
			return err
		}
	} else if store, pass, err = keyStoreFromPEM(cert, privateKey); err != nil {
		return err
	}
	secondarySecret.Data[storeKey] = store
	secondarySecret.Data[passKey] = []byte(pass)
	return nil
}

// Converts a JSONPath expression to a valid Secret Key name by removing invalid characters
func pathToKey(jsonPath string) (string, error) {
	reg, err := regexp.Compile("[^-._a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(jsonPath, ""), nil
}

// Validates projections for the passed backing service resource based on resource GVK
func validateProjections(gvk string, projections []v1alpha1.ResourceProjection) error {
	for idx, projection := range projections {
		if projection.Env != "" && projection.Value {
			return nil
		} else if projection.Env != "" && gvk != ".v1.secret" && gvk != ".v1.configmap" {
			return fmt.Errorf("environment projection requires Secret/ConfigMap in projection %v", idx)
		} else if projection.Transform != (v1alpha1.CertTransform{}) {
			if projection.From != "" || projection.Mount != "" || projection.Env != "" {
				return fmt.Errorf("transform cannot specify from, env, or mount in projection %v", idx)
			}
		}
	}
	return nil
}

// returns true of the passed resourced is a Secret, else false
func isSecret(resource v1alpha1.BackingServiceResource) bool {
	return strings.ToLower(resource.Group+"."+resource.Version+"."+resource.Kind) == ".v1.secret"
}

// returns true of the passed resourced is a ConfigMap, else false
func isConfigMap(resource v1alpha1.BackingServiceResource) bool {
	return strings.ToLower(resource.Group+"."+resource.Version+"."+resource.Kind) == ".v1.configmap"
}

// Reconciles the passed secondary secret with the cluster. A secondary secret is one that is created for
// a backing service whenever a) a value is obtained from a backing service resource other than a Secret or
// ConfigMap, or b) a backing service value is transformed. In both cases, cluster storage is needed for the
// value and so the Operator creates a "secondary secret" to hold such values. There is 0-1 secondary
// secret per backing service.
func reconcileSecondary(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, secondarySecret *corev1.Secret) error {
	if len(secondarySecret.Data)+len(secondarySecret.StringData) != 0 {
		// secondary secret has content so it should exist in the cluster
		_, err := addOrUpdate(r, secondarySecret.Name, instance.Namespace, secondarySecret, &corev1.Secret{},
			util.SecretComparer)
		return err
	} else {
		// secondary secret has no content so it not should exist in the cluster
		return removeIfPresent(r, instance, secondarySecret.Name, instance.Namespace, secondarySecret)
	}
}

// Gets the GVK+Name from the passed backing service resource struct, obtains the corresponding resource from
// the cluster using that GVK+Name, and gets a value from the cluster resource using the passed "from" arg,
// which is a Key if the passed resource is a Secret or ConfigMap otherwise a JSONPath expression.
//
// Any issue results in non-nil return code. As with GetJsonPathValue, an empty return value and nil error can
// indicate that the provided JSON path didn't find anything in the passed resource. If the requested
// resource does not exist in the cluster, a nil error is returned, and a nil value is returned.
//
// Returns [resource value] [resource version] [error]
func getValueFromResource(r *ReconcileNuxeo, resource v1alpha1.BackingServiceResource, namespace string,
	from string) ([]byte, string, error) {
	gvk := strings.ToLower(resource.Group + "." + resource.Version + "." + resource.Kind)
	if gvk == ".v1.secret" {
		if from == "" {
			return nil, "", errors.New("no key provided in projection")
		}
		obj := corev1.Secret{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: resource.Name, Namespace: namespace}, &obj);
			err != nil {
			if apierrors.IsNotFound(err) {
				return nil, "", nil
			}
			return nil, "", err
		}
		return obj.Data[from], obj.ResourceVersion, nil
	} else if gvk == ".v1.configmap" {
		if from == "" {
			return nil, "", errors.New("no key provided in projection")
		}
		obj := corev1.ConfigMap{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: resource.Name, Namespace: namespace}, &obj);
			err != nil {
			if apierrors.IsNotFound(err) {
				return nil, "", nil
			}
			return nil, "", err
		}
		return []byte(obj.Data[from]), obj.ResourceVersion, nil
	} else {
		return getValueByPath(r, resource, namespace, from)
	}
}

// Gets a value from the passed cluster resource using a jsonPath expression
// args
//  r         - operator reconcile struct
//  resource  - a resource definition that provides GVK+Name of a cluster resource
//  namespace - the current namespace
//  from      - a json path expression into 'resource' expressing a value to get
//
// returns
//  value   - the value from the resource at the passed json path
//  version - the resource version from the resource
//  err     - non-nil if error, otherwise version *is* populated and value *may be* populated if
//            the json path yielded a value
func getValueByPath(r *ReconcileNuxeo, resource v1alpha1.BackingServiceResource, namespace string,
	from string) ([]byte, string, error) {
	if from == "" {
		return nil, "", errors.New("no path provided in projection")
	}
	u := unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   resource.Group,
		Version: resource.Version,
		Kind:    resource.Kind,
	}
	u.SetGroupVersionKind(gvk)
	if err := r.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: resource.Name},
		&u); err != nil {
		return nil, "", fmt.Errorf("unable to get resource: %v", resource.Name)
	}
	resVer := ""
	if rv, rve := util.GetJsonPathValueU(u.Object, "{.metadata.resourceVersion}"); rve == nil && rv != nil {
		resVer = string(rv)
	} else if rve != nil {
		return nil, "", rve
	} else if rv == nil {
		return nil, "", fmt.Errorf("unable to get resource version from resource: %v", resource.Name)
	}
	resVal, resErr := util.GetJsonPathValueU(u.Object, from)
	return resVal, resVer, resErr
}

// creates and returns a secondary secret struct in the format required by the Operator
func defaultSecondarySecret(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo,
	backingService v1alpha1.BackingService) corev1.Secret {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name + "-secondary-" + backingService.Name,
			Namespace:   instance.Namespace,
			Annotations: map[string]string{},
		},
		Data: map[string][]byte{},
		Type: corev1.SecretTypeOpaque,
	}
	_ = controllerutil.SetControllerReference(instance, &secret, r.scheme)
	return secret
}

// Generates an annotation key for a secondary secret like: nuxeo.operator.G.V.K.N (or nuxeo.operator.V.K.N if
// group is ""). E.g.: nuxeo.operator.v1.secret.elastic-es-http-certs-public. This annotation is used by the
// operator to know if an upstream resource that is transformed into a secondary secret has changed.
func genAnnotationKey(resource v1alpha1.BackingServiceResource) string {
	key := strings.Replace(fmt.Sprintf("nuxeo.operator.%s.%s.%s.%s",
		strings.ToLower(resource.Group),
		strings.ToLower(resource.Version),
		strings.ToLower(resource.Kind),
		strings.ToLower(resource.Name)), "..", ".", 1)
	return key
}

// If secondary secret exists in-cluster, and has secondary secret annotation, and annotation resource version
// is the same, then returns true, else false. If all these things are true then the secondary secret is current
// with the upstream resource it got its value(s) from.
func secondarySecretIsCurrent(r *ReconcileNuxeo, secondarySecret string, namespace string,
	resource v1alpha1.BackingServiceResource, resourceVersion string) bool {
	obj := corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: secondarySecret, Namespace: namespace}, &obj);
		err != nil {
		return false
	} else {
		expectedAnnotation := genAnnotationKey(resource)
		if existingResVer, ok := obj.Annotations[expectedAnnotation]; ok {
			return existingResVer == resourceVersion
		}
	}
	return false
}

// A valid backing service specifies a preConfigured entry, in which case everything else is ignored, or, it
// specifies a name, and a resource list. A nuxeo.conf is optional
func backingSvcIsValid(backing v1alpha1.BackingService) bool {
	if !reflect.DeepEqual(backing.Preconfigured, v1alpha1.PreconfiguredBackingService{}) {
		return true
	} else {
		return backing.Name != "" && !reflect.DeepEqual(backing.Resources, v1alpha1.BackingServiceResource{})
	}
}

// Uses the passed preconfigured backing service to generate a backing service struct that will wire Nuxeo
// up to a backing service using well-known resources provided by the backing service operator.
func xlatBacking(preconfigured v1alpha1.PreconfiguredBackingService) (v1alpha1.BackingService, error) {
	switch preconfigured.Type {
	case v1alpha1.ECK:
		return preconfigs.EckBacking(preconfigured, backingMountBase)
	case v1alpha1.Strimzi:
		return preconfigs.StrimziBacking(preconfigured, backingMountBase)
	case v1alpha1.Crunchy:
		return preconfigs.CrunchyBacking(preconfigured, backingMountBase)
	default:
		// can only happen if someone adds a preconfig and forgets to add a case statement for it
		return v1alpha1.BackingService{}, fmt.Errorf("unknown pre-config: %v", preconfigured.Type)
	}
}
