FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/http ./cmd/http
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/grpc ./cmd/grpc

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/http /app/bin/http
COPY --from=builder /app/bin/grpc /app/bin/grpc
COPY --from=builder /app/configs /app/configs

EXPOSE 8080 9090

CMD ["/app/bin/http"] 