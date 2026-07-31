# N逼pay 云端部署脚本 (Windows)

Write-Host "=== N逼pay 云端部署脚本 ===" -ForegroundColor Green

# 检查环境变量
if (-not $env:PG_HOST -or -not $env:PG_PASSWORD -or -not $env:JWT_SECRET) {
    Write-Host "错误：请设置必要的环境变量" -ForegroundColor Red
    Write-Host "必需变量：PG_HOST, PG_PASSWORD, JWT_SECRET"
    exit 1
}

# 设置生产环境
$env:APP_ENV = "prod"

# 编译
Write-Host "正在编译..." -ForegroundColor Yellow
go mod tidy
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o wpay main.go

Write-Host "编译完成：wpay" -ForegroundColor Green

# 初始化数据库（如果需要）
if ($env:INIT_DB -eq "true") {
    Write-Host "正在初始化数据库..." -ForegroundColor Yellow
    psql -h $env:PG_HOST -U postgres -c "CREATE ROLE wpay LOGIN PASSWORD '$env:PG_PASSWORD';" 2>$null
    psql -h $env:PG_HOST -U postgres -c "CREATE DATABASE wpay OWNER wpay;" 2>$null
    psql -h $env:PG_HOST -U wpay -d wpay -f sql/init.sql 2>$null
}

Write-Host "部署完成！" -ForegroundColor Green
Write-Host "请将以下文件上传到服务器："
Write-Host "  - wpay (可执行文件)"
Write-Host "  - config/prod.yaml"
Write-Host "  - sql/init.sql (如需初始化数据库)"
Write-Host ""
Write-Host "服务器上需要设置的环境变量见 DEPLOY.md"
