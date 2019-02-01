package controller

import (
	"github.com/aslakknutsen/istio-workspace/pkg/controller/session"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, session.Add)
}
