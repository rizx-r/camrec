FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/camrec ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/camrec /usr/local/bin/camrec
RUN mkdir -p /etc/camrec /var/lib/camrec
COPY config.docker.yaml /etc/camrec/config.yaml
ENV CAMREC_CONFIG=/etc/camrec/config.yaml
EXPOSE 8080
VOLUME ["/var/lib/camrec"]
ENTRYPOINT ["/usr/local/bin/camrec"]

