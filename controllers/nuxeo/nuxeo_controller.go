/*
Copyright 2020 Eric Ace.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nuxeo

import (
	"context"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nuxeov1alpha1 "github.com/aceeric/nuxeo-operator/api/v1alpha1"
)

// NuxeoReconciler reconciles a Nuxeo object
type NuxeoReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=appzygy.net,resources=nuxeos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=appzygy.net,resources=nuxeos/status,verbs=get;update;patch

func (r *NuxeoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nuxeo", req.NamespacedName)
	return r.doReconcile(req)
}

func (r *NuxeoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.controllerConfig(); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&nuxeov1alpha1.Nuxeo{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&routev1.Route{}).
		Owns(&v1beta1.Ingress{}).
		Complete(r)
}
