FROM node:18 AS builder

WORKDIR /app

COPY dashboard/static/package.json ./
COPY dashboard/static/tsconfig.json ./
COPY dashboard/static/app.ts ./

RUN npm install
RUN npm run build

FROM node:18-slim

WORKDIR /app

COPY --from=builder /app/dist ./dist

CMD ["node", "dist/app.js"]
