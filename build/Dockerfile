# Stage 1: Build Go tokka binary
FROM golang:1.24.2-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make all

# Stage 2: Final minimal image
FROM busybox:1.35.0-uclibc AS busybox
FROM gcr.io/distroless/base-debian12 AS runtime

COPY --from=busybox /bin/sh /bin/sh
COPY --from=busybox /bin/mkdir /bin/mkdir
COPY --from=busybox /bin/cat /bin/cat

WORKDIR /app

COPY --from=builder /app/.bin/tokka /bin/tokka
COPY --from=builder /app/build/plugins /tokka/plugins
COPY --from=builder /app/build/middlewares /tokka/middlewares

EXPOSE 8090 8095 9100 9464

CMD ["/bin/tokka"]
