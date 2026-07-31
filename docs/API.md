# 接口调试文档

> **N逼pay** · 连锁自营门店内部聚合订单调度系统 · API v1  
> 基础地址：`http://localhost:8090`（dev）/ `https://{备案域名}`（prod，强制 HTTPS）

---

## 5.1 全局统一返回格式

### 成功

```json
{
  "code": 200,
  "message": "success",
  "data": { }
}
```

### 失败

```json
{
  "code": 400,
  "message": "错误描述"
}
```

| code | 含义 |
|------|------|
| 200 | 成功 |
| 400 | 参数/业务错误 |
| 401 | 未登录或 Token 失效 |
| 403 | 权限不足 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

### 鉴权 Header

后台接口必须携带：

```
Authorization: Bearer {token}
```

或：

```
X-Token: {token}
```

---

## 5.2 三大类接口分组

### 一、收银端开放接口

> 门店内网收银机调用 · 简易限流 · **无需** admin token

#### POST `/api/pay/create` — 创建支付订单

**请求体：**

```json
{
  "amount": 100,
  "pay_type": 1,
  "pay_scene": "h5",
  "return_url": "https://cashier.example.com/done",
  "subject": "堂食消费"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | number | 是 | 金额（**分**），100 = 1.00 元 |
| pay_type | number | 是 | 1=微信，2=支付宝 |
| pay_scene | string | 否 | h5（默认）/ native |
| return_url | string | 建议 | H5/WAP 支付完成回跳 |
| store_id | number/string | 否 | 可选业务门店标识 |
| device_sn | string | 否 | 收银设备流水号 |
| subject | string | 否 | 业务备注 |

**响应 data：**

```json
{
  "order_id": "1983729184729183745",
  "pay_scene": "h5",
  "pay_url": "https://wx.tenpay.com/...",
  "qr_code_url": "",
  "amount": 100,
  "pay_type": 1
}
```

**curl 示例：**

```bash
curl -X POST http://localhost:8090/api/pay/create \
  -H "Content-Type: application/json" \
  -H "X-App-Key: pk_your_app_key" \
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://cashier.example.com/done","subject":"测试"}'
```

> 说明：收银端 API 支持通过 `X-App-Key` 请求头、`app_key` 查询参数或请求体中的 `app_key` 传递平台凭证。

#### GET `/api/pay/query?order_no={order_id}` — 查询订单

```bash
curl -H "X-App-Key: pk_your_app_key" "http://localhost:8090/api/pay/query?order_no=1983729184729183745"
# 或者使用 app_key 查询参数
curl "http://localhost:8090/api/pay/query?order_no=1983729184729183745&app_key=pk_your_app_key"
```

> 备注：`/api/pay/query` 支持 `X-App-Key` 请求头或 `app_key` 查询参数；若平台仅使用 IP 识别，调用方请确保请求来源在平台允许的白名单内。

**响应 data 字段：** `order_id`、`order_status`（0待支付 1已支付）、`total_amount`、`pay_amount`、`pay_type`、`pay_scene`、`qr_code_url`、`pay_time`、`transaction_id`

---

### 二、支付回调公开接口

> 微信/支付宝服务器调用 · **无登录** · 强制验签 + Redis 幂等

#### POST `/api/notify/wx` — 微信回调

- **notify_url 配置**：`https://{备案域名}/api/notify/wx?store_id={门店ID}`
- 验签失败返回 400
- 成功：`{"code":"SUCCESS","message":"成功"}`

#### POST `/api/notify/alipay` — 支付宝回调

- **notify_url 配置**：`https://{备案域名}/api/notify/alipay`
- 成功响应：`success`

> 生产环境请配置微信/支付宝官方 IP 白名单：环境变量 `CALLBACK_IP_WHITELIST`（应用层中间件）+ 防火墙/Nginx 双重限制。
> 经反向代理时需同时配置 `TRUSTED_PROXIES`，否则 `ClientIP` 取到的是代理地址。

---

### 三、后台鉴权接口

> 必须 Token · 分级权限：`super_admin` / `finance` / `operator`

#### 登录

```bash
curl -X POST http://localhost:8090/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

#### 门店管理（super_admin / operator）

##### POST `/api/admin/store/add` — 新增门店

```json
{
  "store_name": "人民路店",
  "store_code": "RM001",
  "address": "上海市XX区人民路100号",
  "tax_subject": "个体户：张三",
  "remark": ""
}
```

> 返回 `id` 即为 **store_id（雪花 ID）**，后续下单/配渠道均使用此 ID。

##### PUT `/api/admin/store/edit` — 修改门店信息

```json
{
  "id": 1234567890123456789,
  "store_name": "人民路旗舰店",
  "address": "新地址",
  "tax_subject": "个体户：张三"
}
```

##### GET `/api/admin/store/list` — 分页列表

| 参数 | 说明 |
|------|------|
| page | 页码，默认 1 |
| page_size | 每页条数，默认 20 |
| status | 1=正常，0=停用 |
| keyword | 搜索门店名/编号 |

```bash
curl "http://localhost:8090/api/admin/store/list?status=1&page=1" \
  -H "Authorization: Bearer {token}"
```

##### PUT `/api/admin/store/status` — 启用/停用

```json
{
  "id": 1234567890123456789,
  "status": 0
}
```

##### DELETE `/api/admin/store/del` — 逻辑删除

```json
{ "id": 1234567890123456789 }
```

或：`DELETE /api/admin/store/del?id=1234567890123456789`

---

#### 支付渠道（super_admin）

##### POST `/api/admin/channel/add` — 新增渠道

```json
{
  "store_id": 1234567890123456789,
  "pay_type": 1,
  "mch_no": "1234567890",
  "mch_key": "APIv3密钥",
  "app_id": "wxXXXXXXXX",
  "serial_no": "证书序列号",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "notify_url": "https://备案域名/api/notify/wx?store_id=1234567890123456789"
}
```

支付宝将 `pay_type` 设为 `2`，填写 `public_key`，`notify_url` 设为 `/api/notify/alipay`。

##### PUT `/api/admin/channel/edit` — 更换商户号/密钥（核心）

```json
{
  "id": 1,
  "mch_no": "新商户号",
  "mch_key": "新密钥",
  "private_key": "新私钥"
}
```

> 修改成功后 Redis 缓存自动失效，**下一笔订单立即使用新配置**，无需重启。

##### GET `/api/admin/channel/list?store_id=` — 查询门店渠道

```bash
curl "http://localhost:8090/api/admin/channel/list?store_id=1234567890123456789" \
  -H "Authorization: Bearer {token}"
```

##### PUT `/api/admin/channel/status` — 关停/启用通道

```json
{
  "id": 1,
  "status": 0
}
```

---

#### 订单查询（全部角色）

##### GET `/api/admin/order/list`

| 参数 | 说明 |
|------|------|
| store_ids | 逗号分隔多家门店，如 `1,2,3` |
| pay_type | 1/2 |
| order_status / status | 0待支付 1已支付 2已关闭 3退款 |
| start_time / end_time | YYYY-MM-DD |
| order_no | 订单号 |
| page / page_size | 分页 |

##### GET `/api/admin/order/detail?order_no=`

---

#### 对账统计（super_admin / finance）

##### GET `/api/admin/stat/summary` — 日/月汇总

| 参数 | 说明 |
|------|------|
| store_ids | 多门店逗号分隔 |
| start_time / end_time | 日期范围 |
| group_by | day（默认）/ month |

##### GET `/api/admin/export/orders` — 导出订单明细 Excel

##### GET `/api/admin/export/stat` — 导出汇总 Excel

```bash
curl "http://localhost:8090/api/admin/export/orders?store_ids=1,2&start_time=2026-01-01&end_time=2026-01-31" \
  -H "Authorization: Bearer {token}" -o orders.xlsx
```

---

## 6 核心业务流程说明

### 6.1 下单流程

```
收银机 POST /api/pay/create
    → 系统选择可用支付通道并调用微信/支付宝 SDK
    → 返回 pay_url（H5/WAP）或 qr_code_url（扫码）
    → 收银端跳转或展示二维码
    → GET /api/pay/query 查单确认
```

### 6.2 回调流程

```
微信/支付宝 POST /api/notify/*
    → 按订单 store_id 取【当前最新】渠道密钥验签
    → Redis 幂等 SetNX
    → 原子更新订单状态 + 写入脱敏 notify_data
```

**密钥更换注意**：完全替换商户后，旧商户延迟回调可能验签失败；建议旧商户保留约 **7 天** 再停用。

### 6.3 对账统计

- 支持勾选**多家门店**批量筛选
- 导出 Excel 供财务核对各个体户商户提现流水
- 系统**不参与**资金转账，仅输出对账报表

---

## 7 部署与安全规范

### 7.1 IP白名单分层说明

系统使用三层IP白名单机制，职责清晰分离：

1. **全局IP白名单**（通过环境变量或配置文件设置）
   - `ADMIN_IP_WHITELIST`: 后台管理接口，仅限内网/员工访问
   - `CALLBACK_IP_WHITELIST`: 支付回调接口，仅限微信/支付宝官方IP
   - `PAY_IP_WHITELIST`: 收银端API，可配置内网收银机IP范围
   - `TRUSTED_PROXIES`: 可信反向代理，防止客户端伪造X-Forwarded-For

2. **平台对接IP白名单**（在「代收平台」管理界面配置）
   - 每个代收平台独立配置`allowed_ips`字段
   - 用于第三方平台通过`app_key`识别身份和IP验证
   - 与全局白名单**并行生效**，必须同时满足
   - **重要**: 启用状态的平台必须配置非空IP白名单

3. **生产环境要求**
   - 全局白名单必须配置，空列表将拒绝所有访问
   - 平台启用前必须配置对接IP白名单
   - 空白名单=拒绝访问，必须显式配置可信IP

| 序号 | 要求 |
|------|------|
| 1 | 生产强制 HTTPS |
| 2 | 密钥/数据库环境变量注入，禁止硬编码 |
| 3 | `/api/admin/*` 限制内网 IP（`ADMIN_IP_WHITELIST` + 防火墙） |
| 4 | 配置三层IP白名单：
|   | - `ADMIN_IP_WHITELIST`: 后台管理接口（仅内网访问）
|   | - `CALLBACK_IP_WHITELIST`: 支付回调接口（仅微信/支付宝官方IP）
|   | - `PAY_IP_WHITELIST`: 收银端API（可配置内网收银机IP）
|   | - **平台对接IP白名单**: 在「代收平台」配置，用于第三方平台通过app_key识别 |
| 5 | 反向代理场景配置 `TRUSTED_PROXIES`，防止伪造 X-Forwarded-For |
| 6 | 渠道缓存 TTL 24h + 随机 jitter，防雪崩 |
| 7 | 敏感日志 6 个月自动清理 |

---

## 8 配套业务说明

系统仅输出每家个体户门店对账报表。财务核对商户提现流水后，按联营协议完成对公资金归集至总公司，**系统全程不参与资金转账**。

---

## 9 调试检查清单

- [ ] `GET /health` 返回 ok
- [ ] 登录获取 token
- [ ] 新增门店获得 store_id
- [ ] 配置微信/支付宝渠道
- [ ] `/api/pay/create` 返回 qr_code_url
- [ ] 模拟回调验签（需真实支付环境）
- [ ] 多门店批量导出 Excel

---

## 附录：Redis 缓存 Key

```
store:channel:{store_id}:{pay_type}   # pay_type: 1=微信 2=支付宝
wpay:admin:token:{admin_id}
wpay:callback:idempotent:{channel}:{transaction_id}
wpay:ratelimit:{ip}
```

缓存 TTL：默认 24 小时 + 0~4 小时随机 jitter（`config/*.yaml` 可配）。
