FROM golang:1.21 as base

WORKDIR /go/src/app

COPY backend/go.mod ./backend/go.sum ./

RUN go mod download

COPY backend/main.go ./
COPY backend/web ./web
COPY backend/store ./store

FROM base as dev-base

RUN go install github.com/githubnemo/CompileDaemon@latest

FROM dev-base as dev

WORKDIR /go/src/app

CMD ["CompileDaemon", "-command=./movieswithfriends"]

FROM base as builder

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM debian:latest as prod

COPY --from=builder /go/bin/app /
CMD ["/app"]
