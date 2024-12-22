FROM golang:1.23 AS base


WORKDIR /go/src/app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./cmd ./
COPY ./web ./web
COPY ./identityaccess/ ./identityaccess
COPY ./partymgmt/ ./partymgmt
COPY ./ui ./ui
COPY ./store ./store
COPY ./migrations ./migrations

FROM base AS dev-base

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/air-verse/air@latest

FROM dev-base AS dev

WORKDIR /go/src/app

CMD ["air"]

FROM base AS builder

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM debian:latest AS prod

COPY --from=builder /go/bin/app /
CMD ["/app"]
