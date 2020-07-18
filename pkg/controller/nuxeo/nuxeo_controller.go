package nuxeo

import (
	"context"
	goerrors "errors"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
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

	if clusterHasRoute(mgr) {
		util.SetIsOpenShift(true)
		if err := registerOpenShiftRoute(); err != nil {
			return err
		}
	} else if !clusterHasIngress(mgr) {
		return goerrors.New("unable to determine cluster type")
	} else if err := registerKubernetesIngress(); err != nil {
		return err
	}

	// Watch for changes to primary resource Nuxeo
	err = c.Watch(&source.Kind{Type: &v1alpha1.Nuxeo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// todo-me Watch for changes to ServiceAccount?
	// todo-me Watch for changes to Owned Secrets?

	// Watch for changes to ConfigMap
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Nuxeo{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to OpenShift Route
	if util.IsOpenShift() {
		err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.Nuxeo{},
		})
		if err != nil {
			return err
		}
	} else {
		// Watch for changes to Kubernetes Ingress
		err = c.Watch(&source.Kind{Type: &v1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.Nuxeo{},
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
// todo-me investigate whether/when to return non-nil err. Stack trace resulting from non-nil error
//  seems of limited use
// todo-me review all (reconcile.Result, error) returns - most should only return error
func (r *ReconcileNuxeo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Nuxeo")

	// Get the Nuxeo CR from the request namespace
	instance := &v1alpha1.Nuxeo{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Nuxeo resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get Nuxeo")
		return reconcile.Result{Requeue: true}, err
	}
	// only configure service/ingress/route for the interactive NodeSet
	var interactiveNodeSet v1alpha1.NodeSet
	if interactiveNodeSet, err = getInteractiveNodeSet(instance.Spec.NodeSets, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcileService(r, instance.Spec.Service, interactiveNodeSet, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcileAccess(r, instance.Spec.Access, interactiveNodeSet, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcileServiceAccount(r, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if _, err = reconcilePvc(r, instance); err != nil {
		return reconcile.Result{}, err
	}
	if err = reconcileClid(r, instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	}
	if result, err := r.reconcileNodeSets(instance, reqLogger); err != nil {
		return reconcile.Result{}, err
	} else if result == (reconcile.Result{Requeue: true}) {
		return result, nil
	}
	return updateNuxeoStatus(r, instance, reqLogger)
}

// Reconcile each NodeSet to a Deployment
func (r *ReconcileNuxeo) reconcileNodeSets(instance *v1alpha1.Nuxeo, reqLogger logr.Logger) (reconcile.Result, error) {
	for _, nodeSet := range instance.Spec.NodeSets {
		if result, err := reconcileNodeSet(r, nodeSet, instance, instance.Spec.RevProxy, reqLogger); err != nil {
			return result, err
		} else if result == (reconcile.Result{Requeue: true}) {
			return result, nil
		}
	}
	return reconcile.Result{}, nil
}

// returns the interactive NodeSet from the passed array, or non-nil error if a) there is no interactive NodeSet
// defined, or b) there is more than one interactive NodeSet defined
func getInteractiveNodeSet(nodeSets []v1alpha1.NodeSet, reqLogger logr.Logger) (v1alpha1.NodeSet, error) {
	toReturn := v1alpha1.NodeSet{}
	for _, nodeSet := range nodeSets {
		if nodeSet.Interactive {
			if toReturn.Name != "" {
				err := goerrors.New("exactly one interactive NodeSet is required")
				reqLogger.Error(err, "Invalid Nuxeo CR")
				return toReturn, err
			}
			toReturn = nodeSet
		}
	}
	return toReturn, nil
}

// returns true if the cluster contains an OpenShift Route type
func clusterHasRoute(mgr manager.Manager) bool {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "route.openshift.io", Version: "v1", Kind: "Route"})
	err := mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: "foo"}, obj)
	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); ok {
			return false
		}
	}
	return true
}

// returns true if the cluster contains a Kubernetes Ingress type
func clusterHasIngress(mgr manager.Manager) bool {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1beta1", Kind: "Ingress"})
	err := mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: "foo"}, obj)
	if err != nil {
		if _, ok := err.(*meta.NoKindMatchError); ok {
			return false
		}
	}
	return true
}

// registerOpenShiftRoute registers OpenShift Route types with the Scheme Builder
func registerOpenShiftRoute() error {
	const GroupName = "route.openshift.io"
	const GroupVersion = "v1"
	SchemeGroupVersion := schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion,
			&routev1.Route{},
			&routev1.RouteList{},
		)
		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
		return nil
	}
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	return SchemeBuilder.AddToScheme(scheme.Scheme)
}

// registerKubernetesIngress registers Kubernetes Ingress types with the Scheme Builder. Note, according to:
//  https://kubernetes.io/blog/2019/07/18/api-deprecations-in-1-16/
// "Use the networking.k8s.io/v1beta1 API version, available since v1.14"
func registerKubernetesIngress() error {
	const GroupName = "networking.k8s.io"
	const GroupVersion = "v1beta1"
	SchemeGroupVersion := schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	addKnownTypes := func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion,
			&v1beta1.Ingress{},
			&v1beta1.IngressList{},
		)
		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
		return nil
	}
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	return SchemeBuilder.AddToScheme(scheme.Scheme)
}
