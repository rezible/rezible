ARG BUILDER_BASE_IMAGE=golang:1.25-alpine
ARG BUILD_ENV=default

FROM ${BUILDER_BASE_IMAGE} AS builder_default
RUN apk add --no-cache ca-certificates

FROM ${BUILDER_BASE_IMAGE} AS builder_dev
RUN apk add --no-cache ca-certificates
COPY scripts/certs/localias-ca.crt /usr/local/share/ca-certificates/localias-ca.crt
RUN update-ca-certificates

FROM builder_${BUILD_ENV} AS builder

WORKDIR /src/backend

COPY packages/backend/go.mod packages/backend/go.sum ./
RUN go mod download

COPY packages/backend/ ./

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/rezible ./cmd/rezible
RUN GOBIN=/out go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

FROM alpine:3.22 AS rez_final

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/rezible /usr/local/bin/rezible
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY packages/backend/migrations /app/migrations

EXPOSE 7002

CMD ["rezible", "serve"]
