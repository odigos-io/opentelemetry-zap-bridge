package bridge

import (
	"context"
	"os"
	"strings"

	autosdk "github.com/agoda-com/opentelemetry-logs-go/autoconfigure/sdk/logs"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	otelSdkDisabled            = "OTEL_SDK_DISABLED"
	instrumentationLibraryName = "github.com/odigos-io/opentelemetry-zap-bridge"
)

type OtelZapCore struct {
	zapcore.Core

	logger logs.Logger
}

// this function creates a new zapcore.Core that can be used with zap.New()
// this instance will translate zap logs to opentelemetry logs and export them
//
// serviceName is required and is used as a resource attribute in the reported telemetry.
// TODO: we should probably extend this to support more options like additional resource attributes
// TODO2: should we also support a way to configure other components and configuration options?
// like exporters, processors, etc.
// Currently a user can configure the SDK only via environment variables which is fair enough
// but advanced users might want more control.
func NewOtelZapCore() zapcore.Core {
	ctx := context.Background()
	loggerProvider := autosdk.NewLoggerProvider(ctx)
	// TODO: what scope name should we use?
	// should we allow this to be conbfigurable?
	// how do we record correct scope name with zap?
	logger := loggerProvider.Logger(instrumentationLibraryName)

	return &OtelZapCore{
		logger: logger,
	}
}

// NewCustomOtelZapCore allows for creating zapcore with customized logger,
// giving thus greater flexibility
func NewCustomOtelZapCore(logger logs.Logger) zapcore.Core {
	return &OtelZapCore{
		logger: logger,
	}
}

// TODO: I guess there is more idomatic way to do this in go
func AttachToZapLogger(logger *zap.Logger) *zap.Logger {
	otelSdlDisabled, defined := os.LookupEnv(otelSdkDisabled)
	// do not register the otel zap core if user set OTEL_SDK_DISABLED=true
	if defined && strings.ToLower(otelSdlDisabled) == "true" {
		return logger
	}

	return logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		otelZapLogger := NewOtelZapCore()
		return zapcore.NewTee(core, otelZapLogger)
	}))
}

// TODO: see how it is implemented in zapcore and consider adding it here as well
func (o *OtelZapCore) Enabled(zapcore.Level) bool {
	return true
}

// TODO: implement this and add the fields to each new log record created
func (o *OtelZapCore) With(fields []zapcore.Field) zapcore.Core {
	return o
}

func (o *OtelZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, o)
}

func (o *OtelZapCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	otelSeverity := convertZapLevelToOtelSeverity(ent.Level)
	severityText := ent.Level.String()

	otelEncoder := newZapOtelEncoder(len(fields))
	for _, field := range fields {
		field.AddTo(otelEncoder)
	}

	lrc := logs.LogRecordConfig{
		Timestamp:      &ent.Time,
		Body:           &ent.Message,
		SeverityNumber: &otelSeverity,
		SeverityText:   &severityText,
		Attributes:     &otelEncoder.OtelAttributes,
	}
	logRecord := logs.NewLogRecord(lrc)
	o.logger.Emit(logRecord)

	return nil
}

func convertZapLevelToOtelSeverity(zapLevel zapcore.Level) logs.SeverityNumber {
	// zap levels seems to be int8 and can be negative!
	// it appears that the debug level is -1, and 0 should not be used for otel severity
	// as stated in the docs of the otel enum.
	// thus I will add 2 and hope for the best
	//
	// TODO: Otel seems to be using different levels of severities which we can map.
	// for example - zapcore.WarnLevel can be mapped to otel's `WARN SeverityNumber = 13`.
	// but - is it always true? what happens when zap is triggered by logr with something like
	// `log.V(2).Info("message")`?
	// We should check these cases and possibly use more accurate mapping.
	return logs.SeverityNumber(zapLevel + 2)
}
