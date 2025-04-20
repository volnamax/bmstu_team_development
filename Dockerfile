# Stage 1: build Go binary
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o app ./cmd/main.go

# Stage 2: run clean image
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
