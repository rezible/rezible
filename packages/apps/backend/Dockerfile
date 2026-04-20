FROM golang:1.26-alpine AS builder

WORKDIR /src/backend

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/rezible ./cmd/rezible

FROM alpine:3.22 AS rez_final

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/rezible /usr/local/bin/rezible
COPY --from=builder /src/backend/migrations /app/migrations

EXPOSE 7002

ENTRYPOINT ["rezible"]
CMD ["serve"]
