# Stage 1: Build
FROM golang:1.23.8-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/main ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/main .

EXPOSE 8080
CMD ["./main"]
