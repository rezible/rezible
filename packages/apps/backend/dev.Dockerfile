FROM golang:1.26-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@v1.65.1

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]