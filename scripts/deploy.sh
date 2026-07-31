#!/bin/bash
# N逼pay 云端部署脚本

set -e

echo "=== N逼pay 云端部署脚本 ==="

# 检查环境变量
if [ -z "$PG_HOST" ] || [ -z "$PG_PASSWORD" ] || [ -z "$JWT_SECRET" ]; then
    echo "错误：请设置必要的环境变量"
    echo "必需变量：PG_HOST, PG_PASSWORD, JWT_SECRET"
    exit 1
fi

# 设置生产环境
export APP_ENV=prod

# 编译
echo "正在编译..."
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wpay main.go

echo "编译完成：wpay"

# 初始化数据库（如果需要）
if [ "$INIT_DB" = "true" ]; then
    echo "正在初始化数据库..."
    psql -h $PG_HOST -U postgres -c "CREATE ROLE wpay LOGIN PASSWORD '$PG_PASSWORD';" || true
    psql -h $PG_HOST -U postgres -c "CREATE DATABASE wpay OWNER wpay;" || true
    psql -h $PG_HOST -U wpay -d wpay -f sql/init.sql || true
fi

echo "部署完成！"
echo "请将以下文件上传到服务器："
echo "  - wpay (可执行文件)"
echo "  - config/prod.yaml"
echo "  - sql/init.sql (如需初始化数据库)"
echo ""
echo "服务器上需要设置的环境变量见 DEPLOY.md"
