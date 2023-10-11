# opentelemetry-zap-bridge

This module convert every log record written to zap logger in golang into OpenTelemetry SDK log record and export it directly from the application

## Project Life Cycle

This project depends on the experimental [opentelemetry-logs-go](https://github.com/agoda-com/opentelemetry-logs-go) project, and is thus experimental as well. It is recommend to use with caution.

## Motivation

Out of the 3 pillars of observability, `logging` is the most mature one. There are many logging libraries and frameworks, and many of them are already used in production. A popular framework for logging in go is [zap logger](https://github.com/uber-go/zap) which this project targets. There are also existing practices and standards 
for how to write logs pipelines that ship these logs to some destination where they are later processed and indexed to aid in system operation tasks.

OpenTelemetry is a new standard for observability, and it is still in its infancy. 
The OpenTelemetry ecosystem offers a rich set of tools to deliver a high quality observability implementation for modern cloud systems, such as:
- [OpenTelemetry Collector]() - a high performance, vendor agnostic, logs processing pipeline with dozens of processors and exporters.
- [OTLP protocol]() - a vendor agnostic, high performance, modern logs protocol.
- Unified framework - one framework to handle all observability signals in a consistent and unified matter.

and many more.

OpenTelemetry defines the [Logs Bridge API](https://opentelemetry.io/docs/specs/otel/logs/bridge-api/):

> It is provided for logging library authors to build log appenders, which use this API to bridge between existing logging libraries and the OpenTelemetry log data model.

This package is an implementation for such bridge for the zap logger in golang.

Overtime, it is expected that more and more systems will adopt OpenTelemetry as their observability standard. This project suggests another piece of the big pazzel, which is to ship logs directly from the application to the OpenTelemetry Collector, using the OTLP protocol. This is done by converting every log record written to zap logger into an OpenTelemetry log record, and then export it using the standard OpenTelemetry SDK and exporters.

### Benefits

- **Minimal change to existing code** - you only need to change one place in your code where you initialize the zap logger, and add a single line of code to attach the OpenTelemetry SDK to it.
- **Less format conversion** - log agents like [fluentbit](https://fluentbit.io/), [logstash](https://www.elastic.co/logstash), or even the [OpenTelemetry collector](https://opentelemetry.io/docs/collector/), usually consume stderr or tails a file to extract log records, and then convert them to a format that can be processed by the new component in the log pipeline. When next component is an OTLP receiver, we can skip this step and convert the zap log records directly to the OpenTelemetry format.
- **No need to maintain agents** - no need for sidecar containers or daemonsets to run log agents, which means less resources and less maintenance.
- **Simpler log pipeline** - the log pipeline is simpler because it only needs to handle one format, the OpenTelemetry format.
- **Modern** - OpenTelemetry is a modern standard, and it is expected to be adopted by more and more systems in the future. This project is a step towards that future.
- **Standard user experience** - The OpenTelemetry SDK provides a standard experience for all programming languages and observability signals, including logs. This means that you can use the environment variables you know and love like `OTEL_SERVICE_NAME` and `OTEL_EXPORTER_OTLP_ENDPOINT` to your service just as you would do with traces, metrics, in any other otel compliant component in your system.

### Drawbacks

- **Experimental** - This project is experimental, and it depends on the experimental [opentelemetry-logs-go]() project. There are still many useful features which are missing and no test coverage. It is recommended to use with caution and research the project before using it in production.
- **No log agent** - This project does not provide a log agent, which means that you will need to use the OpenTelemetry Collector to receive the logs and process them. This is not a drawback per se, but it is a different approach than what you might be used to.
- **Less performant** - This project introduces more overhead on the process that writes the logs, because it needs to convert the log records to the OpenTelemetry format and then use networking to export them from the process. This is not a drawback per se, as these resources were already consumed and payed indirectly by the log agent, but for applications that targeting to squeeze every bit of performance, this might be a drawback.
- **Code change** - This project requires a code change to the application, which means that you will need to recompile and redeploy your application. While it is recommended to process as much as possible outside of the process, in OpenTelemetry collectors, application developers will need to maintain this code - bumping versions, debugging issues, etc. Some organizations prefer for these tasks to be handled by the operations team, which needs to setup log pipelines, without touching any line of code.

## Mechanism

When you write this in your code:
```go
    logger, _ := zap.NewProduction()
```
you are creating a `zap.Logger` instance which is a wrapper around `zapcore.Core` instance. 
`zapcore.Core` is the actual `log backend` that implement a `Write` method that outputs a log records to some destination.

This module provides a `zapcore.Core` implementation that initialize an OpenTelemetry logger SDK,
which is then used to export log records to some destination with an opentelemetry exporter.

## Configuration

Currently the supported configuration method is via OpenTelemetry standard environment variables which you can find [here](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/configuration/sdk-environment-variables.md) and [here for otlp exporter](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md).

Specifically, I found the following environment variables useful:
- `OTEL_SERVICE_NAME` - to add a service name to the logs records resource, which is must-have for them to be useful.
- `OTEL_SDK_DISABLED=true` - to disable the OpenTelemetry SDK, which allows a one switch to turn off otel SDK if needed (for example, local development, or as a step in migration)
- `OTEL_EXPORTER_OTLP_PROTOCOL=grpc` - the default otlp protocol is http, which is less performant than grpc. this is how you can change it.
- `OTEL_EXPORTER_OTLP_INSECURE=true` - **Only if you use internal OpenTelemetry collector gateway** set when you export your logs over insecure (not TLS) connection. It is recommended to run an OpenTelemetry collector in your cluster to server as a gateway to your log vendor or log backend which your use. This redouces the overhead on the application that do not need to encrypt and export to a local receiver in the cluster.
- `OTEL_EXPORTER_OTLP_ENDPOINT`/`OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` - the endpoint of the OpenTelemetry collector to which you want to export your logs. recommended: local OpenTelemetry collector gateway in your cluster. You can also use and otlp compatible receiver like you log vendor endpoints.
- `OTEL_EXPORTER_OTLP_HEADERS` / `OTEL_EXPORTER_OTLP_LOGS_HEADERS` - If you use a vendor, you might get instructions to add some headers to your requests for authentication or other purposes. This is what you need to use for it to work.

The rest of the configuration has reasonable defaults, but you can find browse through them to fine tune to your needs.

## Usage

This package provides a `NewOtelZapCore` function that returns a `zapcore.Core` instance, there is also a wrapper `AttachToZapLogger` that is used to attache to existing zap logger. You can choose to use it in one of the following ways:

### With zap.Logger

If you have a `zap.Logger` instance, which you obtained with `zap.NewProduction()` / `zap.New(...)` etc, you can attach a new `zapcore.Core` to this logger using the `AttachToZapLogger` function:

```go
	logger, _ := zap.NewProduction()
	logger = bridge.AttachToZapLogger(logger, otelServiceName)
```

### With Kubernetes controller-runtime Package

If you use [`kubebuilder`](https://github.com/kubernetes-sigs/kubebuilder), it autogenerates a `zap.Logger` for you:

```go
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
```

You can convert this code to attach an otel sdk to it like this:
```go
import (
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "github.com/go-logr/zapr"
	bridge "github.com/keyval-dev/opentelemetry-zap-bridge"
)

func main() {
    // ...
	logger := zap.NewRaw(zap.UseFlagOptions(&opts))
	logger = bridge.AttachToZapLogger(logger)
	ctrl.SetLogger(zapr.NewLogger(logger))
    // ...
}
```

## With zap.New

You can  use the `NewOtelZapCore` function to create a new `zapcore.Core` instance and use it with `zap.New` to create just otel logger without console or file logging:

```go
import (
    "go.uber.org/zap"
    bridge "github.com/keyval-dev/opentelemetry-zap-bridge"
)

func main() {
    // ...
    logger, _ := zap.New(bridge.NewOtelZapCore(otelServiceName))
    // ...
}
```

Notice that it is not recommended for development, as you will not see any logs in your console. 
In production, this might cause no logs to be written to the file, thus making `kubectl logs` useless.
It will however, make your application more performant, as it will not encode and write to the file.

## With other zapcore.Core

If you are advanced user and you can create your own `zapcore.Core` instance, you can use the `NewOtelZapCore` function to create a new otel sdk `zapcore.Core` instance and use it with `zapcore.NewTee` to combine it with your existing `zapcore.Core` instance:

```go
import (
    "go.uber.org/zap/zapcore"
    bridge "github.com/keyval-dev/opentelemetry-zap-bridge"
)

func main() {
    // yourZapCore := ...
    otelZapLogger := NewOtelZapCore()
    combinedCore := zapcore.NewTee(core, yourZapCore)
    logger := zap.New(combinedCore)
```

## Alternatives

There are other ways to achieve similar results, and each user should evaluate the options and choose the one that fits best to their needs.

Some alternatives are:
- Use log agent like [fluentbit](https://fluentbit.io/) or [logstash](https://www.elastic.co/logstash), or [OpenTelemetry collector](https://grafana.com/docs/opentelemetry/collector/send-logs-to-loki/kubernetes-logs/) to extract logs from the application and convert them to the OpenTelemetry format. This is the most common approach today, and it is being the popular choice for many years. It is a proven approach, and it is recommended to use it if you are not sure.
- If your system runs in kubernetes, you can setup logging (and metrics and traces if you like) with few minutes using OpenSource projects like [Odigos](https://docs.odigos.io/overview). It basically injects a DaemonSet to collect all logs from each k8s node, ship to an opentelemetry gateway collector which is deployed in the cluster, and then export to all the popular log destinations, SAAS or self managed. Disclaimer: This repo is brought to you by [keyval](https://odigos.io/), which is the company behind odigos.
- If you are just getting started, and not yet ready to deploy a full blown logs pipeline, you can simply use `zap.NewProduction` and write logs to stdout/stderr or file, and consume them when needed with `kubectl logs`/`docker logs` or simply read the file.

## Contributing

There are many features and enhancements currently lacking. We welcome and issue and pull request to help improve this project.

## Talk to us

This project is brought to you with ❤️ by the [odigos](https://odigos.io/) team.

We are happy to hear your feedback and answer your questions. Or just say hi and chat about observability.

If you want to try out [odigos](https://odigos.io/) let's talk about it too.
