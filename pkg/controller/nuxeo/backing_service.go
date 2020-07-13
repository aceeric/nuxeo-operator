package nuxeo

import (
	goerrors "errors"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
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
func configureBackingServices(r *ReconcileNuxeo, instance *v1alpha1.Nuxeo, dep *appsv1.Deployment,
	reqLogger logr.Logger) error {
	nuxeoConf := ""
	for _, backingService := range instance.Spec.BackingServices {
		if err := configureBackingService(r, backingService, dep, reqLogger); err != nil {
			return err
		}
		nuxeoConf = strings.TrimSpace(strings.Join([]string{nuxeoConf, backingService.NuxeoConf}, "\r"))
	}
	// update configmap in cluster
	return nil
}

func configureBackingService(r *ReconcileNuxeo, backingService v1alpha1.BackingService, dep *appsv1.Deployment,
	reqLogger logr.Logger) error {
	for _, resource := range backingService.Resources {
		kind := strings.ToLower(resource.Group + "." + resource.Version + "." + resource.Kind)
		if !projectionsAreValid(kind, resource.Projections) {
			return goerrors.New("resource " + resource.Name + " has invalid projections configured")
		}
	}
	// possibilities
	// source     projection action
	// ---------- ---------- ---------------------------------------------------
	// secret/cm  env        add env valueFrom upstream secret or cm key
	// secret/cm  mount      add volume and volmnt for upstream secret or cm key
	// secret/cm  transform  determined by transform
	// other      mount      create secondary secret, add value as key, add
	//                       volume and vol mnt for 2ndary secret key
	// other      transform  determined by transform



	return nil
}

func projectEnvFrom() {

}

// Validates projections for the passed resource GVK. Projections for secret and config map resources must specify .key,
// cannot specify .path, and must specify one of .mount, .env, or .transform. All other resources must specify .path,
// and one of .mount. and .transform
func projectionsAreValid(kind string, projections []v1alpha1.ResourceProjection) bool {
	for _, projection := range projections {
		if kind == ".v1.secret" || kind == ".v1.configmap" {
			if projection.Key == "" || projection.Path != "" ||
				(projection.Env == "" && projection.Mount == "" && projection.Transform == (v1alpha1.CertTransform{})) {
				return false
			} else if projection.Key != "" || projection.Path == "" || projection.Env != "" ||
				(projection.Mount == "" && projection.Transform == (v1alpha1.CertTransform{})) {
				return false
			}
		}
	}
	return true
}