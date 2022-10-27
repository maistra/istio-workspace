package log_test

import (
	"testing"

	"emperror.dev/errors"
	"github.com/maistra/istio-workspace/pkg/log"
)

func TestEmperrorIntegration(t *testing.T) {
	errs := []error{
		errors.WithDetails(errors.New("child1 level message"), "name", "test1-error", "context", "test1"),
		errors.WithDetails(errors.New("child2 level message"), "name", "test2-error", "context", "test2", "test2", "true"),
	}
	err := errors.Combine(errs...)
	err = errors.WrapIfWithDetails(err, "top level error", "context", "parent")

	logger := log.CreateOperatorAwareLogger("emperror-test")
	logger.Error(err, "log level msg", "name", "log-error", "namespace", "test-space")
}

func TestEmperrorIntegration2(t *testing.T) {
	err := errors.WrapIfWithDetails(
		errors.WrapIfWithDetails(
			errors.WrapIfWithDetails(
				errors.NewWithDetails("failed to load", "context", "inner", "inner", "true"),
				"caught inner",
				"context", "child2", "child2", "true",
			),
			"caught child",
			"context", "child", "child", "true",
		),
		"top level error",
		"context", "parent", "parent", "true")

	logger := log.CreateOperatorAwareLogger("emperror-test")
	logger.Error(err, "log level msg", "name", "log-error", "namespace", "test-space")
}
