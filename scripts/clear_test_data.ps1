# Clear all business test data (orders/channels/platforms/stores). Keeps admin + IP whitelist.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

$pgUser = if ($env:PG_USER) { $env:PG_USER } else { "wpay" }
$pgPass = if ($env:PG_PASSWORD) { $env:PG_PASSWORD } else { "wpay123" }
$pgHost = if ($env:PG_HOST) { $env:PG_HOST } else { "127.0.0.1" }
$pgPort = if ($env:PG_PORT) { $env:PG_PORT } else { "5432" }
$pgDb   = if ($env:PG_DATABASE) { $env:PG_DATABASE } else { "wpay" }
$redis  = if ($env:REDIS_ADDR) { $env:REDIS_ADDR } else { "127.0.0.1:6379" }

Write-Host ">>> Clearing PostgreSQL business data ($pgDb) ..." -ForegroundColor Cyan
$env:PGPASSWORD = $pgPass

$psql = Get-Command psql -ErrorAction SilentlyContinue
if ($psql) {
    & psql -h $pgHost -p $pgPort -U $pgUser -d $pgDb -f "$Root\sql\clear_business_data.sql"
} elseif (docker ps --format "{{.Names}}" 2>$null | Select-String -Pattern "wpay-postgres" -Quiet) {
    Get-Content "$Root\sql\clear_business_data.sql" | docker exec -i wpay-postgres psql -U $pgUser -d $pgDb
} else {
    Write-Host "psql/docker unavailable, using go run scripts/clear_data.go ..." -ForegroundColor Yellow
    $env:APP_ENV = if ($env:APP_ENV) { $env:APP_ENV } else { "dev" }
    Push-Location $Root
    go run scripts/clear_data.go
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit 1 }
    Pop-Location
    Write-Host ""
    exit 0
}

Write-Host ">>> Clearing Redis channel cache ..." -ForegroundColor Cyan
if (docker ps --format "{{.Names}}" 2>$null | Select-String -Pattern "wpay-redis" -Quiet) {
    docker exec wpay-redis redis-cli FLUSHDB | Out-Null
    Write-Host "    Redis FLUSHDB done (wpay-redis)" -ForegroundColor Green
} else {
    $redisCli = Get-Command redis-cli -ErrorAction SilentlyContinue
    if ($redisCli) {
        $rh = $redis.Split(":")[0]
        $rp = $redis.Split(":")[1]
        & redis-cli -h $rh -p $rp FLUSHDB | Out-Null
        Write-Host "    Redis FLUSHDB done" -ForegroundColor Green
    } else {
        Write-Host "    Skipped Redis (no redis-cli / wpay-redis)" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Cleared: orders, pay_channel, pay_platform, store, sensitive_logs, channel_history" -ForegroundColor Green
Write-Host "Kept: admin accounts, IP whitelist" -ForegroundColor Green
Write-Host "Re-login and add merchant codes via Settings -> pool tab." -ForegroundColor Yellow
