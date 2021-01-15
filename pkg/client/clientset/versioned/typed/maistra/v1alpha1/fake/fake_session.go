// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/maistra/istio-workspace/pkg/api/maistra/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSessions implements SessionInterface
type FakeSessions struct {
	Fake *FakeMaistraV1alpha1
	ns   string
}

var sessionsResource = schema.GroupVersionResource{Group: "maistra.io", Version: "v1alpha1", Resource: "sessions"}

var sessionsKind = schema.GroupVersionKind{Group: "maistra.io", Version: "v1alpha1", Kind: "Session"}

// Get takes name of the session, and returns the corresponding session object, and an error if there is any.
func (c *FakeSessions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Session, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sessionsResource, c.ns, name), &v1alpha1.Session{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Session), err
}

// List takes label and field selectors, and returns the list of Sessions that match those selectors.
func (c *FakeSessions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SessionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sessionsResource, sessionsKind, c.ns, opts), &v1alpha1.SessionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SessionList{ListMeta: obj.(*v1alpha1.SessionList).ListMeta}
	for _, item := range obj.(*v1alpha1.SessionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sessions.
func (c *FakeSessions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sessionsResource, c.ns, opts))

}

// Create takes the representation of a session and creates it.  Returns the server's representation of the session, and an error, if there is any.
func (c *FakeSessions) Create(ctx context.Context, session *v1alpha1.Session, opts v1.CreateOptions) (result *v1alpha1.Session, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sessionsResource, c.ns, session), &v1alpha1.Session{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Session), err
}

// Update takes the representation of a session and updates it. Returns the server's representation of the session, and an error, if there is any.
func (c *FakeSessions) Update(ctx context.Context, session *v1alpha1.Session, opts v1.UpdateOptions) (result *v1alpha1.Session, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sessionsResource, c.ns, session), &v1alpha1.Session{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Session), err
}

// Delete takes name of the session and deletes it. Returns an error if one occurs.
func (c *FakeSessions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sessionsResource, c.ns, name), &v1alpha1.Session{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSessions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sessionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SessionList{})
	return err
}

// Patch applies the patch and returns the patched session.
func (c *FakeSessions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Session, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sessionsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Session{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Session), err
}
