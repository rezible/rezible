ARG TARGETOS=linux
ARG TARGETARCH=amd64

FROM golang:1.25-alpine AS builder

WORKDIR /src/backend

RUN apk add --no-cache ca-certificates git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/rezible ./cmd/rezible
RUN GOBIN=/out go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

FROM alpine:3.22 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/rezible /usr/local/bin/rezible
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY backend/migrations /app/migrations

EXPOSE 7002

CMD ["rezible", "serve"]
