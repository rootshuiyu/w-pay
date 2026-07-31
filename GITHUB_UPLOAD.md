# GitHub 上传和 Docker 镜像构建指南

## 已创建的文件

- ✅ `Dockerfile` - Docker 多阶段构建配置
- ✅ `.dockerignore` - Docker 构建排除文件
- ✅ `.github/workflows/docker-build.yml` - GitHub Actions 自动构建配置

## 上传到 GitHub 的步骤

### 方式一：使用 GitHub Desktop（推荐）

1. 下载 GitHub Desktop: https://desktop.github.com/
2. 安装并登录 GitHub 账号
3. 点击 "File" → "Add local repository"
4. 选择 `c:\Users\Administrator\Desktop\wei` 目录
5. 点击 "Publish repository" 发布到 GitHub
6. 设置仓库名称（如：wpay）
7. 推送后，GitHub Actions 会自动构建 Docker 镜像

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
git remote add origin https://github.com/你的用户名/wpay.git

# 推送到 GitHub
git branch -M main
git push -u origin main
```

### 方式三：直接在 GitHub 网页上传

1. 在 GitHub 创建新仓库
2. 上传整个 `wei` 文件夹的文件
3. 推送后 GitHub Actions 会自动构建

## GitHub Actions 工作原理

推送到 GitHub 后，GitHub Actions 会自动：
1. 检出代码
2. 使用 Docker Buildx 构建镜像
3. 将镜像推送到 GitHub Container Registry (ghcr.io)
4. 支持多平台构建（目前配置为 linux/amd64）

## 镜像标签规则

- `main` 分支：`ghcr.io/你的用户名/wpay:main`
- `v1.0.0` 标签：`ghcr.io/你的用户名/wpay:1.0.0`
- Pull Request：`ghcr.io/你的用户名/wpay:pr-123`

## 在服务器上使用镜像

构建完成后，在服务器 8.219.129.48 上：

```bash
# 登录 GitHub Container Registry
echo "你的GitHub_PAT" | docker login ghcr.io -u 你的GitHub用户名 --password-stdin

# 拉取镜像
docker pull ghcr.io/你的用户名/wpay:main

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
  ghcr.io/你的用户名/wpay:main
```

## 查看构建状态

在 GitHub 仓库页面点击 "Actions" 标签查看构建进度和日志。

## 注意事项

1. 确保 GitHub 仓库设置为 Public 或你的账号有 Packages 权限
2. 首次构建可能需要几分钟
3. 镜像会存储在 GitHub Container Registry，免费额度有限
