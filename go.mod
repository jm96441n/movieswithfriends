module github.com/jm96441n/movieswithfriends

go 1.23

toolchain go1.23.1

require (
	github.com/exaring/otelpgx v0.7.0
	github.com/gorilla/sessions v1.4.0
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/honeycombio/otel-config-go v1.17.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/playwright-community/playwright-go v0.4901.0
	github.com/pressly/goose/v3 v3.22.1
	github.com/samber/slog-multi v1.2.4
	github.com/testcontainers/testcontainers-go v0.34.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.34.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.9.0
	go.opentelemetry.io/otel/log v0.9.0
	go.opentelemetry.io/otel/sdk/log v0.9.0
	golang.org/x/crypto v0.31.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/containerd/containerd v1.7.18 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.7.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.24.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/sethvargo/go-envconfig v1.1.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/shirou/gopsutil/v4 v4.24.11 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.9.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/host v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.58.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.33.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.33.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.opentelemetry.io/proto/otlp v1.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
