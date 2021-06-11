package log

import (
	"emperror.dev/errors"
	"github.com/go-logr/logr"
)

type Emperror struct {
	logr.Logger
}

func (e Emperror) V(level int) logr.Logger { return Emperror{e.Logger.V(level)} }
func (e Emperror) WithValues(keysAndValues ...interface{}) logr.Logger {
	return Emperror{e.Logger.WithValues(keysAndValues...)}
}
func (e Emperror) WithName(name string) logr.Logger { return Emperror{e.Logger.WithName(name)} }

func (e Emperror) Error(err error, msg string, keysAndValues ...interface{}) {
	kv := uniqueKeys(collectDetails(err, keysAndValues)...)
	e.Logger.Error(err, msg, kv...)
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
