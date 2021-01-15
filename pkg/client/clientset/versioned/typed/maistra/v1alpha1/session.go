// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/maistra/istio-workspace/pkg/api/maistra/v1alpha1"
	scheme "github.com/maistra/istio-workspace/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SessionsGetter has a method to return a SessionInterface.
// A group's client should implement this interface.
type SessionsGetter interface {
	Sessions(namespace string) SessionInterface
}

// SessionInterface has methods to work with Session resources.
type SessionInterface interface {
	Create(ctx context.Context, session *v1alpha1.Session, opts v1.CreateOptions) (*v1alpha1.Session, error)
	Update(ctx context.Context, session *v1alpha1.Session, opts v1.UpdateOptions) (*v1alpha1.Session, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Session, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.SessionList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Session, err error)
	SessionExpansion
}

// sessions implements SessionInterface
type sessions struct {
	client rest.Interface
	ns     string
}

// newSessions returns a Sessions
func newSessions(c *MaistraV1alpha1Client, namespace string) *sessions {
	return &sessions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the session, and returns the corresponding session object, and an error if there is any.
func (c *sessions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Session, err error) {
	result = &v1alpha1.Session{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sessions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Sessions that match those selectors.
func (c *sessions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SessionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SessionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sessions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sessions.
func (c *sessions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sessions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a session and creates it.  Returns the server's representation of the session, and an error, if there is any.
func (c *sessions) Create(ctx context.Context, session *v1alpha1.Session, opts v1.CreateOptions) (result *v1alpha1.Session, err error) {
	result = &v1alpha1.Session{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sessions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(session).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a session and updates it. Returns the server's representation of the session, and an error, if there is any.
func (c *sessions) Update(ctx context.Context, session *v1alpha1.Session, opts v1.UpdateOptions) (result *v1alpha1.Session, err error) {
	result = &v1alpha1.Session{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sessions").
		Name(session.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(session).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the session and deletes it. Returns an error if one occurs.
func (c *sessions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sessions").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sessions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sessions").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched session.
func (c *sessions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Session, err error) {
	result = &v1alpha1.Session{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sessions").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
