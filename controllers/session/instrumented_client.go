package session

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	resourceVectors = []string{"client_func", "source_kind"}

	resources = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_http_total",
			Help: "Number of API calls to API server",
		},
		resourceVectors,
	)

	resourceFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_http_failures_total",
			Help: "Number of API calls to API server that failed",
		},
		resourceVectors,
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(resources, resourceFailures)
}

func NewInstrumentedClient(c client.Client) client.Client {
	return &InstrumentedClient{c: c}
}

type InstrumentedClient struct {
	c client.Client
}

func (i *InstrumentedClient) Scheme() *runtime.Scheme {
	return i.c.Scheme()
}

func (i *InstrumentedClient) RESTMapper() meta.RESTMapper {
	return i.c.RESTMapper()
}

func (i *InstrumentedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return record(i.c.Get(ctx, key, obj), "get", i.getObjectKind(obj))
}

func (i *InstrumentedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return record(i.c.List(ctx, list, opts...), "list", i.getObjectKind(list))
}

func (i *InstrumentedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return record(i.c.Create(ctx, obj, opts...), "create", i.getObjectKind(obj))
}

func (i *InstrumentedClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return record(i.c.Delete(ctx, obj, opts...), "delete", i.getObjectKind(obj))
}

func (i *InstrumentedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return record(i.c.Update(ctx, obj, opts...), "update", i.getObjectKind(obj))
}

func (i *InstrumentedClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return record(i.c.Patch(ctx, obj, patch, opts...), "patch", i.getObjectKind(obj))
}

func (i *InstrumentedClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return record(i.c.DeleteAllOf(ctx, obj, opts...), "delete_all_of", i.getObjectKind(obj))
}

func (i *InstrumentedClient) Status() client.StatusWriter {
	return &InstrumentedStatusWriter{c: i.c, p: i}
}

type InstrumentedStatusWriter struct {
	c client.StatusClient
	p *InstrumentedClient
}

func (i *InstrumentedStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return record(i.c.Status().Update(ctx, obj, opts...), "status_update", i.p.getObjectKind(obj))
}

func (i *InstrumentedStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return record(i.c.Status().Patch(ctx, obj, patch, opts...), "status_patch", i.p.getObjectKind(obj))
}

func (i *InstrumentedClient) getObjectKind(obj runtime.Object) string {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		kinds, _, err := i.Scheme().ObjectKinds(obj)
		if err != nil {
			return ""
		}
		if len(kinds) > 0 {
			return kinds[0].Kind
		}
	}

	return kind
}

func record(err error, name, kind string) error {
	if err != nil && !k8sErrors.IsNotFound(err) {
		resourceFailures.WithLabelValues(name, kind).Inc()
	} else {
		resources.WithLabelValues(name, kind).Inc()
	}

	return err
}
