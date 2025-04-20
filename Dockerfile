FROM golang:1.23.8

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o app ./cmd/main.go

EXPOSE 8080
CMD ["./app"]
