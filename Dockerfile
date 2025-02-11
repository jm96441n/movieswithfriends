## DEV BUILD
FROM golang:1.23.6-bookworm AS dev-base
RUN groupadd -g 1000 myuser && \
  useradd -u 1000 -g 1000 -m myuser
USER myuser
RUN go install github.com/air-verse/air@latest

FROM dev-base AS dev-code
USER myuser
WORKDIR /home/myuser/app
COPY ./go.mod ./go.sum ./
RUN go mod download

FROM dev-code AS dev
USER myuser
WORKDIR /home/myuser/app
CMD ["air"]

## PROD BUILD
FROM golang:1.23.6-bookworm AS builder
RUN groupadd -g 1000 myuser && \
  useradd -u 1000 -g 1000 -m myuser
USER myuser
WORKDIR /home/myuser/app
COPY ./.git ./.git
COPY ./go.mod ./go.sum ./
COPY ./cmd ./cmd
COPY ./tools ./tools
COPY ./metrics ./metrics
COPY ./web ./web
COPY ./identityaccess ./identityaccess
COPY ./partymgmt ./partymgmt
COPY ./ui ./ui
COPY ./migrations ./migrations
RUN go mod download && go run ./tools/assetbuilder/ && CGO_ENABLED=0 go build -o /go/bin/app ./cmd/movieswithfriends/ 

FROM gcr.io/distroless/static-debian12 AS prod
RUN groupadd -g 1000 myuser && \
  useradd -u 1000 -g 1000 -m myuser
USER myuser
COPY --from=builder /go/bin/app /
CMD ["/app"]
