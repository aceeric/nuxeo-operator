package nuxeo

import (
	"context"
	goerrors "errors"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	nuxeov1alpha1 "nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
	"nuxeo-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nuxeo")

// Add creates a new Nuxeo Controller and adds it to the passed Manager. The Manager will set fields
// on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNuxeo{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add creates a new Controller registered with the passed Manager. The passed Reconciler 'r' is
// provided to the controller as the controller's reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nuxeo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Nuxeo
	err = c.Watch(&source.Kind{Type: &nuxeov1alpha1.Nuxeo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &nuxeov1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &nuxeov1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// todo-me Watch for changes to ServiceAccount?

	// Watch for changes to ConfigMap
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &nuxeov1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to OpenShift Route
	if util.IsOpenShift() {
		err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &nuxeov1alpha1.Nuxeo{},
		})
		if err != nil {
			return err
		}
	} else {
		// Watch for changes to Kubernetes Ingress
		err = c.Watch(&source.Kind{Type: &v1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &nuxeov1alpha1.Nuxeo{},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// blank assignment to verify that ReconcileNuxeo implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNuxeo{}

// ReconcileNuxeo reconciles a Nuxeo object
type ReconcileNuxeo struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile is the main reconciler. It reads the state of the cluster for a Nuxeo object and may alter
// cluster state based on the state of the Nuxeo object Spec. Note: The Controller will requeue the Request
// to be processed again if the returned error is non-nil or Result.Requeue is true, otherwise upon
// completion it will remove the work from the queue.
//
// todo-me reconcile the nodeset last since the deployment depends on many other assets
func (r *ReconcileNuxeo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Nuxeo")

	// Get the Nuxeo CR from the request namespace
	instance := &nuxeov1alpha1.Nuxeo{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Nuxeo CR not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Nuxeo resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Nuxeo")
		return reconcile.Result{}, err
	}
	// The Nuxeo CR spec contains a list of NodeSets. Each NodeSet defines a Deployment. For example, a
	// Nuxeo CR might define one NodeSet for a set of interactive Nuxeo pods, and a second NodeSet for a
	// set of worker Nuxeo pods. This results in two Deployments: one controlling interactive Pods,
	// and one controlling worker Pods. The interactive pods will be made accessible by the Operator via a
	// Route/Ingress. The non-interactive pods will not be accessible outside of the cluster.
	var interactiveNodeSet *nuxeov1alpha1.NodeSet = nil
	for _, nodeSet := range instance.Spec.NodeSets {
		if result, err := reconcileNodeSet(r, nodeSet, instance, instance.Spec.RevProxy, reqLogger); err != nil {
			return result, err
		} else if result == (reconcile.Result{Requeue: true}) {
			return result, nil
		}
		if nodeSet.Interactive {
			if interactiveNodeSet != nil {
				err = goerrors.New("nuxeo validation error")
				reqLogger.Error(err, "Only one interactive NodeSet is allowed")
				return reconcile.Result{}, err
			}
			interactiveNodeSet = &nodeSet
		}
		if _, err = reconcileNuxeoConf(r, instance, nodeSet, reqLogger); err != nil {
			return reconcile.Result{}, err
		}
	}
	// ensure that exactly one interactive nodeset was specified in the CR
	if interactiveNodeSet == nil {
		err = goerrors.New("nuxeo validation error")
		reqLogger.Error(err, "No interactive NodeSets specified")
		return reconcile.Result{}, err
	}
	if _, err = reconcileService(r, instance.Spec.Service, *interactiveNodeSet, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	forcePassthrough := false
	if interactiveNodeSet.NuxeoConfig.TlsSecret != "" {
		// if Nuxeo is terminating TLS then force tls passthrough termination in the route/ingress
		forcePassthrough = true
	}
	if _, err = reconcileAccess(r, instance.Spec.Access, forcePassthrough, *interactiveNodeSet, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcileServiceAccount(r, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcilePvc(r, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcileClid(r, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	updateNuxeoStatus(r, instance, reqLogger)
	if err = r.client.Status().Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update Nuxeo status", "Namespace", instance.Namespace, "Name", instance.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
