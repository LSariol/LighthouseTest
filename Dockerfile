# syntax=docker/dockerfile:1

############################
# Builder stage
############################
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Enable Go modules and caching
ENV CGO_ENABLED=0 \
    GO111MODULE=on

COPY go.mod ./
COPY VERSION ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go mod download || true

COPY cmd ./cmd
COPY internal ./internal

# Allow overriding version via build-arg, falling back to VERSION file
ARG VERSION_ARG
ENV APP_VERSION=${VERSION_ARG}

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

############################
# Runtime stage
############################
FROM alpine:3.20
WORKDIR /srv
RUN adduser -D -u 10001 app
COPY --from=builder /out/server /usr/local/bin/server
COPY VERSION /srv/VERSION

ENV PORT=2001 \
    APP_VERSION=""

USER app
EXPOSE 2001
ENTRYPOINT ["/usr/local/bin/server"]

