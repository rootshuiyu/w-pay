# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 wpay && \
    adduser -D -u 1000 -G wpay wpay

# 设置工作目录
WORKDIR /app

# 复制本地构建的二进制文件
COPY wpay .
COPY config ./config
COPY sql ./sql

# 添加执行权限
RUN chmod +x wpay

# 创建必要的目录
RUN mkdir -p /app/certs && \
    chown -R wpay:wpay /app

# 切换到非root用户
USER wpay

# 暴露端口
EXPOSE 8443 8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# 启动应用
CMD ["./wpay"]
