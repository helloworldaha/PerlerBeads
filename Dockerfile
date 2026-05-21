FROM golang:1.21-bookworm AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM node:20-bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server .

RUN mkdir -p uploads output static

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD ["./server"]
