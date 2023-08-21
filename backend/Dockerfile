FROM golang:1.21 as builder

WORKDIR /go/src/app

COPY backend/go.mod ./backend/go.sum ./

RUN go mod download

COPY backend/main.go ./
COPY backend/web ./web
COPY backend/store ./store
COPY backend/templates ./templates

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM debian:latest

COPY --from=builder /go/bin/app /
CMD ["/app"]
