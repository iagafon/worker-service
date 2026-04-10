FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY . .
RUN go build -o worker-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/worker-service .

ENTRYPOINT ["./worker-service"]