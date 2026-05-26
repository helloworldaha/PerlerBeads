FROM node:20-bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
RUN git init && \
    git config user.name "Docker Build" && \
    git config user.email "docker@example.com" && \
    git commit --allow-empty -m "Initial commit"

COPY . /app/

ENV GOPROXY=https://goproxy.cn,direct

RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go \
    && rm -rf /var/lib/apt/lists/* && \
    cd /app && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server && \
    mkdir -p uploads output static && \
    apt-get remove -y --auto-remove golang-go

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD ["./server"]
