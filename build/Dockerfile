# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /src

COPY . .
RUN go build -o /bin/kairyu ./cmd/main

FROM alpine:3.20
COPY --from=builder /kairyu /usr/local/bin/kairyu
ENTRYPOINT ["/usr/local/bin/kairyu"]
