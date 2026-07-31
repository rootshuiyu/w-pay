# N逼pay E2E automation - covers VERIFY.md four core scenarios
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

function Test-Port([int]$Port) {
    try {
        $c = New-Object System.Net.Sockets.TcpClient
        $c.Connect("127.0.0.1", $Port)
        $c.Close()
        return $true
    } catch {
        return $false
    }
}

Write-Host "==> [1/4] Prepare test deps (PostgreSQL + Redis)" -ForegroundColor Cyan

# Prefer docker-compose.test.yml (5433/6380); fallback to local PG/Redis (5432/6379, database wpay_test)
$useDocker = $false
if (Get-Command docker -ErrorAction SilentlyContinue) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    docker compose -f docker-compose.test.yml up -d --wait 2>&1 | Out-Host
    $useDocker = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $prev
}

Write-Host "==> [2/4] Configure test env" -ForegroundColor Cyan
$env:APP_ENV = "test"
$env:WPAY_TEST_MODE = "1"
$env:JWT_SECRET = "test-jwt-secret"
$env:PG_HOST = "127.0.0.1"
$env:PG_DATABASE = "wpay_test"
$env:REDIS_DB = "1"

if ($useDocker) {
    Write-Host "    deps: docker-compose.test.yml (PostgreSQL:5433 / Redis:6380)" -ForegroundColor Green
    $env:PG_PORT = "5433"
    $env:PG_USER = "postgres"
    $env:PG_PASSWORD = "test123"
    $env:REDIS_ADDR = "127.0.0.1:6380"
} else {
    Write-Host "    Docker unavailable, fallback local PostgreSQL:5432 / Redis:6379" -ForegroundColor Yellow
    $env:PG_PORT = "5432"
    $env:PG_USER = "wpay"
    $env:PG_PASSWORD = "wpay123"
    $env:REDIS_ADDR = "127.0.0.1:6379"

    if (-not (Test-Port 5432)) {
        Write-Host "PostgreSQL 5432 is not listening. Start DB or Docker Desktop first." -ForegroundColor Red
        exit 1
    }
    if (-not (Test-Port 6379)) {
        Write-Host "Redis 6379 is not listening. Start Redis/Memurai or Docker Desktop first." -ForegroundColor Red
        exit 1
    }
    Write-Host "    require test DB: CREATE DATABASE wpay_test OWNER wpay;" -ForegroundColor DarkGray
}

Write-Host "==> [3/4] Download Go modules" -ForegroundColor Cyan
go mod download

Write-Host "==> [4/4] Run E2E tests (build tag: e2e)" -ForegroundColor Cyan
go test -tags=e2e -v -count=1 -timeout=5m ./tests/e2e/...
if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "E2E failed. See output above." -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "  E2E all scenarios passed" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
