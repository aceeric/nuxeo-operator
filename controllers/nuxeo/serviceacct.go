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
	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServiceAccount creates a service account for the Nuxeo deployments to run under. At present, there isn't
// anything in the service account spec - so this is just a placeholder in case any special service-related
// capabilities are needed in the future
func (r *NuxeoReconciler) reconcileServiceAccount(instance *v1alpha1.Nuxeo) error {
	svcAcctName := NuxeoServiceAccountName
	expected, err := r.defaultServiceAccount(instance, svcAcctName)
	if err != nil {
		return err
	}
	_, err = r.addOrUpdate(svcAcctName, instance.Namespace, expected, &corev1.ServiceAccount{}, util.NopComparer)
	return err
}

// defaultServiceAccount creates and returns a service account struct
func (r *NuxeoReconciler) defaultServiceAccount(instance *v1alpha1.Nuxeo,
	svcAcctName string) (*corev1.ServiceAccount, error) {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcAcctName,
			Namespace: instance.Namespace,
		},
	}
	_ = controllerutil.SetControllerReference(instance, &sa, r.Scheme)
	return &sa, nil
}
