package controller

import (
	"nuxeo-operator/pkg/controller/nuxeo"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, nuxeo.Add)
}
