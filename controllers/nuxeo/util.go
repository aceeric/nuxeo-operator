package nuxeo

import (
	"fmt"

	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
)

var NuxeoServiceAccountName = "nuxeo"

// GetNuxeoContainer walks the container array in the passed deployment and returns a ref to the container
// named "nuxeo". If not found, returns a nil container ref and an error.
func GetNuxeoContainer(dep *v1.Deployment) (*v12.Container, error) {
	for i := 0; i < len(dep.Spec.Template.Spec.Containers); i++ {
		if dep.Spec.Template.Spec.Containers[i].Name == "nuxeo" {
			return &dep.Spec.Template.Spec.Containers[i], nil
		}
	}
	return nil, fmt.Errorf("could not find a container named 'nuxeo' in the deployment")
}
