# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.23.4

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags='-s -w' -o /out/router ./cmd/router

FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app \
    && apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/router /usr/local/bin/router
USER app
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/router"]
