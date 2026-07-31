#!/usr/bin/env bash
# Wpay E2E 自动化测试 — 覆盖 VERIFY.md 四个核心场景
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> [1/4] 启动测试依赖 (PostgreSQL + Redis)"
docker compose -f docker-compose.test.yml up -d --wait

echo "==> [2/4] 初始化测试数据库"
export APP_ENV=test
export WPAY_TEST_MODE=1
export PG_HOST=127.0.0.1
export PG_PORT=5433
export PG_USER=postgres
export PG_PASSWORD=test123
export PG_DATABASE=wpay_test
export REDIS_ADDR=127.0.0.1:6380
export REDIS_DB=1
export JWT_SECRET=test-jwt-secret

if command -v psql &>/dev/null; then
  PGPASSWORD=test123 psql -h 127.0.0.1 -p 5433 -U postgres -d wpay_test -f sql/init.sql 2>/dev/null || true
fi

echo "==> [3/4] 下载依赖"
go mod tidy

echo "==> [4/4] 运行 E2E 测试 (build tag: e2e)"
go test -tags=e2e -v -count=1 -timeout=5m ./tests/e2e/...

echo ""
echo "=========================================="
echo "  E2E 全部场景测试通过"
echo "=========================================="
