package model

import (
	"reflect"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	resourceVectors = []string{"source_namespace", "source_kind", "source_action"}

	resources = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resources_total",
			Help: "Number of resources processed",
		},
		resourceVectors,
	)
	resourceFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resources_failures_total",
			Help: "Number of failed resources",
		},
		resourceVectors,
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(resources, resourceFailures)
}

// Sync is the entry point for ensuring the desired state for the given Ref is up-to-date.
type Sync func(SessionContext, Ref, ModificatorController, LocatedReporter, ModificatorStatusReporter)

func NewSync(locators []Locator, modificators []Modificator) Sync {
	return func(context SessionContext, ref Ref, modify ModificatorController, locatedReporter LocatedReporter, modificationReporter ModificatorStatusReporter) {
		instrumentedReporter := instrumentedModificationStatusReporter(modificationReporter)
		located := LocatorStore{}
		for _, locator := range locators {
			err := locator(
				context,
				ref,
				located.Store,
				located.Report,
			)

			if err != nil {
				context.Log.Error(err, "locating failed", "locator", runtime.FuncForPC(reflect.ValueOf(locator).Pointer()).Name())
			}
		}
		if !modify(located.Store) {
			return
		}

		locatedReporter(located.Store)

		for _, modificator := range modificators {
			modificator(context, ref, located.Store, instrumentedReporter)
		}
	}
}

func instrumentedModificationStatusReporter(report ModificatorStatusReporter) func(ModificatorStatus) {
	return func(status ModificatorStatus) {
		resources.WithLabelValues(status.Namespace, status.Kind, string(status.Action)).Inc()
		if !status.Success {
			resourceFailures.WithLabelValues(status.Namespace, status.Kind, string(status.Action)).Inc()
		}
		report(status)
	}
}
