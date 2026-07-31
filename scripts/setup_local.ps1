#Requires -RunAsAdministrator
param(
    [switch]$StartOnly
)

# N逼pay 本地环境一键安装脚本（Windows）
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " N逼pay 本地环境安装（PostgreSQL + Redis + Go）" -ForegroundColor Cyan
Write-Host " 说明：本项目数据库为 PostgreSQL 16+，缓存为 Redis 7（可用 Memurai 替代）" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan

function Test-Cmd($name) { return [bool](Get-Command $name -ErrorAction SilentlyContinue) }

if (-not $StartOnly) {
    if (-not (Test-Cmd go)) {
        Write-Host "[1/4] 安装 Go ..." -ForegroundColor Green
        winget install -e --id GoLang.Go --source winget --accept-package-agreements --accept-source-agreements
    }
    if (-not (Test-Cmd docker)) {
        Write-Host "[2/4] 安装 Docker Desktop ..." -ForegroundColor Green
        winget install -e --id Docker.DockerDesktop --source winget --accept-package-agreements --accept-source-agreements
        Write-Host "  >>> 安装后请重启电脑并启动 Docker Desktop，再运行: .\scripts\setup_local.ps1 -StartOnly" -ForegroundColor Yellow
        exit 0
    }
}

$env:Path = "C:\Program Files\Go\bin;" + $env:Path

if (-not (Test-Cmd docker)) {
    Write-Host "Docker 未就绪。" -ForegroundColor Red
    exit 1
}

Write-Host "[3/4] 启动 PostgreSQL + Redis ..." -ForegroundColor Green
docker compose up -d --wait
docker compose ps

Write-Host "[4/4] 编译 N逼pay ..." -ForegroundColor Green
go mod tidy
go build -o wpay.exe main.go

Write-Host ""
Write-Host "完成！启动命令见 docs/LOCAL_SETUP.md" -ForegroundColor Green
