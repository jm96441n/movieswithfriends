FROM golang:1.23-bookworm AS base

WORKDIR /go/src/app

COPY ./go.mod ./go.sum ./

RUN go mod download


## DEV BUILD
FROM base AS dev-base

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/air-verse/air@latest

FROM dev-base AS dev

WORKDIR /go/src/app

CMD ["air"]


## PROD BUILD
FROM base AS builder

COPY ./cmd ./cmd
COPY ./web ./web
COPY ./identityaccess/ ./identityaccess
COPY ./partymgmt/ ./partymgmt
COPY ./ui ./ui
COPY ./store ./store
COPY ./migrations ./migrations

RUN CGO_ENABLED=0 go build -o /go/bin/app ./cmd/movieswithfriends/ 

FROM gcr.io/distroless/static-debian12 AS prod

COPY --from=builder /go/bin/app /
CMD ["/app"]
