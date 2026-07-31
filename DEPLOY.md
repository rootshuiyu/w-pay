# N逼pay 部署说明

## 一、部署架构

```
                    ┌─────────────┐
  收银机（内网） ──► │  N逼pay API  │ ◄── 后台管理（内网/VPN）
                    │  Gin + Go   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
          PostgreSQL    Redis      微信/支付宝
                                    （回调 HTTPS）
```

## 二、服务器要求

- Linux（推荐 Ubuntu 22.04 / CentOS 7+）
- 2 核 4G 内存起（上千门店建议 4 核 8G）
- 已备案域名 + SSL 证书（生产强制 HTTPS）
- 开放端口：443（API），5432/6379 仅内网

## 三、生产环境变量

```bash
export APP_ENV=prod
export PG_HOST=10.0.0.10
export PG_PORT=5432
export PG_USER=wpay
export PG_PASSWORD=<强密码>
export PG_DATABASE=wpay
export PG_SSLMODE=require
export REDIS_ADDR=10.0.0.11:6379
export REDIS_PASSWORD=<redis密码>
export JWT_SECRET=<随机64位字符串>
export TLS_CERT_FILE=/etc/wpay/certs/server.crt
export TLS_KEY_FILE=/etc/wpay/certs/server.key

# 访问来源白名单（逗号分隔 IP 或 CIDR）
export ADMIN_IP_WHITELIST=10.0.0.0/8,192.168.0.0/16
export CALLBACK_IP_WHITELIST=<微信/支付宝官方回调 IP>
```

## 四、部署步骤

### 1. 编译

```bash
cd wpay
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wpay main.go
```

### 2. 初始化数据库

```bash
psql -h $PG_HOST -U postgres -c "CREATE ROLE wpay LOGIN PASSWORD '强密码';"
psql -h $PG_HOST -U postgres -c "CREATE DATABASE wpay OWNER wpay;"
psql -h $PG_HOST -U wpay -d wpay -f sql/init.sql
```

### 3. 创建初始管理员

生产环境不要依赖 dev 自动初始化，手动插入 bcrypt 哈希后的密码。

### 4. systemd 服务

```ini
# /etc/systemd/system/wpay.service
[Unit]
Description=N逼pay 聚合订单调度服务
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=wpay
WorkingDirectory=/opt/wpay
EnvironmentFile=/opt/wpay/.env
ExecStart=/opt/wpay/wpay
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable wpay
sudo systemctl start wpay
```

### 5. Nginx 反向代理（可选）

后台管理接口建议限制内网 IP，收银端和支付回调走 HTTPS 公网。

```nginx
location /api/admin/ {
    allow 10.0.0.0/8;
    allow 192.168.0.0/16;
    deny  all;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_pass http://127.0.0.1:8090;
}
```

应用层同样内置 IP 白名单（`ADMIN_IP_WHITELIST` / `CALLBACK_IP_WHITELIST`），与 Nginx 形成双重防护。
经反向代理时需配置 `TRUSTED_PROXIES`（如 `127.0.0.1`），应用会调用 Gin `SetTrustedProxies`，否则 `ClientIP` 取到的都是代理地址；留空则不信任任何 `X-Forwarded-For`。

## 五、支付渠道配置要点

### 微信

1. 商户平台开通 Native 支付
2. 配置 APIv3 密钥、下载商户证书
3. 回调 URL：`https://{备案域名}/api/notify/wx?store_id={store_id}`

### 支付宝

1. 开通当面付（precreate）
2. 回调 URL：`https://{备案域名}/api/notify/alipay`

## 六、安全 checklist

- 生产 JWT_SECRET 使用随机强密钥
- PostgreSQL/Redis 不对公网开放
- `/api/admin` 限制内网或 VPN 访问
- 修改默认 admin 密码
- 启用 HTTPS
- 本系统无资金划转、提现、分账代码

## 七、扩容建议

- 订单表按 store_id 或 created_at 月份水平分表
- 门店数超过 5000 时考虑 Redis Cluster
- 对账查询走 PostgreSQL 只读副本
