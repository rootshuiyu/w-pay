# 接口逐项核对 & 自测验证清单

本文档对照业务需求逐项核对代码实现，并提供上线前必测的 4 个关键场景。

---

## 一、接口逐项核对（全部已实现）

### 1. 门店管理（无数量上限）

| 需求 | 路径 | 实现位置 | 状态 |
|------|------|----------|------|
| 新增门店，无硬编码上限 | `POST /api/admin/store/add` | `controller/store.go` → `service/store.go`（雪花 store_id） | ✅ |
| 修改个体户主体、地址 | `PUT /api/admin/store/edit` | Body: `id`, `tax_subject`, `address` | ✅ |
| 分页列表，筛选启停 | `GET /api/admin/store/list` | Query: `page`, `page_size`, `status`, `keyword` | ✅ |
| 一键启停门店 | `PUT /api/admin/store/status` | Body: `id`, `status`（1/0） | ✅ |
| 逻辑删除，财务备查 | `DELETE /api/admin/store/del` | GORM `DeletedAt` 软删除 | ✅ |

**合规**：代码中无 `MAX_STORE=10` 等常量；`store.go` 注释明确无数量限制。

---

### 2. 支付渠道（随时更换商户，热更新）

| 需求 | 路径 | 实现位置 | 状态 |
|------|------|----------|------|
| 新增微信/支付宝通道 | `POST /api/admin/channel/add` | `pay_type`: 1/2 | ✅ |
| 更换商户号/密钥/证书 | `PUT /api/admin/channel/edit` | 变更前归档 `pay_channel_history` + `DEL` Redis | ✅ |
| 查询门店全部渠道 | `GET /api/admin/channel/list?store_id=` | 返回脱敏后的渠道列表 | ✅ |
| 单独关停某店某通道 | `PUT /api/admin/channel/status` | Body: `id`, `status`（0关停） | ✅ |

**热更新机制**：

```
PUT /api/admin/channel/edit
  → 归档旧密钥到 pay_channel_history（7天）
  → DELETE Redis store:channel:{store_id}:{pay_type}
  → 下一笔 POST /api/pay/create 回源 DB 使用新商户
```

---

### 3. 订单 & 财务对账

| 需求 | 路径 | 实现位置 | 状态 |
|------|------|----------|------|
| 多条件筛选订单 | `GET /api/admin/order/list` | `store_ids`, `pay_type`, `status`, 日期 | ✅ |
| 多门店汇总统计 | `GET /api/admin/stat/summary` | `store_ids` + 日/月分组 | ✅ |
| 导出订单明细 | `GET /api/admin/export/orders` | Excel，匹配提现账单 | ✅ |
| 导出多店汇总 | `GET /api/admin/export/stat` | Excel，联营归集做账 | ✅ |

**收银端（补充）**：

| 路径 | 说明 |
|------|------|
| `POST /api/pay/create` | 创建订单，返回 `qr_code_url` |
| `GET /api/pay/query?order_no=` | 查单 |

**回调（补充）**：

| 路径 | 说明 |
|------|------|
| `POST /api/notify/wx?store_id=` | 微信验签 + 历史密钥兼容 |
| `POST /api/notify/alipay` | 支付宝验签 + 历史密钥兼容 |

---

### 4. 统一返回结构

```json
{"code":200,"message":"success","data":{}}
```

| code | 场景 | 实现 |
|------|------|------|
| 200 | 成功 | `common.CodeSuccess` |
| 401 | 未登录 | `middleware.AuthToken` |
| 403 | 无权限 | `middleware.RequireRole` |
| 400/429/500 | 业务/限流/异常 | `common/response.go` |

---

## 二、合规要点落地核对

| 要点 | 落地方式 | 状态 |
|------|----------|------|
| 门店无上限 | 无固定门店数常量；雪花 ID 无限扩容 | ✅ |
| 剥离资金逻辑 | 无转账/提现/分账/代付接口与代码 | ✅ |
| 隐私合规 | `desensitize.go`；订单仅存脱敏 `notify_data`；不存 openid/授权码 | ✅ |
| HTTPS | `config/prod.yaml` `tls_enabled: true` | ✅ |
| 后台内网 | `middleware.IPWhitelist("admin")` + `ADMIN_IP_WHITELIST`；另见 `DEPLOY.md` Nginx 示例 | ✅ 代码+部署 |
| 回调 IP 白名单 | `middleware.IPWhitelist("callback")` + `CALLBACK_IP_WHITELIST` | ✅ 代码+运维 |
| 可信反代 | `TRUSTED_PROXIES` → Gin `SetTrustedProxies`（防伪造 XFF） | ✅ 代码 |
| 缓存防雪崩 | TTL 24h + jitter | ✅ |
| 敏感日志 6 月清理 | cron `0 3 * * *` | ✅ |
| 渠道历史 7 天清理 | cron `0 4 * * *` | ✅ |

---

## 三、开发执行顺序（已完成）

1. ✅ 4 张核心表 SQL：`admin` / `store` / `pay_channel` / `orders` + `pay_channel_history`
2. ✅ Golang 分层脚手架
3. ✅ Redis 缓存热更新
4. ✅ 下单 + 回调验签 + 幂等
5. ✅ 后台管理 + 统计 + Excel 导出
6. ✅ 定时任务：关单 + 日志清理 + 缓存兜底 + 历史密钥清理

---

## 四、自测验证重点（必测 4 场景）

### 场景 1：扩容测试 — 10+ 门店不串号

```text
步骤：
1. 循环 POST /api/admin/store/add 创建 ≥10 家门店
2. 每家 POST /api/admin/channel/add 配置不同 mch_no
3. 分别 POST /api/pay/create 下单

预期：
- 各门店 order 的 store_id、channel_id 正确
- 微信/支付宝预下单使用各自 mch_no（查日志 mch=）
- 绝不允许 A 店订单走 B 店商户号
```

### 场景 2：更换商户热更新

```text
步骤：
1. 门店 A 配置渠道 mch_no=OLD，下单得 order_1
2. PUT /api/admin/channel/edit 改为 mch_no=NEW
3. 立即 POST /api/pay/create 得 order_2（不重启服务）

预期：
- order_2 使用 NEW 商户（日志/商户平台可见）
- 无需重启进程
```

### 场景 3：回调兼容旧密钥

```text
步骤：
1. 旧商户配置下创建 order_1（待支付）
2. PUT /api/admin/channel/edit 更换密钥（系统自动归档旧密钥）
3. 模拟/真实触发 order_1 的旧商户回调

预期：
- 系统依次尝试当前密钥 + pay_channel_history 中旧密钥
- order_1 验签通过并更新为已支付
- 7 天后历史密钥自动清理
```

### 场景 4：对账导出一致性

```text
步骤：
1. 选 store_ids=1,2,3，日期范围含已支付订单
2. GET /api/admin/stat/summary?store_ids=1,2,3&...
3. GET /api/admin/export/stat 与 /api/admin/export/orders

预期：
- summary 已支付金额 = export/orders 中 pay_amount 合计
- export/stat 各店汇总与 summary 一致
- Excel 可直接交财务做联营归集
```

---

## 五、后续可扩展（架构已预留，无需重构）

| 扩展项 | 建议路径 | 说明 |
|--------|----------|------|
| 门店操作员子账号 | `PUT /api/admin/user/bind-store` | 已有 `operator` 角色，可加 store 绑定 |
| 退款状态查询 | `GET /api/admin/order/refund-status` | 只读查询，不代付 |
| 小票打印 | `POST /api/pay/print` | 内网调用 |
| 异常告警 | webhook / 消息队列 | 大额单、回调失败计数 |

---

## 六、快速冒烟命令

> 服务默认端口 **8090**（8080 常被 PostgreSQL 自带 PEM 界面占用），可用 `SERVER_PORT` 覆盖。

```bash
# 健康
curl http://localhost:8090/health

# 登录
TOKEN=$(curl -s -X POST http://localhost:8090/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 新增门店
curl -X POST http://localhost:8090/api/admin/store/add \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"store_name":"测试店","tax_subject":"个体户测试"}'

# 门店列表
curl "http://localhost:8090/api/admin/store/list?status=1" \
  -H "Authorization: Bearer $TOKEN"
```

详细参数见 [API.md](./API.md)。

---

## 七、自动化测试

```powershell
# 一键 E2E：Docker 可用时起 PostgreSQL:5433 + Redis:6380，否则自动回退本机 5432/6379
.\scripts\run_e2e.ps1
```

或手动指定依赖：

```powershell
$env:APP_ENV="test"; $env:WPAY_TEST_MODE="1"
$env:PG_USER="wpay"; $env:PG_PASSWORD="wpay123"; $env:PG_DATABASE="wpay_test"
$env:REDIS_ADDR="127.0.0.1:6379"; $env:REDIS_DB="1"
go test -tags=e2e -v -count=1 ./tests/e2e/...
```

环境变量 `WPAY_TEST_MODE=1` 启用 mock 支付，无需真实微信/支付宝密钥。

当前 12 个用例（4 场景 + 冒烟 + 路径核对）全部通过。若 PostgreSQL/Redis 连不上，`TestMain` 打印 `SKIP:` 并以 0 退出。
