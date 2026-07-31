# N逼pay 本地运行指南（Windows）

> **说明**：本项目数据库为 **PostgreSQL**，缓存为 Redis（Windows 本机可用 **Memurai** 替代，协议完全兼容）。

## 当前机器已就绪的环境

| 组件 | 状态 | 说明 |
|------|------|------|
| Go 1.26 | 已安装 | `go version` 验证 |
| PostgreSQL 18 | 已安装并运行 | 服务 `postgresql-x64-18`，端口 **5432** |
| PostgreSQL 16 | 已安装并运行 | 服务 `postgresql-x64-16`，端口 5433（备用） |
| Memurai（Redis 兼容） | 已安装并运行 | 服务 `Memurai`，端口 **6379** |
| 数据库 `wpay` / `wpay_test` | 已建好 | 账号 `wpay` / `wpay123`，6 张表已初始化 |

## 启动 N逼pay 服务

```powershell
cd c:\Users\Administrator\Desktop\wei

$env:APP_ENV="dev"
$env:PG_HOST="127.0.0.1"
$env:PG_PORT="5432"
$env:PG_USER="wpay"
$env:PG_PASSWORD="wpay123"
$env:PG_DATABASE="wpay"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:JWT_SECRET="local-dev-secret"

.\wpay.exe        # 或 go run main.go
```

访问：http://localhost:8090/health （8080 被 PostgreSQL 自带 PEM 管理界面占用，故默认改为 **8090**，可用 `SERVER_PORT` 覆盖）

默认管理员：`admin` / `admin123`

## Docker 方式（可选）

若 Docker Desktop 可用（需 WSL2，安装后重启一次）：

```powershell
docker compose up -d --wait
```

将启动 PostgreSQL 16（5432，账号 wpay/wpay123）+ Redis 7（6379），首次自动执行 `sql/init.sql` 建表。

> 注意：本机 5432 已被原生 PostgreSQL 18 占用，二选一即可；同时使用时需修改 compose 端口映射。

## 运行自动化测试

```powershell
# Docker 可用时用 docker-compose.test.yml（5433/6380）；
# Docker 不可用时自动回退本机 PostgreSQL:5432 + Redis:6379 的 wpay_test 库
.\scripts\run_e2e.ps1
```

## 常用命令

```powershell
# PostgreSQL 命令行（18 版）
$env:PGPASSWORD="wpay123"
& "D:\Program Files\PostgreSQL\18\bin\psql.exe" -U wpay -h 127.0.0.1 -d wpay

# 查看表
#   \dt

# Redis（Memurai）命令行
& "C:\Program Files\Memurai\memurai-cli.exe" ping

# 服务管理
Get-Service postgresql-x64-18, Memurai
```

## 故障排查

### 端口冲突

- **8080**：被 PostgreSQL 自带 PEM httpd 占用 → N逼pay 已默认用 8090
- **5432**：原生 PostgreSQL 18；docker compose 里的 postgres 也映射 5432，两者不能同时跑

### Docker daemon 未运行

Docker Desktop 依赖 WSL2。若 `docker info` 报 500 错误或找不到管道：

1. 以管理员执行 `wsl --install --no-distribution`
2. **重启电脑**，启动 Docker Desktop 等待 Running
3. 本地开发不依赖 Docker，可直接用上面的原生 PostgreSQL + Memurai
