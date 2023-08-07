FROM golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./
COPY ./web ./web
COPY ./store ./store
COPY ./templates ./templates

RUN go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/bin/app ./

CMD ["./app"]
