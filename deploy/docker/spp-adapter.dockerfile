FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/spp-adapter ./cmd/spp-adapter

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bin/spp-adapter .
COPY deploy/k8s/configs/spp-adapter-config.yaml ./config.yaml
COPY GeoIP2_City.mmdb /GeoIP2_City.mmdb

EXPOSE 8083
CMD ["./spp-adapter", "-config", "./config.yaml"]