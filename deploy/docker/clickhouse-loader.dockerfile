FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/clickhouse-loader ./cmd/clickhouse-loader

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bin/clickhouse-loader .
COPY deploy/k8s/configs/clickhouse-loader-config.yaml ./config.yaml
EXPOSE 8080
CMD ["./clickhouse-loader", "-config", "./config.yaml"]