package metrics

// NoopTelemetry is a no-op implementation of the TelemetryProvider interface.
type NoopTelemetry struct {
	serviceName string
}

// NewNoopTelemetry creates a new NoopTelemetry instance.
func NewNoopTelemetry(cfg Config) (*NoopTelemetry, error) {
	return &NoopTelemetry{serviceName: cfg.ServiceName}, nil
}
