# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /app

CMD [ "pwd" ]
CMD [ "ls -l" ]

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kairyu ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/kairyu /usr/local/bin/kairyu

ENTRYPOINT ["/usr/local/bin/kairyu"]
