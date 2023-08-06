FROM golang:1.20 as builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY main.go ./

RUN go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/bin/app ./

CMD ["./app"]
