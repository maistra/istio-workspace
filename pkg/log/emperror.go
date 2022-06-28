package log

import (
	"emperror.dev/errors"
	"github.com/go-logr/logr"
)

// Assert conformance to the interface.
var _ logr.LogSink = &errorCollector{}

type errorCollector struct {
	logr.LogSink
}

func New(l logr.Logger) logr.Logger {
	return logr.New(&errorCollector{l.GetSink()})
}

func (e errorCollector) Error(err error, msg string, keysAndValues ...interface{}) {
	kv := uniqueKeys(collectDetails(err, keysAndValues)...)
	e.LogSink.Error(err, msg, kv...)
}

func uniqueKeys(kv ...interface{}) []interface{} {
	for i := 0; i < len(kv); i += 2 {
		if v, ok := kv[i].(string); ok {
			for n := i + 2; n < len(kv); n += 2 {
				if v2, ok := kv[n].(string); ok {
					if v == v2 {
						kv = append(kv[:n], kv[n+2:]...)
						n -= 2
					}
				}
			}
		}
	}

	return kv
}

func collectDetails(err error, keysAndValues []interface{}) []interface{} {
	kv := keysAndValues
	if details := errors.GetDetails(err); len(details) > 0 {
		kv = append(details, kv...)
	}

	type errorCollection interface {
		Errors() []error
	}

	if errs, ok := err.(errorCollection); ok { //nolint:errorlint // no need for errors.As as we're doing the unwrapping
		for _, er := range errs.Errors() {
			kv = collectDetails(er, kv)
		}
	}
	if cause := errors.Unwrap(err); cause != nil {
		kv = collectDetails(cause, kv)
	}

	return kv
}
