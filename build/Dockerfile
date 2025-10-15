# syntax=docker/dockerfile:1.6

# Stage 1: Build TypeScript admin
FROM node:20-alpine AS admin-builder

WORKDIR /admin

COPY admin/static/package.json admin/static/tsconfig.json ./

RUN npm install

COPY admin/static/*.ts ./
COPY admin/static/*.css ./
COPY admin/static/index.html ./

RUN npx tsc

# Stage 2: Build Go Kairyu binary
FROM golang:1.24.2-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make all

# Stage 3: Final minimal image
FROM gcr.io/distroless/base-debian12 AS runtime

WORKDIR /app

COPY --from=builder /app/build/kairyu /bin/kairyu
COPY --from=builder /app/build/plugins /plugins

COPY --from=admin-builder /admin/dist ./admin/static/dist
COPY --from=admin-builder /admin/index.html ./admin/static/index.html
COPY --from=admin-builder /admin/style.css ./admin/static/style.css

EXPOSE 8090 8095

CMD ["/bin/kairyu"]
