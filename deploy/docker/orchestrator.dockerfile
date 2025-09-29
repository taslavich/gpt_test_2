FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/orchestrator ./cmd/orchestrator

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bin/orchestrator .
COPY deploy/k8s/configs/orchestrator-config.yaml ./config.yaml
EXPOSE 8081
CMD ["./orchestrator", "-config", "./config.yaml"]