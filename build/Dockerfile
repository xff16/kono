# syntax=docker/dockerfile:1

FROM golang:1.24.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kairyu ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/kairyu /usr/local/bin/kairyu

ENTRYPOINT ["/usr/local/bin/kairyu"]
