# Full end-to-end: health + platform isolation + pay flow
$Base = "http://127.0.0.1:8090"
$ErrorActionPreference = "Stop"
$Root = "c:\Users\Administrator\Desktop\wei"

function Invoke-Api {
    param(
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Path,
        $Body = $null,
        [hashtable]$ExtraHeaders = @{},
        [string]$Token = ""
    )
    $headers = @{ "Content-Type" = "application/json" }
    foreach ($k in $ExtraHeaders.Keys) { $headers[$k] = $ExtraHeaders[$k] }
    if ($Token) { $headers["Authorization"] = "Bearer $Token" }
    $uri = "$Base$Path"
    if ($null -ne $Body) {
        $json = $Body | ConvertTo-Json -Depth 8 -Compress
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers -Body ([System.Text.Encoding]::UTF8.GetBytes($json))
    }
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers
}

function Assert($cond, $msg) {
    if (-not $cond) { throw "ASSERT FAIL: $msg" }
}

Write-Host "`n========== 0. Health ==========" -ForegroundColor Cyan
$h = Invoke-Api GET "/health"
Assert ($h.status -eq "ok") "health not ok"
Write-Host "  OK status=$($h.status)"

Write-Host "`n========== 1. Admin login ==========" -ForegroundColor Cyan
$login = Invoke-Api POST "/api/admin/login" -Body @{ username = "admin"; password = "admin123" }
$token = $login.data.token
Assert ($token) "no token"
Write-Host "  OK role=$($login.data.role)"

$suffix = Get-Date -Format "HHmmssfff"

Write-Host "`n========== 2. Create 2 pay platforms ==========" -ForegroundColor Cyan
$platA = Invoke-Api POST "/api/admin/platform/add" -Token $token -Body @{
    platform_name = "SimPlatform-A-$suffix"
    platform_code = "SPA-$suffix"
}
$platB = Invoke-Api POST "/api/admin/platform/add" -Token $token -Body @{
    platform_name = "SimPlatform-B-$suffix"
    platform_code = "SPB-$suffix"
}
$keyA = $platA.data.app_key
$keyB = $platB.data.app_key
$idPlatA = $platA.data.id
$idPlatB = $platB.data.id
Write-Host "  PlatformA id=$idPlatA key=$keyA"
Write-Host "  PlatformB id=$idPlatB key=$keyB"

Write-Host "`n========== 3. Create stores + channels ==========" -ForegroundColor Cyan
$storeA = Invoke-Api POST "/api/admin/store/add" -Token $token -Body @{
    store_name = "SimStore-A-$suffix"; tax_subject = "GT-A"; address = "A"
}
$storeB = Invoke-Api POST "/api/admin/store/add" -Token $token -Body @{
    store_name = "SimStore-B-$suffix"; tax_subject = "GT-B"; address = "B"
}
$idStoreA = $storeA.data.id
$idStoreB = $storeB.data.id

$chIdsA = @()
$chIdsB = @()
foreach ($spec in @(
    @{ store = $idStoreA; mch = "MCH_PA1_$suffix"; plat = "A" }
    @{ store = $idStoreA; mch = "MCH_PA2_$suffix"; plat = "A" }
    @{ store = $idStoreB; mch = "MCH_PB1_$suffix"; plat = "B" }
    @{ store = $idStoreB; mch = "MCH_PB2_$suffix"; plat = "B" }
)) {
    $ch = Invoke-Api POST "/api/admin/channel/add" -Token $token -Body @{
        store_id         = $spec.store
        pay_type         = 1
        pool_enabled     = 1
        daily_limit_fen  = 500000
        single_limit_fen = 200000
        mch_no           = $spec.mch
        mch_key          = "test_key"
        app_id           = "wx_sim"
        serial_no        = "SN001"
        private_key      = "MOCK_KEY"
        notify_url       = "http://127.0.0.1:8090/api/notify/wx?store_id=$($spec.store)"
    }
    if ($spec.plat -eq "A") { $chIdsA += $ch.data.id } else { $chIdsB += $ch.data.id }
    Write-Host "  + $($spec.mch) id=$($ch.data.id) -> platform $($spec.plat)"
}

Write-Host "`n========== 4. Bind channels to platforms ==========" -ForegroundColor Cyan
Invoke-Api PUT "/api/admin/platform/set-channels" -Token $token -Body @{
    platform_id = $idPlatA; channel_ids = $chIdsA
} | Out-Null
Invoke-Api PUT "/api/admin/platform/set-channels" -Token $token -Body @{
    platform_id = $idPlatB; channel_ids = $chIdsB
} | Out-Null
Write-Host "  PlatformA channels: $($chIdsA -join ',')"
Write-Host "  PlatformB channels: $($chIdsB -join ',')"

Write-Host "`n========== 5. Platform pool overview ==========" -ForegroundColor Cyan
$poolA = Invoke-Api GET "/api/admin/platform/pool?platform_id=$idPlatA&pay_type=1" -Token $token
$poolB = Invoke-Api GET "/api/admin/platform/pool?platform_id=$idPlatB&pay_type=1" -Token $token
Write-Host "  PoolA count=$($poolA.data.list.Count) mchs=$($poolA.data.list.mch_no -join ', ')"
Write-Host "  PoolB count=$($poolB.data.list.Count) mchs=$($poolB.data.list.mch_no -join ', ')"

Write-Host "`n========== 6. Cashier create (platform A x2) ==========" -ForegroundColor Cyan
$ordersA = @()
foreach ($amt in @(10000, 12000)) {
    $r = Invoke-Api POST "/api/pay/create" -ExtraHeaders @{ "X-App-Key" = $keyA } -Body @{
        amount = $amt; pay_type = 1; pay_scene = "h5"
        return_url = "https://cashier.example.com/done"; subject = "A-$amt"
    }
    $ordersA += $r.data
    Write-Host "  A order=$($r.data.order_id) pay_url=$($r.data.pay_url)"
    Assert ($r.data.pay_url -match "MCH_PA") "platform A order used wrong mch in pay_url"
}

Write-Host "`n========== 7. Cashier create (platform B) ==========" -ForegroundColor Cyan
$rB = Invoke-Api POST "/api/pay/create" -ExtraHeaders @{ "X-App-Key" = $keyB } -Body @{
    amount = 8000; pay_type = 1; pay_scene = "h5"
    return_url = "https://cashier.example.com/done"; subject = "B-8000"
}
Write-Host "  B order=$($rB.data.order_id) pay_url=$($rB.data.pay_url)"
Assert ($rB.data.pay_url -match "MCH_PB") "platform B order used wrong mch"

Write-Host "`n========== 8. Public API hides internal fields ==========" -ForegroundColor Cyan
Assert ($null -eq $rB.data.mch_no) "mch_no should not expose"
Assert ($null -eq $rB.data.channel_id) "channel_id should not expose"
Write-Host "  OK create response has no mch_no/channel_id"

Write-Host "`n========== 9. Simulate pay + query ==========" -ForegroundColor Cyan
Set-Location $Root
go run ./scripts/sim_mark_paid.go "-order=$($ordersA[0].order_id)" "-amount=$($ordersA[0].amount)"
$q = Invoke-Api GET "/api/pay/query?order_no=$($ordersA[0].order_id)"
Assert ($q.data.order_status -eq 1) "order not paid"
Write-Host "  OK order_status=1 transaction=$($q.data.transaction_id)"

Write-Host "`n========== 10. Admin order has platform_id ==========" -ForegroundColor Cyan
$detail = Invoke-Api GET "/api/admin/order/detail?order_no=$($ordersA[0].order_id)" -Token $token
Write-Host "  platform_id=$($detail.data.platform_id) channel=$($detail.data.channel_id)"
Assert ([string]$detail.data.platform_id -eq [string]$idPlatA) "wrong platform on order"

Write-Host "`n========== 11. Reject invalid app_key ==========" -ForegroundColor Cyan
try {
    Invoke-Api POST "/api/pay/create" -ExtraHeaders @{ "X-App-Key" = "pk_invalid_key" } -Body @{
        amount = 100; pay_type = 1
    } | Out-Null
    throw "should reject invalid app_key"
} catch {
    $msg = $_.ErrorDetails.Message
    if (-not $msg) { $msg = $_.Exception.Message }
    Write-Host "  OK rejected: $msg"
}

Write-Host "`n========== 12. Frontend static ==========" -ForegroundColor Cyan
$fe = Invoke-WebRequest -Uri "$Base/" -UseBasicParsing -TimeoutSec 5
Assert ($fe.StatusCode -eq 200) "frontend not served"
Assert ($fe.Content -match "index") "no spa index"
Write-Host "  OK web/dist served at /"

Write-Host "`n========== ALL PASSED ==========" -ForegroundColor Green
Write-Host "Admin: $Base/#/platforms  $Base/#/channels  $Base/#/orders"
Write-Host "PlatformA app_key: $keyA"
