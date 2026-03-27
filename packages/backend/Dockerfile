FROM golang:1.26-alpine AS builder

WORKDIR /src/backend

ARG MIGRATE_VERSION=4.19.1
RUN mkdir -p /out && \
    wget -qO- "https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    | tar -xz -C /out \
    && chmod +x /out/migrate

COPY packages/backend/go.mod packages/backend/go.sum ./
RUN go mod download

COPY packages/backend/ ./

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/rezible ./cmd/rezible

FROM alpine:3.22 AS rez_final

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/rezible /usr/local/bin/rezible
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY packages/backend/migrations /app/migrations

EXPOSE 7002

CMD ["rezible", "serve"]
