package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Metric represents a metric that can be collected by the server.
type Metric struct {
	Name        string
	Unit        string
	Description string
}

// Config holds the configuration for the telemetry.
type Config struct {
	ServiceName       string
	ServiceVersion    string
	Enabled           bool
	CollectorEndpoint string
}

type TelemetryProvider interface {
	IncreaseUserRegisteredCounter(context.Context, *slog.Logger)
	MeterInt64Counter(metric Metric) (otelmetric.Int64Counter, error)
	Shutdown(ctx context.Context)
	MeterProvider() otelmetric.MeterProvider
	TracerProvider() *trace.TracerProvider
	Logger() *slog.Logger
}

// telemetry is a wrapper around the OpenTelemetry logger, meter, and tracer.
type telemetry struct {
	lp             *log.LoggerProvider
	meterProvider  *metric.MeterProvider
	tracerProvider *trace.TracerProvider
	logger         *slog.Logger
	meter          otelmetric.Meter
	tracer         oteltrace.Tracer
	cfg            Config
}

// NewTelemetry creates a new telemetry instance.
func NewTelemetry(ctx context.Context, cfg Config) (TelemetryProvider, error) {
	rp := newResource(cfg.ServiceName, cfg.ServiceVersion)

	lp, err := newLoggerProvider(ctx, rp, cfg.CollectorEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	logger := slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			otelslog.NewHandler("movieswithfriends", otelslog.WithLoggerProvider(lp)),
		))

	mp, err := newMeterProvider(ctx, rp, cfg.CollectorEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}
	meter := mp.Meter(cfg.ServiceName)

	tp, err := newTracerProvider(ctx, rp, cfg.CollectorEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	tracer := tp.Tracer(cfg.ServiceName)

	return &telemetry{
		lp:             lp,
		meterProvider:  mp,
		tracerProvider: tp,
		logger:         logger,
		meter:          meter,
		tracer:         tracer,
		cfg:            cfg,
	}, nil
}

// MeterInt64UpDownCounter creates a new int64 up down counter metric.
func (t *telemetry) MeterInt64Counter(metric Metric) (otelmetric.Int64Counter, error) { //nolint:ireturn
	counter, err := t.meter.Int64Counter(
		metric.Name,
		otelmetric.WithDescription(metric.Description),
		otelmetric.WithUnit(metric.Unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter: %w", err)
	}

	return counter, nil
}

func (t *telemetry) Logger() *slog.Logger {
	return t.logger
}

func (t *telemetry) MeterProvider() otelmetric.MeterProvider {
	return t.meterProvider
}

func (t *telemetry) TracerProvider() *trace.TracerProvider {
	return t.tracerProvider
}

// Shutdown shuts down the logger, meter, and tracer.
func (t *telemetry) Shutdown(ctx context.Context) {
	t.lp.Shutdown(ctx)
	t.meterProvider.Shutdown(ctx)
	t.tracerProvider.Shutdown(ctx)
}

func SpanFromContext(ctx context.Context, spanName string) (context.Context, oteltrace.Span, *otelhttp.Labeler) {
	ctx, span := oteltrace.SpanFromContext(ctx).TracerProvider().Tracer("metrics").Start(ctx, spanName)
	labeler, _ := otelhttp.LabelerFromContext(ctx)
	return ctx, span, labeler
}

func ErrorOccurredAttribute() attribute.KeyValue {
	return attribute.Bool("error", true)
}

func ErrorTypeAttribute(errType string) attribute.KeyValue {
	return attribute.String("errorType", errType)
}
