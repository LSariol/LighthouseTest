# -- Build --
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o lighthousetest ./cmd/lighthousetest

# -- Final --
FROM alpine:latest
WORKDIR /app

RUN mkdir -p /app/lighthousetest

COPY --from=builder /app/lighthousetest /lighthousetest

CMD ["/lighthousetest"]