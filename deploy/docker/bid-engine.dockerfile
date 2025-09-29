FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bid-engine ./cmd/bid-engine

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bin/bid-engine .
COPY deploy/k8s/configs/bid-engine-config.yaml ./config.yaml
EXPOSE 8080
CMD ["./bid-engine", "-config", "./config.yaml"]