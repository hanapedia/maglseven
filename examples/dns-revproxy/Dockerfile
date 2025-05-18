FROM golang:1.24 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o maglseven-proxy ./examples/dns-revproxy/main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/maglseven-proxy .

ENV LISTEN_PORT=8080
EXPOSE 8080

CMD ["./maglseven-proxy"]

