// Based on https://github.com/operator-framework/operator-lib/blob/8c3d48f55639528bcee4432b570bc6671900b75d/handler/enqueue_annotation.go
// Main changes are about allowing to store multiple values of a given annotation as comma-separated list.
//
// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package reference

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	crtHandler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/maistra/istio-workspace/pkg/log"
)

const (
	// NamespacedNameAnnotation is an annotation whose value encodes the name and namespace of a resource to
	// reconcile when a resource containing this annotation changes. Valid values are of the form
	// `<namespace>/<name>` for namespace-scoped owners and `<name>` for cluster-scoped owners.
	NamespacedNameAnnotation = "maistra.io/istio-workspaces"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "enqueue")
	}
)

// EnqueueRequestForAnnotation enqueues Request containing the Name and Namespace specified in the
// annotations of the object that is the source of the Event. The source of the event triggers reconciliation
// of the parent resource which is identified by annotations. `NamespacedNameAnnotation` uniquely identify an owner resource to reconcile.
//
// handler.EnqueueRequestForAnnotation can be used to trigger reconciliation of resources which are
// cross-referenced.  This allows a namespace-scoped dependent to trigger reconciliation of an owner
// which is in a different namespace, and a cluster-scoped dependent can trigger the reconciliation
// of a namespace(scoped)-owner.
//
// As an example, consider the case where we would like to watch virtualservices based on which we reconcile
// Istio Workspace Sessions. We rely on using annotations as a way of tracking the resources that we manipulate
// to avoid running into the built-in garbage collection included with OwnerRef. Using this approach we could implement the following:
//
//	if err := c.Watch(&source.Kind{
//		// Watch VirtualService
//		Type: &istio.VirtualService{}},
//
//		// Enqueue ReplicaSet reconcile requests using the namespacedName annotation value in the request.
//		&handler.EnqueueRequestForAnnotation{schema.GroupKind{Group:"istio.io", Kind:"VirtualService"}}); err != nil {
//			entryLog.Error(err, "unable to watch ClusterRole")
//			os.Exit(1)
//		}
//	}
//
// With this watch, the Istio Workspace controller would receive a request to reconcile
// "namespace/session1,session2" based on a change to a VirtualService that has the following annotations:
//
//	annotations:
//		maistra.io/istio-workspaces: "namespace/session1,session2"
//
// Note: multiple sessions can manipulate the same resource, therefore the annonation is a list defined as comma-separated values.
//
// Though an annotation-based watch handler removes the boundaries set by native owner reference implementation,
// the garbage collector still respects the scope restrictions. For example,
// if a parent creates a child resource across scopes not supported by owner references, it becomes the
// responsibility of the reconciler to clean up the child resource. Hence, the resource utilizing this handler
// SHOULD ALWAYS BE IMPLEMENTED WITH A FINALIZER.
type EnqueueRequestForAnnotation struct {
	Type schema.GroupKind
}

var _ crtHandler.EventHandler = &EnqueueRequestForAnnotation{}

// Create reacts to create event and schedules reconcile.
func (e *EnqueueRequestForAnnotation) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if ok, reqs := e.getAnnotationRequests(evt.Object); ok {
		logChange("Reference Object created", evt.Object, reqs)
		e.addToQueue(q, reqs)
	}
}

// Update reacts to update event and schedules reconcile.
func (e *EnqueueRequestForAnnotation) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if ok, reqs := e.getAnnotationRequests(evt.ObjectOld); ok {
		logChange("Reference Object updated", evt.ObjectOld, reqs)
		e.addToQueue(q, reqs)
	}
	if ok, reqs := e.getAnnotationRequests(evt.ObjectNew); ok {
		logChange("Reference Object updated", evt.ObjectNew, reqs)
		e.addToQueue(q, reqs)
	}
}

// Delete reacts to delete event and schedules reconcile.
func (e *EnqueueRequestForAnnotation) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if ok, reqs := e.getAnnotationRequests(evt.Object); ok {
		logChange("Reference Object deleted", evt.Object, reqs)
		e.addToQueue(q, reqs)
	}
}

// Generic reacts to any other event (e.g. reconcile Autoscaling, or a Webhook) and schedules reconcile.
func (e *EnqueueRequestForAnnotation) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	if ok, reqs := e.getAnnotationRequests(evt.Object); ok {
		logChange("Reference Object generic", evt.Object, reqs)
		e.addToQueue(q, reqs)
	}
}

// Add helps in adding 'NamespacedNameAnnotation' to object based on
// the NamespaceName. The object gets the annotations from owner's namespace, name, group
// and kind. In other terms, object can be said to be the dependent having annotations from the owner.
// When a watch is set on the object, the annotations help to identify the owner and trigger reconciliation.
// Annotations are accumulated as a comma-separated list.
func Add(owner types.NamespacedName, object client.Object) error {
	if owner.Namespace == "" {
		return fmt.Errorf("%T does not have a namespace, cannot call Add", owner) //nolint:goerr113 //reason useful to have owner in error
	}
	if owner.Name == "" {
		return fmt.Errorf("%T does not have a name, cannot call Add", owner) //nolint:goerr113 //reason useful to have owner in error
	}

	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	references := strings.Split(annotations[NamespacedNameAnnotation], ",")
	reference := fmt.Sprintf("%s/%s", owner.Namespace, owner.Name)

	if len(references) == 1 && references[0] == "" {
		references[0] = reference
	} else {
		for _, r := range references {
			if r == reference {
				return nil
			}
		}
		references = append(references, reference)
	}

	annotations[NamespacedNameAnnotation] = strings.Join(references, ",")

	object.SetAnnotations(annotations)

	return nil
}

// Remove helps in removing 'NamespacedNameAnnotation' from object based on
// the NamespaceName. The object gets the annotations from owner's namespace, name, group
// and kind. Annotations are accumulated as a comma-separated list, thus removal will change the content of the list.
func Remove(owner types.NamespacedName, object client.Object) error {
	if owner.Namespace == "" {
		return fmt.Errorf("%T does not have a namespace, cannot call Remove", owner) //nolint:goerr113 //reason useful to have owner in error
	}
	if owner.Name == "" {
		return fmt.Errorf("%T does not have a name, cannot call Remove", owner) //nolint:goerr113 //reason useful to have owner in error
	}

	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	existingReferences := strings.Split(annotations[NamespacedNameAnnotation], ",")
	reference := fmt.Sprintf("%s/%s", owner.Namespace, owner.Name)

	var references []string
	for _, r := range existingReferences {
		if r != reference {
			references = append(references, r)
		}
	}

	annotations[NamespacedNameAnnotation] = strings.Join(references, ",")

	object.SetAnnotations(annotations)

	return nil
}

// Get converts annotations defined for NamespacedNameAnnotation as comma-separated list to a slice.
func Get(object client.Object) []types.NamespacedName {
	var typeNames []types.NamespacedName
	annotations := object.GetAnnotations()
	if annotations == nil {
		return typeNames
	}

	existingReferences := strings.Split(annotations[NamespacedNameAnnotation], ",")
	for _, ref := range existingReferences {
		if ref != "" {
			typeNames = append(typeNames, parseNamespacedName(ref))
		}
	}

	return typeNames
}

// addToQueue adds a slice of Reconcile Requests to the queue.
func (e *EnqueueRequestForAnnotation) addToQueue(q workqueue.RateLimitingInterface, requests []reconcile.Request) {
	for _, request := range requests {
		q.Add(request)
	}
}

// getAnnotationRequests checks if the provided object has the annotations so as to enqueue the reconcile request.
func (e *EnqueueRequestForAnnotation) getAnnotationRequests(object client.Object) (bool, []reconcile.Request) {
	requests := []reconcile.Request{}

	typeNames := Get(object)
	if len(typeNames) == 0 {
		return false, requests
	}

	for _, typeName := range typeNames {
		requests = append(requests, reconcile.Request{NamespacedName: typeName})
	}

	return true, requests
}

// parseNamespacedName parses the provided string to extract the namespace and name into a
// types.NamespacedName. The edge case of empty string is handled prior to calling this function.
func parseNamespacedName(namespacedNameString string) types.NamespacedName {
	values := strings.SplitN(namespacedNameString, "/", 2)

	switch len(values) {
	case 1:
		return types.NamespacedName{Name: values[0]}
	default:
		return types.NamespacedName{Namespace: values[0], Name: values[1]}
	}
}

func logChange(message string, object client.Object, requests []reconcile.Request) {
	logger().Info(message,
		"kind", object.GetObjectKind().GroupVersionKind().String(),
		"namespace", object.GetNamespace(),
		"name", object.GetName(),
		"targets", requests)
}
