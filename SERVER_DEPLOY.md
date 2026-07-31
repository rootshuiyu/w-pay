# 服务器部署指南

## 服务器信息
- **IP**: 8.219.129.48
- **密码**: china1120@@
- **用户**: root

## 连接方式

由于 Windows 缺少 SSH 客户端，请使用以下工具之一：

### 方式一：使用 PuTTY
1. 下载 PuTTY: https://www.putty.org/
2. 打开 PuTTY，输入 Host Name: `8.219.129.48`
3. Port: `22`
4. 点击 Open，输入用户名 `root` 和密码 `china1120@@`

### 方式二：使用 Xshell
1. 下载 Xshell: https://www.xshell.com/zh/xshell/
2. 新建会话，输入主机 `8.219.129.48`
3. 端口 `22`，用户名 `root`
4. 连接并输入密码

### 方式三：安装 Windows OpenSSH（可选）
在 PowerShell（管理员）运行：
```powershell
Add-WindowsCapability -Online -Name OpenSSH.Client~~~~0.0.1.0
```
然后可以使用 `ssh root@8.219.129.48` 连接

## 服务器部署步骤

连接到服务器后，按以下步骤操作：

### 1. 更新系统
```bash
apt update && apt upgrade -y
```

### 2. 安装必要依赖
```bash
# 安装 PostgreSQL 16
apt install postgresql-16 postgresql-contrib-16 -y

# 安装 Redis
apt install redis-server -y

# 安装 Go 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装其他工具
apt install git nginx certbot python3-certbot-nginx -y
```

### 3. 配置 PostgreSQL
```bash
# 启动 PostgreSQL
systemctl start postgresql
systemctl enable postgresql

# 创建数据库用户和数据库
sudo -u postgres psql << EOF
CREATE ROLE wpay LOGIN PASSWORD 'wpay123';
CREATE DATABASE wpay OWNER wpay;
\q
EOF

# 初始化数据库表（需要先上传项目文件）
# 后续步骤执行
```

### 4. 配置 Redis
```bash
# 启动 Redis
systemctl start redis-server
systemctl enable redis-server

# 设置 Redis 密码（可选）
# 编辑 /etc/redis/redis.conf，设置 requirepass
```

### 5. 创建项目目录
```bash
mkdir -p /opt/wpay
mkdir -p /opt/wpay/config
mkdir -p /opt/wpay/certs
mkdir -p /opt/wpay/sql
```

### 6. 上传项目文件到服务器

使用 WinSCP 或 FileZilla 上传以下文件到 `/opt/wpay/`：
- `wpay` (编译后的可执行文件)
- `config/prod.yaml`
- `sql/init.sql`
- `web/dist/` (前端构建产物，如果需要)

### 7. 初始化数据库
```bash
cd /opt/wpay
sudo -u wpay psql -d wpay -f sql/init.sql
```

### 8. 设置环境变量
```bash
# 创建环境变量文件
cat > /opt/wpay/.env << EOF
APP_ENV=prod
PG_HOST=127.0.0.1
PG_PORT=5432
PG_USER=wpay
PG_PASSWORD=wpay123
PG_DATABASE=wpay
PG_SSLMODE=require
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_SECRET=请替换为随机64位字符串
TLS_CERT_FILE=/opt/wpay/certs/server.crt
TLS_KEY_FILE=/opt/wpay/certs/server.key
ADMIN_IP_WHITELIST=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
CALLBACK_IP_WHITELIST=
PAY_IP_WHITELIST=
TRUSTED_PROXIES=127.0.0.1,::1
EOF
```

**重要**: 请修改 `JWT_SECRET` 为随机强密钥，可以使用以下命令生成：
```bash
openssl rand -base64 48
```

### 9. 配置 SSL 证书
```bash
# 如果有域名，使用 Let's Encrypt
# certbot --nginx -d your-domain.com

# 或者使用自签名证书（测试用）
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /opt/wpay/certs/server.key \
  -out /opt/wpay/certs/server.crt
```

### 10. 创建 systemd 服务
```bash
cat > /etc/systemd/system/wpay.service << EOF
[Unit]
Description=N逼pay 聚合订单调度服务
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/wpay
EnvironmentFile=/opt/wpay/.env
ExecStart=/opt/wpay/wpay
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable wpay
systemctl start wpay
```

### 11. 检查服务状态
```bash
systemctl status wpay
journalctl -u wpay -f
```

### 12. 配置 Nginx（可选）
```bash
cat > /etc/nginx/sites-available/wpay << EOF
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass https://127.0.0.1:443;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }

    location /api/admin/ {
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
        proxy_pass https://127.0.0.1:443;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

ln -s /etc/nginx/sites-available/wpay /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

### 13. 开放防火墙端口
```bash
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

## 本地编译并上传

在本地 Windows 上执行：

```powershell
# 设置环境变量
$env:APP_ENV="prod"
$env:PG_HOST="8.219.129.48"
$env:PG_PASSWORD="wpay123"
$env:JWT_SECRET="your-secret-key"
$env:REDIS_ADDR="8.219.129.48:6379"

# 编译
go mod tidy
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o wpay main.go
```

然后使用 WinSCP 或 FileZilla 将 `wpay` 文件上传到服务器的 `/opt/wpay/` 目录。

## 验证部署

1. 检查服务状态：`systemctl status wpay`
2. 查看日志：`journalctl -u wpay -f`
3. 测试健康检查：`curl https://8.219.129.48/health`
4. 访问管理后台：`https://8.219.129.48`（默认账号：admin/admin123）

## 故障排除

- 服务无法启动：检查日志 `journalctl -u wpay -n 50`
- 数据库连接失败：检查 PostgreSQL 状态 `systemctl status postgresql`
- Redis 连接失败：检查 Redis 状态 `systemctl status redis-server`
- 端口被占用：检查 `netstat -tlnp | grep :443`
