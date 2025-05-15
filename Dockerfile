FROM golang:1.24 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o maglev-proxy ./cmd/main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/maglev-proxy .

ENV LISTEN_PORT=8080
EXPOSE 8080

CMD ["./maglev-proxy"]

