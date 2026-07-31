# GitHub 上传和 Docker 镜像构建指南

## 已创建的文件

- ✅ `Dockerfile` - Docker 构建配置
- ✅ `.dockerignore` - Docker 构建排除文件
- ✅ `wpay` - 预构建的 Go 二进制文件

## 上传到 GitHub 的步骤

### 方式一：使用 GitHub Desktop（推荐）

1. 下载 GitHub Desktop: https://desktop.github.com/
2. 安装并登录 GitHub 账号
3. 点击 "File" → "Add local repository"
4. 选择 `c:\Users\Administrator\Desktop\wei` 目录
5. 点击 "Publish repository" 发布到 GitHub
6. 设置仓库名称（如：w-pay）

### 方式二：使用 Git 命令行

需要先安装 Git: https://git-scm.com/download/win

```powershell
# 初始化 Git 仓库
cd c:\Users\Administrator\Desktop\wei
git init

# 添加所有文件
git add .

# 提交
git commit -m "Initial commit"

# 添加远程仓库（替换为你的 GitHub 仓库地址）
git remote add origin https://github.com/你的用户名/w-pay.git

# 推送到 GitHub
git branch -M main
git push -u origin main
```

### 方式三：直接在 GitHub 网页上传

1. 在 GitHub 创建新仓库
2. 上传整个 `wei` 文件夹的文件

## 手动构建 Docker 镜像

由于 GitHub Actions 缓存问题，建议手动构建 Docker 镜像：

### 在本地构建

```powershell
# 进入项目目录
cd c:\Users\Administrator\Desktop\wei

# 构建 Docker 镜像
docker build -t wpay:latest .
```

### 在服务器上构建

1. 克隆代码到服务器：
```bash
git clone https://github.com/rootshuiyu/w-pay.git
cd w-pay
```

2. 构建 Docker 镜像：
```bash
docker build -t wpay:latest .
```

## 在服务器上使用镜像

在服务器 8.219.129.48 上：

```bash
# 运行容器
docker run -d \
  --name wpay \
  -p 443:443 \
  -p 8090:8090 \
  -e APP_ENV=prod \
  -e PG_HOST=127.0.0.1 \
  -e PG_PORT=5432 \
  -e PG_USER=wpay \
  -e PG_PASSWORD=wpay123 \
  -e PG_DATABASE=wpay \
  -e REDIS_ADDR=127.0.0.1:6379 \
  -e JWT_SECRET=你的密钥 \
  -e TLS_CERT_FILE=/app/certs/server.crt \
  -e TLS_KEY_FILE=/app/certs/server.key \
  -v /opt/wpay/certs:/app/certs \
  wpay:latest
```

## 注意事项

1. 确保 Docker 已安装并运行
2. 确保 PostgreSQL 和 Redis 已配置
3. 确保 SSL 证书已放置在正确位置
4. 首次运行前需要初始化数据库
