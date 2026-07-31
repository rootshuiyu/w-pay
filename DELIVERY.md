# N逼pay — 连锁自营门店内部聚合订单调度系统 — 开发交付文档

## 1 项目概述

### 1.1 项目定位

本系统为**自营内部专用**订单调度平台，无对外收单业务、无资金清算能力。门店数量**无上限**，后台可自由新增、编辑、停用任意多家个体户门店；每家门店独立配置微信/支付宝商户，支持随时更换商户密钥与商户号，**配置热更新无需重启服务**。

系统用于：统一收银码生成、支付异步回调、订单管理、财务对账，配合个体户资金归集至总公司的联营经营模式。

### 1.2 合规核心约束（项目红线）

| 序号 | 约束 |
|------|------|
| 1 | 资金链路一清直达门店个体户商户，系统全程隔离资金，无二清/资金池 |
| 2 | 支付隐私数据严格脱敏，敏感日志 6 个月自动清理 |
| 3 | 仅实控人自有门店使用，永不对外开放商业化收单 |
| 4 | 禁止写死门店数量上限（不得出现固定 10 家等常量） |

### 1.3 技术栈

- 后端：Go 1.22 + Gin
- ORM：GORM v2
- 数据库：PostgreSQL 16+
- 缓存：Redis 7.0（门店渠道热更新）
- 支付 SDK：go-pay/gopay（Go 生态等价 IJPay-Go 封装）
- 定时任务：robfig/cron/v3
- 部署：Linux + HTTPS + ICP 备案域名

---

## 2 系统架构分层

| 包名 | 职责 |
|------|------|
| `config` | 全局配置，敏感项环境变量注入 |
| `common` | 统一返回、雪花 ID、JWT、脱敏、限流、Redis Key |
| `model` | admin / store / pay_channel / orders 表结构体 |
| `dao` | 数据库 CRUD + Redis 缓存读写 |
| `service` | 门店、渠道、下单、回调、对账业务逻辑 |
| `controller` | HTTP 接口控制器 |
| `task` | 超时关单、日志清理 cron |
| `router` | 后台鉴权 / 收银开放 / 支付回调路由分组 |

---

## 3 数据库表设计

### 3.1 admin 管理员账号表

| 字段 | 说明 |
|------|------|
| id | 主键 |
| username | 登录名 |
| password_hash | bcrypt 加密密码 |
| role | super_admin / finance / operator |
| phone | 手机号 |
| status | 1正常 0禁用 |
| created_at / updated_at / deleted_at | 审计与逻辑删除 |

### 3.2 store 门店信息表（无数量限制）

| 字段 | 说明 |
|------|------|
| id | **store_id，雪花主键** |
| store_name | 门店名称 |
| address | 经营地址 |
| tax_subject | 个体户主体全称 |
| status | 1正常 0停用 |

### 3.3 pay_channel 门店支付渠道表

| 字段 | 说明 |
|------|------|
| store_id | 关联门店 |
| pay_type | **1=微信 2=支付宝** |
| mch_no / mch_key | 商户号、密钥 |
| app_id / serial_no / private_key / public_key | 渠道参数 |
| notify_url | 回调地址 |
| status | 1启用 0关停 |

> 更换商户：直接编辑记录；停用渠道：status 置 0。

### 3.4 orders 订单主表

| 字段 | 说明 |
|------|------|
| order_no | 全局雪花 order_id |
| store_id / pay_type | 门店与渠道 |
| total_amount / pay_amount | 订单/实付金额（分） |
| order_status | 0待支付 1已支付 2已关闭 3退款 |
| device_sn / subject | 设备流水号、备注 |
| notify_data | **脱敏回调报文摘要**（非明文隐私） |
| pay_time | 支付完成时间 |

---

## 4 核心特性：门店 & 渠道热更新机制

```
系统启动 ──► 批量预热 pay_channel ──► Redis
                Key: store:channel:{store_id}:{pay_type}

后台修改渠道 ──► DB 更新成功 ──► DEL 对应 Redis Key
下次下单 ──► 缓存 miss ──► 查 DB ──► 回填 Redis ──► 立即使用新配置
```

**效果**：修改商户号、密钥、开关渠道后，新订单立刻生效，**服务无需重启**。

**扩容**：`pay_channel(store_id, pay_type)` 联合唯一索引，上千门店查询无瓶颈。

---

## 5 接口文档规范

### 5.1 全局统一返回格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| code | 含义 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未登录 |
| 403 | 无权限 |
| 429 | 限流 |
| 500 | 服务器错误 |

### 5.2 三大类接口分组

| 分组 | 前缀 | 鉴权 |
|------|------|------|
| 收银端开放 | `/api/pay` | 限流，无 token |
| 支付回调 | `/api/notify` | 验签 + 幂等 |
| 后台管理 | `/api/admin` | JWT Token + 分级权限 |

**收银端**
- `POST /api/pay/create` — 创建支付订单，返回付款码链接
- `GET /api/pay/query` — 查询订单

> 注意：收银 API 需携带 `X-App-Key`，也可使用 `app_key` 查询参数；若平台使用 IP 识别则需由白名单来源调用。

**支付回调**
- `POST /api/notify/wx?store_id=` — 微信回调
- `POST /api/notify/alipay` — 支付宝回调

**门店管理**（super_admin / operator）
- `POST /api/admin/store/add` — 新增门店
- `PUT /api/admin/store/edit` — 修改门店
- `GET /api/admin/store/list` — 分页列表（可筛 status）
- `PUT /api/admin/store/status` — 启用/停用
- `DELETE /api/admin/store/del` — 逻辑删除

**支付渠道**（super_admin）
- `POST /api/admin/channel/add` — 新增渠道
- `PUT /api/admin/channel/edit` — 更换商户号/密钥
- `GET /api/admin/channel/list?store_id=` — 门店渠道列表
- `PUT /api/admin/channel/status` — 关停/启用通道

**订单与对账**
- `GET /api/admin/order/list` — 订单查询（支持多 store_ids）
- `GET /api/admin/stat/summary` — 日/月汇总
- `GET /api/admin/export/orders` — 导出明细 Excel
- `GET /api/admin/export/stat` — 导出汇总 Excel

> 完整 curl 示例见 [docs/API.md](./docs/API.md)

### 5.4 下单请求示例

```json
{
  "store_id": 1234567890123456789,
  "amount": 100,
  "pay_type": 1,
  "device_sn": "POS-001",
  "subject": "堂食"
}
```

> `pay_type`：1=微信，2=支付宝；`amount` 单位：**分**

### 5.5 渠道配置示例

```json
{
  "store_id": 1234567890123456789,
  "pay_type": 1,
  "mch_no": "1234567890",
  "mch_key": "APIv3密钥",
  "app_id": "wxXXXX",
  "serial_no": "证书序列号",
  "private_key": "-----BEGIN PRIVATE KEY-----...",
  "notify_url": "https://备案域名/api/notify/wx?store_id=1234567890123456789"
}
```

---

## 6 核心业务流程变更点

1. **下单**：通过 `store_id` 动态匹配任意门店最新渠道；渠道修改后缓存自动失效，实时读取最新参数。
2. **回调**：按订单 `store_id` 查询当前最新密钥验签；完全替换商户后旧延迟回调可能失败，建议旧商户保留约 7 天再停用。
3. **对账**：支持勾选多家门店批量导出汇总报表，适配多个体户统一财务归集。

---

## 7 部署与安全规范

1. 生产环境强制 HTTPS
2. 密钥、数据库信息环境变量注入，禁止硬编码
3. 后台接口防火墙限制内网 IP 访问
4. 回调接口配置微信、支付宝官方 IP 白名单
5. 缓存 TTL + 随机 jitter 兜底，防止缓存雪崩
6. 敏感日志 6 个月自动清理

---

## 8 配套业务说明

系统仅输出每家个体户门店对账报表，财务核对商户提现流水后，按照联营协议完成个体户对公资金归集至总公司，**系统全程不参与资金转账操作**。

---

## 9 交付清单

| 序号 | 交付物 |
|------|--------|
| 1 | Go 完整分层源代码（无门店数量硬限制，渠道热更新） |
| 2 | PostgreSQL 建表 SQL（`sql/init.sql`） |
| 3 | README（环境搭建、编译、启动、缓存机制） |
| 4 | 接口调试文档（`docs/API.md`） |
| 5 | 部署说明（`DEPLOY.md`） |

---

## 10 定时任务

| Cron | 任务 |
|------|------|
| `*/5 * * * *` | 关闭超时 15 分钟未支付订单 |
| `0 3 * * *` | 清理 6 个月前 sensitive_logs |
| `0 * * * *` | 兜底预热渠道缓存 |

---

## 11 绝对禁止功能

1. 资金划转、提现、分账、代付、资金池
2. 商户入驻、第三方商家接入
3. 持久化 openid、授权码、私钥明文
4. 跨门店混用商户号
5. 门店数量硬编码上限

---

## 12 快速验证

```bash
go mod tidy
APP_ENV=dev go run main.go

# 健康检查
curl http://localhost:8080/health

# 登录
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

默认 dev 管理员：`admin` / `admin123`（生产务必修改）。
