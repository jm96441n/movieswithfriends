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

// MetricInvitationAcceptedCounter is a metric counting the number of successfully accepted invitations.
var MetricInvitationAcceptedCounter = Metric{
	Name:        "invite_successfully_accepted",
	Unit:        "{count}",
	Description: "Measures the number of successfully accepted invitations.",
}

var (
	userRegisterOnce      = NewRetryableOnce()
	userRegisteredCounter otelmetric.Int64Counter

	invitationAcceptedOnce    = NewRetryableOnce()
	invitationAcceptedCounter otelmetric.Int64Counter
)

// TODO: handle attributes
func (t *telemetry) IncreaseUserRegisteredCounter(ctx context.Context, logger *slog.Logger) {
	err := userRegisterOnce.Do(func() error {
		var err error
		userRegisteredCounter, err = t.MeterInt64Counter(MetricUserRegisteredCounter)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "Successfully created user registered counter")
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create user registered counter", slog.Any("err", err))
		return
	}

	logger.InfoContext(ctx, "Incrementing user signed up counter")
	userRegisteredCounter.Add(ctx, 1)
}

func (t *telemetry) IncreseInvitationAcceptedCounter(ctx context.Context, logger *slog.Logger) {
	err := invitationAcceptedOnce.Do(func() error {
		var err error
		invitationAcceptedCounter, err = t.MeterInt64Counter(MetricInvitationAcceptedCounter)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "Successfully created invitation accepted counter")
		return nil
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create invitation accepted counter", slog.Any("err", err))
		return
	}

	logger.InfoContext(ctx, "Incrementing invitation accepted counter")
	invitationAcceptedCounter.Add(ctx, 1)
}
