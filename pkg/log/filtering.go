package log

import (
	"time"

	"go.uber.org/zap/buffer"

	"go.uber.org/zap/zapcore"
)

// filteringEncoder enables filtering out fields.
// For example that can be useful for CLI logger if you want to skip printing additional context as JSON.
type filteringEncoder struct {
	delegate        zapcore.Encoder
	fieldsToInclude map[string]struct{}
}

func newFilteringEncoder(delegate zapcore.Encoder, includedFields ...string) filteringEncoder {
	fields := make(map[string]struct{}, len(includedFields))
	for _, field := range includedFields {
		fields[field] = struct{}{}
	}
	return filteringEncoder{delegate: delegate, fieldsToInclude: fields}
}

func (f filteringEncoder) shouldSkip(key string) bool {
	_, ok := f.fieldsToInclude[key]
	return !ok
}

func (f filteringEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	if f.shouldSkip(key) {
		return nil
	}
	return f.delegate.AddArray(key, marshaler)
}

func (f filteringEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	if f.shouldSkip(key) {
		return nil
	}
	return f.delegate.AddObject(key, marshaler)
}

func (f filteringEncoder) AddBinary(key string, value []byte) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddBinary(key, value)
}

func (f filteringEncoder) AddByteString(key string, value []byte) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddByteString(key, value)
}

func (f filteringEncoder) AddBool(key string, value bool) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddBool(key, value)
}

func (f filteringEncoder) AddComplex128(key string, value complex128) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddComplex128(key, value)
}

func (f filteringEncoder) AddComplex64(key string, value complex64) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddComplex64(key, value)
}

func (f filteringEncoder) AddDuration(key string, value time.Duration) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddDuration(key, value)
}

func (f filteringEncoder) AddFloat64(key string, value float64) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddFloat64(key, value)
}

func (f filteringEncoder) AddFloat32(key string, value float32) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddFloat32(key, value)
}

func (f filteringEncoder) AddInt(key string, value int) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddInt(key, value)
}

func (f filteringEncoder) AddInt64(key string, value int64) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddInt64(key, value)
}

func (f filteringEncoder) AddInt32(key string, value int32) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddInt32(key, value)
}

func (f filteringEncoder) AddInt16(key string, value int16) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddInt16(key, value)
}

func (f filteringEncoder) AddInt8(key string, value int8) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddInt8(key, value)
}

func (f filteringEncoder) AddString(key, value string) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddString(key, value)
}

func (f filteringEncoder) AddTime(key string, value time.Time) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddTime(key, value)
}

func (f filteringEncoder) AddUint(key string, value uint) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUint(key, value)
}

func (f filteringEncoder) AddUint64(key string, value uint64) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUint64(key, value)
}

func (f filteringEncoder) AddUint32(key string, value uint32) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUint32(key, value)
}

func (f filteringEncoder) AddUint16(key string, value uint16) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUint16(key, value)
}

func (f filteringEncoder) AddUint8(key string, value uint8) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUint8(key, value)
}

func (f filteringEncoder) AddUintptr(key string, value uintptr) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.AddUintptr(key, value)
}

func (f filteringEncoder) AddReflected(key string, value interface{}) error {
	if f.shouldSkip(key) {
		return nil
	}
	return f.delegate.AddReflected(key, value)
}

func (f filteringEncoder) OpenNamespace(key string) {
	if f.shouldSkip(key) {
		return
	}
	f.delegate.OpenNamespace(key)
}

func (f filteringEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	return f.delegate.EncodeEntry(entry, fields)
}

func (f filteringEncoder) Clone() zapcore.Encoder {
	return filteringEncoder{f.delegate.Clone(), f.fieldsToInclude}
}
