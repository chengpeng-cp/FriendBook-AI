# ============================================================
# FriendBook AI 多阶段构建 Dockerfile
# 阶段1 builder：Go 编译环境，产出静态二进制
# 阶段2 runtime：极简 alpine 运行环境（含时区数据）
# ============================================================

# === 阶段 1：构建 ===
FROM golang:1.23-alpine AS builder
WORKDIR /app

# 先拷贝依赖清单再下载（利用 Docker 层缓存：代码改动不触发依赖重下）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译（CGO_ENABLED=0 产出静态二进制，运行环境无需 glibc）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

# === 阶段 2：运行 ===
FROM alpine:3.19
WORKDIR /app

# 时区（Asia/Shanghai）+ CA 证书（HTTPS 外呼 OpenAI 需要）
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 只拷贝必要产物：二进制 + 配置
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./server"]
