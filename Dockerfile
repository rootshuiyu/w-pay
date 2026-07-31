# 多阶段构建 - Go构建阶段
FROM golang:1.22-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 设置 Go 代理
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖文件
COPY go.mod go.sum ./

# 复制源代码
COPY . .

# 构建应用（自动下载依赖）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -ldflags="-s -w" -o wpay main.go

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 wpay && \
    adduser -D -u 1000 -G wpay wpay

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/wpay .
COPY --from=builder /build/config ./config
COPY --from=builder /build/sql ./sql

# 创建必要的目录
RUN mkdir -p /app/certs && \
    chown -R wpay:wpay /app

# 切换到非root用户
USER wpay

# 暴露端口
EXPOSE 443 8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# 启动应用
CMD ["./wpay"]
