package nuxeo

//import (
//	"context"
//	"github.com/go-logr/logr"
//	appsv1 "k8s.io/api/apps/v1"
//
//	nuxeov1alpha1 "nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
//
//	corev1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/types"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"sigs.k8s.io/controller-runtime/pkg/controller"
//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
//	"sigs.k8s.io/controller-runtime/pkg/handler"
//	logf "sigs.k8s.io/controller-runtime/pkg/log"
//	"sigs.k8s.io/controller-runtime/pkg/manager"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//	"sigs.k8s.io/controller-runtime/pkg/source"
//)
//
//var log = logf.Log.WithName("controller_nuxeo")
//
///**
//* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
//* business logic.  Delete these comments after modifying this file.*
// */
//
//// Add creates a new Nuxeo Controller and adds it to the Manager. The Manager will set fields on the Controller
//// and Start it when the Manager is Started.
//func Add(mgr manager.Manager) error {
//	return add(mgr, newReconciler(mgr))
//}
//
//// newReconciler returns a new reconcile.Reconciler
//func newReconciler(mgr manager.Manager) reconcile.Reconciler {
//	return &ReconcileNuxeo{client: mgr.GetClient(), scheme: mgr.GetScheme()}
//}
//
//// add adds a new Controller to mgr with r as the reconcile.Reconciler
//func add(mgr manager.Manager, r reconcile.Reconciler) error {
//	// Create a new controller
//	c, err := controller.New("nuxeo-controller", mgr, controller.Options{Reconciler: r})
//	if err != nil {
//		return err
//	}
//
//	// Watch for changes to primary resource Nuxeo
//	err = c.Watch(&source.Kind{Type: &nuxeov1alpha1.Nuxeo{}}, &handler.EnqueueRequestForObject{})
//	if err != nil {
//		return err
//	}
//
//	// TODO THIS SHOULD WATCH WHAT??
//	// TODO(user): Modify this to be the types you create that are owned by the primary resource
//	// Watch for changes to secondary resource Pods and requeue the owner Nuxeo
//	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
//		IsController: true,
//		OwnerType:    &nuxeov1alpha1.Nuxeo{},
//	})
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// blank assignment to verify that ReconcileNuxeo implements reconcile.Reconciler
//var _ reconcile.Reconciler = &ReconcileNuxeo{}
//
//// ReconcileNuxeo reconciles a Nuxeo object
//type ReconcileNuxeo struct {
//	// This client, initialized using mgr.Client() above, is a split client
//	// that reads objects from the cache and writes to the apiserver
//	client client.Client
//	scheme *runtime.Scheme
//}
//
//// Reconcile reads that state of the cluster for a Nuxeo object and makes changes based on the state read
//// and what is in the Nuxeo.Spec
//// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
//// a Nuxeo Deployment for each Nuxeo CR
//// Note:
//// The Controller will requeue the Request to be processed again if the returned error is non-nil or
//// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
//func (r *ReconcileNuxeo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
//	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
//	reqLogger.Info("Reconciling Nuxeo")
//
//	// Get the Nuxeo CR from the request namespace
//	instance := &nuxeov1alpha1.Nuxeo{}
//	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
//	if err != nil {
//		if errors.IsNotFound(err) {
//			// Request object not found, could have been deleted after reconcile request.
//			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
//			// Return and don't requeue
//			reqLogger.Info("Nuxeo resource not found. Ignoring since object must be deleted")
//			return reconcile.Result{}, nil
//		}
//		// Error reading the object - requeue the request.
//		reqLogger.Error(err, "Failed to get Nuxeo")
//		return reconcile.Result{}, err
//	}
//
//	// The Nuxeo CR spec contains a list of NodeSets. Each NodeSet defines a Deployment. The typical
//	// use case is to define one NodeSet for a set of interactive Nuxeo nodes, and one NodeSet for a
//	// set of worker Nuxeo nodes. If this is done, the result is two deployments - one controlling
//	// the interactive Pods, and one for controlling the worker Pods.
//	for _, nodeSet := range instance.Spec.NodeSets {
//		result, err := reconcileNodeSet(r, nodeSet, instance, reqLogger)
//		if err != nil {
//			return result, err
//		}
//	}
//
//	//// TODO DO I NEED THIS?
//	//// Update the Nuxeo status with the pod names
//	//// List the pods for this nuxeo's deployment
//	//podList := &corev1.PodList{}
//	//listOpts := []client.ListOption{
//	//	client.InNamespace(instance.Namespace),
//	//	client.MatchingLabels(labelsForNuxeo(instance.Name)),
//	//}
//	//if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
//	//	reqLogger.Error(err, "Failed to list pods", "instance.Namespace", instance.Namespace, "instance.Name", instance.Name)
//	//	return reconcile.Result{}, err
//	//}
//	//podNames := getPodNames(podList.Items)
//	//
//	//// Update status.Nodes if needed
//	//if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
//	//	instance.Status.Nodes = podNames
//	//	err := r.client.Status().Update(context.TODO(), instance) // TODO?
//	//	if err != nil {
//	//		reqLogger.Error(err, "Failed to update Nuxeo status")
//	//		return reconcile.Result{}, err
//	//	}
//	//}
//	return reconcile.Result{}, nil
//}
//
//// deploymentForNuxeo returns a nuxeo Deployment object
//func (r *ReconcileNuxeo) deploymentForNuxeo(nux *nuxeov1alpha1.Nuxeo) *appsv1.Deployment {
//	ls := labelsForNuxeo(nux.Name)
//
//	dep := &appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      nux.Name,
//			Namespace: nux.Namespace,
//		},
//		Spec: appsv1.DeploymentSpec{
//			Selector: &metav1.LabelSelector{
//				MatchLabels: ls,
//			},
//			Template: corev1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{
//					Labels: ls,
//				},
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{{
//						Image:   "nusybox",  // "nuxeo:10.10",
//						Name:    "nuxeo",
//						Command: []string{"sleep", "3600"},
//						Ports: []corev1.ContainerPort{{
//							ContainerPort: 8080,
//							Name:          "nuxeo",
//						}},
//					}},
//				},
//			},
//		},
//	}
//	// Set Nuxeo instance as the owner and controller
//	controllerutil.SetControllerReference(nux, dep, r.scheme)
//	return dep
//}
//
//// labelsForNuxeo returns the labels for selecting the resources
//// belonging to the given Nuxeo CR name.
//func labelsForNuxeo(name string) map[string]string {
//	return map[string]string{"app": "nuxeo", "nuxeo_cr": name} // TODO nuxeo_cr?
//}
//
//// Reconcile the passed NodeSet to it's matching Deployment
//func reconcileNodeSet(r *ReconcileNuxeo, nodeSet nuxeov1alpha1.NodeSet, instance *nuxeov1alpha1.Nuxeo,
//	reqLogger logr.Logger) (reconcile.Result, error) {
//	found := &appsv1.Deployment{}
//	err := r.client.Get(context.TODO(), types.NamespacedName{Name: nodeSet.Name, Namespace: instance.Namespace}, found)
//	if err != nil && errors.IsNotFound(err) {
//		// Define a new deployment
//		// TODO HOW TO MERGE DEFAULT WITH EXPLICIT
//		dep := r.deploymentForNuxeo(instance)
//		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
//		err = r.client.Create(context.TODO(), dep)
//		if err != nil {
//			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
//			return reconcile.Result{}, err
//		}
//		// Deployment created successfully - return and requeue
//		return reconcile.Result{Requeue: true}, nil
//	} else if err != nil {
//		reqLogger.Error(err, "Failed to get Deployment")
//		return reconcile.Result{}, err
//	}
//	// TODO
//	// get actual
//	// get expected
//	// if not same
//	//   make actual look like expected
//	//   call r.Client.Update(expected)
//	return reconcile.Result{Requeue: true}, nil
//}