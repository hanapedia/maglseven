FROM golang:1.24 AS builder

WORKDIR /app
COPY hello.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o hello-server hello.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/hello-server .

ENV PORT=8080
EXPOSE 8080

CMD ["./hello-server"]
