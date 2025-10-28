# Stage 1: Build Go bravka binary
FROM golang:1.24.2-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make all

# Stage 2: Final minimal image
FROM gcr.io/distroless/base-debian12 AS runtime

WORKDIR /app

COPY --from=builder /app/build/bravka /bin/bravka
COPY --from=builder /app/build/plugins /bravka/plugins
COPY --from=builder /app/build/middlewares /bravka/middlewares

EXPOSE 8090 8095

CMD ["/bin/bravka"]
