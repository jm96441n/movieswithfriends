package metrics

import (
	"context"
	"log/slog"

	otelmetric "go.opentelemetry.io/otel/metric"
)

// MetricUserRegisteredCounter is a metric counting the number of successfully registered users.
var MetricUserRegisteredCounter = Metric{
	Name:        "user_successfully_registered",
	Unit:        "{count}",
	Description: "Measures the number of successfully registered users.",
}

var (
	userRegisterOnce      = NewRetryableOnce()
	userRegisteredCounter otelmetric.Int64Counter
)

// TODO: handle attributes
func (t *Telemetry) IncreaseUserRegisteredCounter(ctx context.Context, logger *slog.Logger) {
	err := userRegisterOnce.Do(func() error {
		var err error
		userRegisteredCounter, err = t.MeterInt64Counter(MetricUserRegisteredCounter)
		if err != nil {
			return err
		}
		logger.Info("Successfully created user registered counter")
		return nil
	})
	if err != nil {
		logger.Error("failed to create user registered counter", slog.Any("err", err))
		return
	}

	logger.Info("Incrementing user signed up counter")
	userRegisteredCounter.Add(ctx, 1)
}
