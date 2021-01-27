package controllers

import (
	"github.com/maistra/istio-workspace/controllers/session"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, session.Add)
}
