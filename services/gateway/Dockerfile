# 多阶段构建 Dockerfile for API Gateway

FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o api-gateway cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates 和 tzdata
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 创建非 root 用户
RUN adduser -D -s /bin/sh gateway

# 设置工作目录
WORKDIR /home/gateway

# 从构建阶段复制二进制文件
COPY --from=builder /app/api-gateway .

# 复制配置文件
COPY --from=builder /app/.env.example .env.example

# 修改文件所有者
RUN chown -R gateway:gateway /home/gateway

# 切换到非 root 用户
USER gateway

# 暴露端口
EXPOSE 8000 9000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8000/health || exit 1

# 运行应用
CMD ["./api-gateway"]