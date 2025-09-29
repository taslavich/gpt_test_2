FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/router ./cmd/router

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bin/router .
COPY deploy/k8s/configs/router-config.yaml ./config.yaml
EXPOSE 8082
CMD ["./router", "-config", "./config.yaml"]