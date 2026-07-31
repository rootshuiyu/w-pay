# Simulate: multi-merchant pool + H5 pay + callback + quota
$Base = "http://127.0.0.1:8090"
$ErrorActionPreference = "Stop"

function Invoke-Api {
    param(
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Path,
        $Body = $null,
        [string]$Token = ""
    )
    $headers = @{ "Content-Type" = "application/json" }
    if ($Token) { $headers["Authorization"] = "Bearer $Token" }
    $uri = "$Base$Path"
    if ($null -ne $Body) {
        $json = $Body | ConvertTo-Json -Depth 6 -Compress
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers -Body ([System.Text.Encoding]::UTF8.GetBytes($json))
    }
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers
}

Write-Host "`n=== 1. Admin login ===" -ForegroundColor Cyan
$login = Invoke-Api -Method POST -Path "/api/admin/login" -Body @{ username = "admin"; password = "admin123" }
$token = $login.data.token
Write-Host "OK role=$($login.data.role)"

Write-Host "`n=== 2. Create 2 stores ===" -ForegroundColor Cyan
$suffix = Get-Date -Format "HHmmss"
$storeA = Invoke-Api -Method POST -Path "/api/admin/store/add" -Token $token -Body @{
    store_name = "SimStore-A-$suffix"; tax_subject = "GT-A"; address = "Addr-1"
}
$storeB = Invoke-Api -Method POST -Path "/api/admin/store/add" -Token $token -Body @{
    store_name = "SimStore-B-$suffix"; tax_subject = "GT-B"; address = "Addr-2"
}
$idA = $storeA.data.id
$idB = $storeB.data.id
Write-Host "StoreA=$idA StoreB=$idB"

Write-Host "`n=== 3. Add 3 WeChat channels to pool ===" -ForegroundColor Cyan
$specs = @(
    @{ store = $idA; mch = "MCH_ZHANG_01"; daily = 50000; single = 20000 }
    @{ store = $idA; mch = "MCH_ZHANG_02"; daily = 30000; single = 10000 }
    @{ store = $idB; mch = "MCH_LI_01";    daily = 80000; single = 50000 }
)
foreach ($m in $specs) {
    Invoke-Api -Method POST -Path "/api/admin/channel/add" -Token $token -Body @{
        store_id         = $m.store
        pay_type         = 1
        pool_enabled     = 1
        daily_limit_fen  = $m.daily
        single_limit_fen = $m.single
        mch_no           = $m.mch
        mch_key          = "test_key_$($m.mch)"
        app_id           = "wx_sim_test"
        serial_no        = "SN001"
        private_key      = "MOCK_PRIVATE_KEY"
        notify_url       = "http://127.0.0.1:8090/api/notify/wx?store_id=$($m.store)"
    } | Out-Null
    Write-Host "  + $($m.mch) daily_limit=$($m.daily/100) CNY"
}

Write-Host "`n=== 4. Pool overview ===" -ForegroundColor Cyan
$pool = Invoke-Api -Method GET -Path "/api/admin/channel/pool?pay_type=1" -Token $token
foreach ($ch in $pool.data.list) {
    $lim = if ($ch.daily_limit_fen -gt 0) { "$($ch.daily_limit_fen/100)" } else { "unlimited" }
    Write-Host "  $($ch.mch_no) used=$($ch.daily_used_fen/100) limit=$lim"
}

Write-Host "`n=== 5. Cashier API: 3 orders ===" -ForegroundColor Cyan
$orders = @()
foreach ($amt in @(10000, 15000, 8000)) {
    $r = Invoke-Api -Method POST -Path "/api/pay/create" -Body @{
        amount     = $amt
        pay_type   = 1
        pay_scene  = "h5"
        return_url = "https://cashier.example.com/done"
        subject    = "sim-$amt"
    }
    $d = $r.data
    $orders += $d
    Write-Host "  order=$($d.order_id) scene=$($d.pay_scene) amt=$($d.amount)"
    Write-Host "    pay_url=$($d.pay_url)"
}

Write-Host "`n=== 6. H5 create ===" -ForegroundColor Cyan
$h5 = Invoke-Api -Method POST -Path "/api/pay/create" -Body @{
    amount = 5000; pay_type = 1; pay_scene = "h5"
    return_url = "https://h5.example.com/ok"; subject = "h5-test"
}
Write-Host "  order=$($h5.data.order_id) pay_url=$($h5.data.pay_url)"

Write-Host "`n=== 7. Simulate payment success (1st order) ===" -ForegroundColor Cyan
$first = $orders[0]
Set-Location "c:\Users\Administrator\Desktop\wei"
go run ./scripts/sim_mark_paid.go "-order=$($first.order_id)" "-amount=$($first.amount)"

Write-Host "`n=== 8. Query order (cashier) ===" -ForegroundColor Cyan
$q = Invoke-Api -Method GET -Path "/api/pay/query?order_no=$($first.order_id)"
Write-Host "  order=$($q.data.order_id) status=$($q.data.order_status) (1=paid)"

Write-Host "`n=== 9. Pool quota after payment (admin) ===" -ForegroundColor Cyan
$detail = Invoke-Api -Method GET -Path "/api/admin/order/detail?order_no=$($first.order_id)" -Token $token
$paidChannelId = $detail.data.channel_id
$pool2 = Invoke-Api -Method GET -Path "/api/admin/channel/pool?pay_type=1" -Token $token
$pool2.data.list | Where-Object { $_.id -eq $paidChannelId } | ForEach-Object {
    Write-Host "  $($_.mch_no) daily_used=$($_.daily_used_fen/100) CNY"
}

Write-Host "`n=== 10. Reject over single limit ===" -ForegroundColor Cyan
try {
    Invoke-Api -Method POST -Path "/api/pay/create" -Body @{ amount = 9999999; pay_type = 1; pay_scene = "native" } | Out-Null
    Write-Host "  FAIL should reject"
} catch {
    $msg = $_.ErrorDetails.Message
    if (-not $msg) { $msg = $_.Exception.Message }
    Write-Host "  OK rejected: $msg"
}

Write-Host "`n=== DONE ===" -ForegroundColor Green
Write-Host "Admin: $Base/#/channels  Orders: $Base/#/orders"
