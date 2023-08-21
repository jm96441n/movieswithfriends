FROM golang:1.21 as builder

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./
COPY ./web ./web
COPY ./store ./store
COPY ./templates ./templates

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM debian:latest

COPY --from=builder /go/bin/app /
CMD ["/app"]
