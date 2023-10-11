package bridge

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

type zapOtelEncoder-test struct {
	zapcore.Encoder

	OtelAttributes []attribute.KeyValue
}

func newZapOtelEncoder(numberOfFields int) *zapOtelEncoder {
	return &zapOtelEncoder{
		OtelAttributes: make([]attribute.KeyValue, 0, numberOfFields),
	}
}

func (z *zapOtelEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	// TODO: use array encoder to add homogeneous arrays to the record
	return nil
}

func (z *zapOtelEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	// TODO: use object encoder to add k8s fields like in
	// https://github.com/kubernetes-sigs/controller-runtime/blob/c93e2abcb28eb71fccad7dc565f0547cc07e5566/pkg/log/zap/kube_helpers.go#L49
	return nil
}

func (z *zapOtelEncoder) AddBinary(key string, value []byte) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, "<binary data>"))
}

func (z *zapOtelEncoder) AddByteString(key string, value []byte) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, string(value)))
}

func (z *zapOtelEncoder) AddBool(key string, value bool) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Bool(key, value))
}

func (z *zapOtelEncoder) AddComplex128(key string, value complex128) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, "<complex128>"))
}

func (z *zapOtelEncoder) AddComplex64(key string, value complex64) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, "<complex64>"))
}

func (z *zapOtelEncoder) AddDuration(key string, value time.Duration) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, value.String()))
}

func (z *zapOtelEncoder) AddFloat64(key string, value float64) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Float64(key, value))
}

func (z *zapOtelEncoder) AddFloat32(key string, value float32) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Float64(key, float64(value)))
}

func (z *zapOtelEncoder) AddInt(key string, value int) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int(key, int(value)))
}

func (z *zapOtelEncoder) AddInt64(key string, value int64) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, value))
}

func (z *zapOtelEncoder) AddInt32(key string, value int32) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddInt16(key string, value int16) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddInt8(key string, value int8) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddString(key, value string) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, value))
}

func (z *zapOtelEncoder) AddTime(key string, value time.Time) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, value.String()))
}

func (z *zapOtelEncoder) AddUint(key string, value uint) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int(key, int(value)))
}

func (z *zapOtelEncoder) AddUint64(key string, value uint64) {
	asInt64 := int64(value)
	if asInt64 > 0 {
		z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, asInt64))
	} else {
		z.OtelAttributes = append(z.OtelAttributes, attribute.String(key, "<overflowed uint64>"))
	}
}

func (z *zapOtelEncoder) AddUint32(key string, value uint32) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddUint16(key string, value uint16) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddUint8(key string, value uint8) {
	z.OtelAttributes = append(z.OtelAttributes, attribute.Int64(key, int64(value)))
}

func (z *zapOtelEncoder) AddUintptr(key string, value uintptr) {
	// ignoring pointers
}

func (z *zapOtelEncoder) AddReflected(key string, value interface{}) error {
	// TODO: add some kube aware reflection like:
	// https://github.com/kubernetes-sigs/controller-runtime/blob/c93e2abcb28eb71fccad7dc565f0547cc07e5566/pkg/log/zap/kube_helpers.go#L49
	return nil
}

func (z *zapOtelEncoder) OpenNamespace(key string) {
	// TODO: how should this be translated to opentelemetry?
}
