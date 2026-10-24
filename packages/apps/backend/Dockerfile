FROM golang:1.27-alpine AS builder

WORKDIR /src/backend

COPY packages/apps/backend/go.mod packages/apps/backend/go.sum ./
RUN go mod download

COPY packages/apps/backend .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/rezible ./cmd/rezible

FROM alpine:3.22 AS rez_final

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/rezible /usr/local/bin/rezible

EXPOSE 7002

ENTRYPOINT ["rezible"]
CMD ["serve"]
