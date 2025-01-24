## DEV BUILD
FROM golang:1.23-bookworm AS dev-base

RUN go install github.com/air-verse/air@latest

FROM dev-base AS dev-code

WORKDIR /go/src/app

COPY ./go.mod ./go.sum ./

RUN go mod download

FROM dev-code AS dev

WORKDIR /go/src/app

CMD ["air"]

## PROD BUILD
FROM golang:1.23-bookworm AS builder

WORKDIR /go/src/app

COPY ./go.mod ./go.sum ./
COPY ./cmd ./cmd
COPY ./web ./web
COPY ./identityaccess/ ./identityaccess
COPY ./partymgmt/ ./partymgmt
COPY ./ui ./ui
COPY ./migrations ./migrations

RUN go mod download && CGO_ENABLED=0 go build -o /go/bin/app ./cmd/movieswithfriends/ 

FROM gcr.io/distroless/static-debian12 AS prod

COPY --from=builder /go/bin/app /
CMD ["/app"]
