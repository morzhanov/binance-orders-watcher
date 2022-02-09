FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o app /app/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /go/apps/binance-watcher
COPY --from=builder /app/app /go/apps/binance-watcher/app
COPY ./tls /go/apps/binance-watcher/tls
COPY ./internal/client/templates /go/apps/binance-watcher/internal/client/templates
COPY ./.env /go/apps/binance-watcher/.env
EXPOSE 443
CMD ["/go/apps/binance-watcher/app"]
