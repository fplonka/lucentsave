FROM golang:1.24-bookworm AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY src/*.go src/
RUN cd src && CGO_ENABLED=0 go build -o /build/lucentsave .

FROM node:20-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev
COPY postSimplifyingServer.js ./

COPY --from=builder /build/lucentsave ./src/lucentsave
COPY src/templates/ ./src/templates/
COPY src/static/ ./src/static/

WORKDIR /app/src
EXPOSE 8080
CMD ["./lucentsave"]
