# N逼pay — 连锁自营门店内部聚合订单调度系统

> **接口调试**：[docs/API.md](./docs/API.md) · **逐项核对 & 自测**：[docs/VERIFY.md](./docs/VERIFY.md)  
> **交付文档**：[DELIVERY.md](./DELIVERY.md) · **部署说明**：[DEPLOY.md](./DEPLOY.md)

## 环境搭建

**依赖**：Go 1.22+ · PostgreSQL 16+ · Redis 7.0（Windows 下可用 Memurai 替代）

本地详细步骤见 **[docs/LOCAL_SETUP.md](./docs/LOCAL_SETUP.md)**

**方式一：Docker 一键**：

```powershell
docker compose up -d --wait
$env:APP_ENV="dev"; $env:PG_USER="wpay"; $env:PG_PASSWORD="wpay123"; $env:PG_DATABASE="wpay"
$env:REDIS_ADDR="127.0.0.1:6379"; $env:JWT_SECRET="local-dev-secret"
go run main.go
```

**方式二：本机 PostgreSQL**：

```powershell
# 1. 建库（psql 以超级用户执行）
psql -U postgres -c "CREATE ROLE wpay LOGIN PASSWORD 'wpay123';"
psql -U postgres -c "CREATE DATABASE wpay OWNER wpay;"
psql -U wpay -d wpay -f sql/init.sql

# 2. 环境变量
$env:APP_ENV="dev"
$env:PG_USER="wpay"; $env:PG_PASSWORD="wpay123"; $env:PG_DATABASE="wpay"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:JWT_SECRET="dev-secret"

# 3. 编译启动
go mod tidy
go run main.go
```

服务默认端口 **8090**（`SERVER_PORT` 可覆盖），健康检查：`GET http://localhost:8090/health`

## 管理后台前端（Vue 3）

`web/` 目录为 Vue 3 + Vite + Element Plus 管理后台，含登录、门店管理、支付渠道、订单查询、对账汇总（Excel 导出）。

```powershell
cd web
npm install
npm run build      # 产物输出 web/dist，Go 服务启动时自动托管
# 或开发模式（热更新，代理 /api 到 8090）：
npm run dev        # http://localhost:5173
```

构建后直接访问 **http://localhost:8090** 即为管理后台界面。

默认管理员：`admin` / `admin123`

## 编译

```bash
CGO_ENABLED=0 go build -o wpay.exe main.go
```

## 缓存机制说明

| 机制 | 说明 |
|------|------|
| Key 格式 | `store:channel:{store_id}:{pay_type}` |
| 预热 | 启动时批量写入所有启用渠道 |
| 热更新 | 后台改渠道 → `DEL` Key → 下单一律回源 DB |
| 防雪崩 | TTL 24h + 随机 jitter（见 `config/*.yaml`） |
| 兜底 | 每小时 cron 全量预热 |

## 接口路由一览（§5.2）

### 收银端（统一开放 API）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/pay/create` | 创建支付订单，返回跳转链接或二维码 |
| GET | `/api/pay/query` | 查单 |

对接方只需调用统一 API，按返回的 `pay_url` / `qr_code_url` 完成收款。每家对接平台在后台配置独立 **代收平台** 与绑定的商户码池，请求头携带 `X-App-Key` 即可（或由 IP 自动识别），平台之间互不串池。

示例：
```bash
curl -X POST http://localhost:8090/api/pay/create \
  -H "Content-Type: application/json" \
  -H "X-App-Key: pk_your_app_key" \
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://收银页/完成"}'

curl "http://localhost:8090/api/pay/create?app_key=pk_your_app_key" \
  -H "Content-Type: application/json" \
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://收银页/完成"}'

curl -H "X-App-Key: pk_your_app_key" "http://localhost:8090/api/pay/query?order_no=ORDER_ID"
# 或者使用 app_key 查询参数
curl "http://localhost:8090/api/pay/query?order_no=ORDER_ID&app_key=pk_your_app_key"

> 说明：收银端接口支持通过 `X-App-Key` 请求头或 `app_key` 参数传递平台凭证，平台若使用 IP 识别则需来自已配置的白名单地址。
```

**手机浏览器（默认）**：`pay_scene=h5` → 微信 H5 / 支付宝 WAP，`window.location.href = pay_url` 即可唤起；失败自动降级扫码。

```json
POST /api/pay/create
{ "amount": 100, "pay_type": 1, "pay_scene": "h5", "return_url": "https://收银页/完成" }
// PC 扫码： "pay_scene": "native"
// 可选 "store_id": "门店ID" 业务标识
```

> 生产环境请在 `config/prod.yaml` 配置 `pay.h5_app_url` 为收银站点域名。商户平台需开通 **微信 H5 支付**、**支付宝手机网站支付**。

### 支付回调

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/notify/wx?store_id=` | 微信回调 |
| POST | `/api/notify/alipay` | 支付宝回调 |

### 后台管理（需 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/store/add` | 新增门店 |
| PUT | `/api/admin/store/edit` | 编辑门店 |
| GET | `/api/admin/store/list` | 门店列表 |
| PUT | `/api/admin/store/status` | 启停门店 |
| DELETE | `/api/admin/store/del` | 删除门店 |
| GET | `/api/admin/platform/list` | 代收平台列表 |
| PUT | `/api/admin/platform/set-channels` | 绑定商户码到平台 |
| GET | `/api/admin/platform/pool` | 平台码池额度 |
| POST | `/api/admin/channel/add` | 新增渠道 |
| PUT | `/api/admin/channel/edit` | 更换商户配置 |
| GET | `/api/admin/channel/list` | 门店渠道列表 |
| PUT | `/api/admin/channel/status` | 关停通道 |
| GET | `/api/admin/order/list` | 订单查询 |
| GET | `/api/admin/stat/summary` | 对账汇总 |
| GET | `/api/admin/export/orders` | 导出明细 |
| GET | `/api/admin/export/stat` | 导出汇总 |

## 统一返回

```json
{ "code": 200, "message": "success", "data": {} }
```

## 合规要点

- 门店数量**无上限**，代码无固定上限常量
- **无**资金划转/提现/分账逻辑
- 支付隐私脱敏，不持久化 openid/授权码明文
- 生产 HTTPS（`tls_enabled`）+ 后台内网访问 + 回调 IP 白名单

IP 白名单可在管理后台 **基础配置 → IP 白名单** 可视化维护（立即生效）；亦可通过环境变量注入初始值：

```powershell
$env:ADMIN_IP_WHITELIST="10.0.0.0/8,192.168.0.0/16"   # 后台限内网（种子导入）
$env:PAY_IP_WHITELIST="203.0.113.10,192.168.1.0/24"   # 收银 API 限对接方出口 IP
$env:CALLBACK_IP_WHITELIST="203.0.113.5,198.51.100.0/24"  # 回调限支付平台官方 IP
$env:TRUSTED_PROXIES="127.0.0.1"                       # 同机 Nginx 反代时填写
```

`config/prod.yaml` 默认已限制后台为三段私有网段；启用策略但无合法条目时按 fail-closed 全部拒绝。`TRUSTED_PROXIES` 留空时不信任 `X-Forwarded-For`。

详细请求/响应示例见 **[docs/API.md](./docs/API.md)**。

## 自动化测试（E2E）

覆盖 [docs/VERIFY.md](./docs/VERIFY.md) 四个核心场景：

| 场景 | 测试文件 |
|------|----------|
| 1. 10+ 门店不串号 | `tests/e2e/scenario1_store_test.go` |
| 2. 渠道热更新 | `tests/e2e/scenario2_channel_test.go` |
| 3. 回调历史密钥 | `tests/e2e/scenario3_callback_test.go` |
| 4. 对账导出一致 | `tests/e2e/scenario4_reconcile_test.go` |

```powershell
# 一键运行（需 Docker + Go 1.22）
.\scripts\run_e2e.ps1
```

```powershell
# 手动运行（复用本机 PostgreSQL，无需 Docker）
$env:APP_ENV="test"; $env:WPAY_TEST_MODE="1"
$env:PG_HOST="127.0.0.1"; $env:PG_PORT="5432"
$env:PG_USER="wpay"; $env:PG_PASSWORD="wpay123"; $env:PG_DATABASE="wpay_test"
$env:REDIS_ADDR="127.0.0.1:6379"; $env:REDIS_DB="1"
go test -tags=e2e -v -count=1 ./tests/e2e/...
```

> 测试库 `wpay_test` 需预先创建：`psql -U postgres -c "CREATE DATABASE wpay_test OWNER wpay;"`
> 若 `PostgreSQL`/`Redis` 连不上，`TestMain` 会打印 `SKIP:` 并以 0 退出，不会误判为失败。

`WPAY_TEST_MODE=1` 跳过真实支付 SDK，mock 付款码 URL 含商户号便于断言不串号。
