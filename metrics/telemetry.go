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

// Telemetry is a wrapper around the OpenTelemetry logger, meter, and tracer.
type Telemetry struct {
	lp             *log.LoggerProvider
	MeterProvider  *metric.MeterProvider
	TracerProvider *trace.TracerProvider
	Logger         *slog.Logger
	meter          otelmetric.Meter
	tracer         oteltrace.Tracer
	cfg            Config
}

// NewTelemetry creates a new telemetry instance.
func NewTelemetry(ctx context.Context, cfg Config) (*Telemetry, error) {
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

	return &Telemetry{
		lp:             lp,
		MeterProvider:  mp,
		TracerProvider: tp,
		Logger:         logger,
		meter:          meter,
		tracer:         tracer,
		cfg:            cfg,
	}, nil
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

// MeterInt64UpDownCounter creates a new int64 up down counter metric.
func (t *Telemetry) MeterInt64Counter(metric Metric) (otelmetric.Int64Counter, error) { //nolint:ireturn
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

// Shutdown shuts down the logger, meter, and tracer.
func (t *Telemetry) Shutdown(ctx context.Context) {
	t.lp.Shutdown(ctx)
	t.MeterProvider.Shutdown(ctx)
	t.TracerProvider.Shutdown(ctx)
}
