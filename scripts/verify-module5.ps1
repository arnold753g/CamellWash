param(
    [switch]$VerifyPayments,
    [switch]$VerifyReports,
    [switch]$KeepEnvironment,
    [int]$ServerPort = 18085
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Net.Http

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendRoot = Join-Path $projectRoot 'backend'
$workspaceTemp = Join-Path $projectRoot '.tmp'
New-Item -ItemType Directory -Force -Path $workspaceTemp | Out-Null
Get-Content (Join-Path $projectRoot '.env') | ForEach-Object {
    if ($_ -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') { [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process') }
}

$databaseName = 'lavadero_m5_e2e_' + (Get-Date -Format 'yyyyMMddHHmmss')
if ($databaseName -notmatch '^lavadero_m5_e2e_[0-9]{14}$') { throw 'Unsafe temporary database name' }
if (Get-NetTCPConnection -LocalPort $ServerPort -State Listen -ErrorAction SilentlyContinue) { throw "Port $ServerPort is already in use" }
$createDb = (Get-Command createdb.exe -ErrorAction Stop).Source
$dropDb = (Get-Command dropdb.exe -ErrorAction Stop).Source
$psql = (Get-Command psql.exe -ErrorAction Stop).Source
$env:PGPASSWORD = $env:DB_PASSWORD
$env:GOCACHE = Join-Path $workspaceTemp 'go-build'
$apiExecutable = Join-Path $workspaceTemp 'm5-api.exe'
$apiProcess = $null

function New-WebSession { New-Object Microsoft.PowerShell.Commands.WebRequestSession }
function Invoke-TestRequest($method, $path, $body, $session, $headers = @{}) {
    try {
        $arguments = @{ Uri = "$script:apiBase$path"; Method = $method; WebSession = $session; Headers = $headers; UseBasicParsing = $true }
        if ($null -ne $body) {
            $arguments.ContentType = 'application/json; charset=utf-8'
            $json = $body | ConvertTo-Json -Compress -Depth 8
            $arguments.Body = [Text.Encoding]::UTF8.GetBytes($json)
        }
        Invoke-WebRequest @arguments
    } catch { $_.Exception.Response }
}
function Get-Status($response) { [int]$response.StatusCode }
function Get-Body($response) {
    if ($response -is [Microsoft.PowerShell.Commands.HtmlWebResponseObject]) { return [string]$response.Content }
    if ($response.Content) { return [string]$response.Content }
    $stream = $response.GetResponseStream()
    if ($null -eq $stream) { return '' }
    $reader = New-Object IO.StreamReader($stream)
    try { return $reader.ReadToEnd() } finally { $reader.Dispose() }
}
function Get-JSON($response) { (Get-Body $response) | ConvertFrom-Json }
function Assert-Status($response, $expected, $message) {
    $actual = Get-Status $response
    if ($actual -ne $expected) {
        $body = Get-Body $response
        $detail = if ($body) { $body } else { '<empty response body>' }
        throw "$message (expected $expected, got $actual): $detail"
    }
}

& $createDb -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER $databaseName
if ($LASTEXITCODE -ne 0) { throw 'Could not create temporary database' }

try {
    $env:DB_NAME = $databaseName; $env:APP_ENV = 'test'; $env:DB_AUTO_MIGRATE = 'false'
    $env:SERVER_HOST = '127.0.0.1'; $env:SERVER_PORT = [string]$serverPort; $env:AUTH_COOKIE_SECURE = 'false'
    if ($KeepEnvironment) { $env:ALLOWED_ORIGINS = "$env:ALLOWED_ORIGINS,http://localhost:3001,http://127.0.0.1:3001" }
    $env:SEED_DEVELOPMENT = 'true'; $env:AUTH_RATE_LIMIT_PER_MINUTE = '200'; $env:API_RATE_LIMIT_PER_MINUTE = '1000'
    Push-Location $backendRoot
    go run ./cmd/migrate; if ($LASTEXITCODE -ne 0) { throw 'Migration failed' }
    go run ./cmd/migrate; if ($LASTEXITCODE -ne 0) { throw 'Second migration failed' }
    go run ./cmd/seed; if ($LASTEXITCODE -ne 0) { throw 'Seed failed' }
    go build -o $apiExecutable ./cmd/api; if ($LASTEXITCODE -ne 0) { throw 'API build failed' }
    Pop-Location

    $startInfo = New-Object Diagnostics.ProcessStartInfo
    $startInfo.FileName = $apiExecutable; $startInfo.WorkingDirectory = $backendRoot; $startInfo.UseShellExecute = $true
    $startInfo.WindowStyle = [Diagnostics.ProcessWindowStyle]::Hidden
    $apiProcess = [Diagnostics.Process]::Start($startInfo)
    $script:apiBase = "http://127.0.0.1:$serverPort/api/v1"
    $healthy = $false
    for ($attempt = 0; $attempt -lt 20; $attempt++) { Start-Sleep -Milliseconds 500; try { if ((Invoke-WebRequest "http://127.0.0.1:$serverPort/health" -UseBasicParsing).StatusCode -eq 200) { $healthy = $true; break } } catch {} }
    if (-not $healthy) { throw 'API did not become healthy' }

    $origin = @{ Origin = 'http://localhost:3000' }
    $adminSession = New-WebSession
    $adminLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $env:DEV_ADMIN_EMAIL; password = $env:DEV_ADMIN_PASSWORD } $adminSession $origin
    Assert-Status $adminLogin 200 'Administrator login failed'
    $adminHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $adminLogin).data.csrf_token }

    $managerPassword = 'correct horse manager orders 2026!'
    $washPassword = 'correct horse wash orders 2026!'
    $washTwoPassword = 'correct horse wash two orders 2026!'
    $fixtures = @(
        @{ email = 'manager-orders@example.test'; password = $managerPassword; first_name = 'Mara'; last_name = 'Orden'; role = 'manager'; status = 'active' },
        @{ email = 'wash-one-orders@example.test'; password = $washPassword; first_name = 'Luz'; last_name = 'Uno'; role = 'wash_staff'; status = 'active' },
        @{ email = 'wash-two-orders@example.test'; password = $washTwoPassword; first_name = 'Sol'; last_name = 'Dos'; role = 'wash_staff'; status = 'active' }
    )
    $fixtureIDs = @{}
    foreach ($fixture in $fixtures) {
        $result = Invoke-TestRequest 'POST' '/admin/users' $fixture $adminSession $adminHeaders
        Assert-Status $result 201 'Could not create role fixture'
        $fixtureIDs[$fixture.email] = (Get-JSON $result).data.id
    }
    $managerID = $fixtureIDs['manager-orders@example.test']; $washOneID = $fixtureIDs['wash-one-orders@example.test']; $washTwoID = $fixtureIDs['wash-two-orders@example.test']

    $managerSession = New-WebSession
    $managerLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = 'manager-orders@example.test'; password = $managerPassword } $managerSession $origin
    Assert-Status $managerLogin 200 'Manager login failed'
    $managerHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $managerLogin).data.csrf_token }
    $washOneSession = New-WebSession
    $washOneLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = 'wash-one-orders@example.test'; password = $washPassword } $washOneSession $origin
    Assert-Status $washOneLogin 200 'Wash one login failed'
    $washOneHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $washOneLogin).data.csrf_token }
    $washTwoSession = New-WebSession
    $washTwoLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = 'wash-two-orders@example.test'; password = $washTwoPassword } $washTwoSession $origin
    Assert-Status $washTwoLogin 200 'Wash two login failed'
    $washTwoHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $washTwoLogin).data.csrf_token }

    $clientDisplayName = 'Cliente ' + [char]0x00D3 + 'rdenes'
    $clientResult = Invoke-TestRequest 'POST' '/clients' @{ full_name = $clientDisplayName; ci = $null; phone = '70020001' } $adminSession $adminHeaders
    Assert-Status $clientResult 201 'Client creation failed'; $clientID = (Get-JSON $clientResult).data.id
    $vehicleTypeID = @((Get-JSON (Invoke-TestRequest 'GET' '/catalogs/vehicle-types' $null $adminSession)).data.vehicle_types)[0].id
    $vehicleResult = Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = $null; brand = 'Toyota'; model = 'Competición'; color = 'Negro' } $adminSession $adminHeaders
    Assert-Status $vehicleResult 201 'Vehicle creation failed'; $vehicleID = (Get-JSON $vehicleResult).data.id

    $matrix = Get-JSON (Invoke-TestRequest 'GET' "/admin/service-prices/matrix?vehicle_type_id=$vehicleTypeID" $null $adminSession)
    $cells = @($matrix.data.cells)
    if ($cells.Count -lt 3) { throw 'Not enough services for order E2E' }
    $first = $cells[0]; $second = $cells[1]; $missing = $cells[2]
    Assert-Status (Invoke-TestRequest 'POST' '/admin/service-prices' @{ service_id = $first.service_id; vehicle_type_id = $vehicleTypeID; price = '25.10' } $adminSession $adminHeaders) 201 'First price failed'
    Assert-Status (Invoke-TestRequest 'POST' '/admin/service-prices' @{ service_id = $second.service_id; vehicle_type_id = $vehicleTypeID; price = '14.90' } $adminSession $adminHeaders) 201 'Second price failed'

    $future = (Get-Date).ToUniversalTime().AddHours(2).ToString('o')
    $createBody = @{ vehicle_id = $vehicleID; service_ids = @($first.service_id, $second.service_id); staff_ids = @($washOneID, $washTwoID); estimated_delivery_at = $future }
    $created = Invoke-TestRequest 'POST' '/orders' $createBody $managerSession $managerHeaders
    Assert-Status $created 201 'Atomic order creation failed'
    $order = (Get-JSON $created).data; $orderID = $order.id
    if ($order.total -ne '40.00' -or $order.services.Count -ne 2 -or $order.staff.Count -ne 2 -or $order.history.Count -ne 1) { throw 'Order aggregate was not created atomically or total is wrong' }
    if ($order.services[0].service_name_snapshot.Length -lt 2) { throw 'Service name snapshot is missing' }

    $beforeInvalid = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -c 'SELECT COUNT(*) FROM orders;'
    Assert-Status (Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($first.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future; total = '0.01'; price = '0.01'; status = 'DELIVERED' } $managerSession $managerHeaders) 400 'Mass-assigned protected fields were accepted'
    Assert-Status (Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($first.service_id, $first.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $managerSession $managerHeaders) 422 'Duplicate service was accepted'
    Assert-Status (Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($first.service_id); staff_ids = @($washOneID, $washOneID); estimated_delivery_at = $future } $managerSession $managerHeaders) 422 'Duplicate staff was accepted'
    Assert-Status (Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($missing.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $managerSession $managerHeaders) 422 'Service without price was accepted'
    Assert-Status (Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($first.service_id); staff_ids = @($washOneID); estimated_delivery_at = (Get-Date).ToUniversalTime().AddHours(-1).ToString('o') } $managerSession $managerHeaders) 422 'Past estimated delivery was accepted'
    $afterInvalid = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -c 'SELECT COUNT(*) FROM orders;'
    if ($beforeInvalid.Trim() -ne $afterInvalid.Trim()) { throw 'Failed creations left partial orders' }

    $originalName = @($order.services | Where-Object { $_.service_id -eq $first.service_id })[0].service_name_snapshot
    Assert-Status (Invoke-TestRequest 'PATCH' "/admin/service-prices/$($first.service_id)/$vehicleTypeID" @{ price = '99.00' } $adminSession $adminHeaders) 200 'Price change failed'
    Assert-Status (Invoke-TestRequest 'PATCH' "/admin/services/$($first.service_id)" @{ category_id = $first.category_id; name = 'Servicio renombrado después' } $adminSession $adminHeaders) 200 'Service rename failed'
    $snapshotDetail = (Get-JSON (Invoke-TestRequest 'GET' "/orders/$orderID" $null $managerSession)).data
    $snapshotLine = @($snapshotDetail.services | Where-Object { $_.service_id -eq $first.service_id })[0]
    if ($snapshotLine.price_snapshot -ne '25.10' -or $snapshotLine.service_name_snapshot -ne $originalName -or $snapshotDetail.total -ne '40.00') { throw 'Historical price or name changed after catalog update' }

    Assert-Status (Invoke-TestRequest 'GET' '/orders' $null $washOneSession) 200 'Assigned staff could not list own orders'
    Assert-Status (Invoke-TestRequest 'GET' "/orders/$orderID" $null $washTwoSession) 200 'Second assigned staff could not read order'
    $unassignedOrder = Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($second.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $adminSession $adminHeaders
    Assert-Status $unassignedOrder 201 'Unassigned fixture order failed'; $unassignedOrderID = (Get-JSON $unassignedOrder).data.id
    Assert-Status (Invoke-TestRequest 'GET' "/orders/$unassignedOrderID" $null $washTwoSession) 404 'Unassigned staff read another order (IDOR)'
    Assert-Status (Invoke-TestRequest 'POST' "/orders/$unassignedOrderID/transitions" @{ version = 1; target = 'IN_PROGRESS'; reason = '' } $washTwoSession $washTwoHeaders) 404 'Unassigned staff changed another order (IDOR)'

    $started = Invoke-TestRequest 'POST' "/orders/$orderID/transitions" @{ version = 1; target = 'IN_PROGRESS'; reason = '' } $washOneSession $washOneHeaders
    Assert-Status $started 200 'Assigned staff could not start order'
    $cookieName = 'starter_session'; $washCookie = @($washOneSession.Cookies.GetCookies($script:apiBase) | Where-Object { $_.Name -eq $cookieName })[0]
    $cookieContainer = New-Object Net.CookieContainer; $cookieContainer.Add((New-Object Net.Cookie($cookieName, $washCookie.Value, '/api/v1', '127.0.0.1')))
    $httpHandler = New-Object Net.Http.HttpClientHandler; $httpHandler.CookieContainer = $cookieContainer; $httpClient = New-Object Net.Http.HttpClient($httpHandler)
    try {
        $tasks = @()
        foreach ($observation in @('Primera solicitud', 'Segunda solicitud')) {
            $request = New-Object Net.Http.HttpRequestMessage([Net.Http.HttpMethod]::Post, "$script:apiBase/orders/$orderID/transitions")
            $request.Headers.Add('Origin', 'http://localhost:3000'); $request.Headers.Add('X-CSRF-Token', $washOneHeaders['X-CSRF-Token'])
            $request.Content = New-Object Net.Http.StringContent((@{ version = 2; target = 'QUALITY_CONTROL'; reason = $observation } | ConvertTo-Json -Compress), [Text.Encoding]::UTF8, 'application/json')
            $tasks += $httpClient.SendAsync($request)
        }
        [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]$tasks)
        $statuses = @($tasks | ForEach-Object { [int]$_.Result.StatusCode } | Sort-Object)
        if (($statuses -join ',') -ne '200,409') { throw "Concurrent transitions returned $($statuses -join ',')" }
    } finally { $httpClient.Dispose(); $httpHandler.Dispose() }
    $quality = (Get-JSON (Invoke-TestRequest 'GET' "/orders/$orderID" $null $managerSession)).data
    if ($quality.status -ne 'QUALITY_CONTROL' -or @($quality.history | Where-Object { $_.to_status -eq 'QUALITY_CONTROL' }).Count -ne 1) { throw 'Concurrent transition corrupted state/history' }
    Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/transitions" @{ version = $quality.version; target = 'FINISHED'; reason = '' } $washOneSession $washOneHeaders) 403 'Wash staff approved quality'
    $finished = Invoke-TestRequest 'POST' "/orders/$orderID/transitions" @{ version = $quality.version; target = 'FINISHED'; reason = '' } $managerSession $managerHeaders
    Assert-Status $finished 200 'Manager could not approve quality'; $finishedData = (Get-JSON $finished).data
    Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/transitions" @{ version = $finishedData.version; target = 'DELIVERED'; reason = '' } $managerSession $managerHeaders) 409 'Delivery was not blocked pending payment'

    $cancelFixture = Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($second.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $adminSession $adminHeaders
    Assert-Status $cancelFixture 201 'Cancel fixture failed'; $cancelID = (Get-JSON $cancelFixture).data.id
    Assert-Status (Invoke-TestRequest 'POST' "/orders/$cancelID/cancel" @{ version = 1; reason = '' } $managerSession $managerHeaders) 422 'Cancellation without reason was accepted'
    $cancelled = Invoke-TestRequest 'POST' "/orders/$cancelID/cancel" @{ version = 1; reason = 'Cliente canceló la atención' } $managerSession $managerHeaders
    Assert-Status $cancelled 200 'Manager cancellation failed'; $cancelledData = (Get-JSON $cancelled).data
    Assert-Status (Invoke-TestRequest 'POST' "/orders/$cancelID/transitions" @{ version = $cancelledData.version; target = 'IN_PROGRESS'; reason = '' } $managerSession $managerHeaders) 409 'Cancelled order was reopened'

    $editableFixture = Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($second.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $adminSession $adminHeaders
    Assert-Status $editableFixture 201 'Edit fixture failed'; $editable = (Get-JSON $editableFixture).data
    $later = (Get-Date).ToUniversalTime().AddHours(3).ToString('o')
    $updated = Invoke-TestRequest 'PATCH' "/orders/$($editable.id)" @{ version = $editable.version; staff_ids = @($washOneID, $washTwoID); estimated_delivery_at = $later } $managerSession $managerHeaders
    Assert-Status $updated 200 'Pending order update failed'; $updatedData = (Get-JSON $updated).data
    $inProgress = Invoke-TestRequest 'POST' "/orders/$($editable.id)/transitions" @{ version = $updatedData.version; target = 'IN_PROGRESS'; reason = '' } $managerSession $managerHeaders
    Assert-Status $inProgress 200 'Manager start failed'; $inProgressData = (Get-JSON $inProgress).data
    Assert-Status (Invoke-TestRequest 'PATCH' "/orders/$($editable.id)" @{ version = $inProgressData.version; service_ids = @($first.service_id); staff_ids = @($washOneID); estimated_delivery_at = $later } $managerSession $managerHeaders) 409 'Services changed after order started'
    Assert-Status (Invoke-TestRequest 'PATCH' "/orders/$($editable.id)" @{ version = $inProgressData.version; staff_ids = @($washOneID); estimated_delivery_at = $later } $managerSession $managerHeaders) 200 'Allowed in-progress fields could not be updated'

    $clientHistory = Get-JSON (Invoke-TestRequest 'GET' "/clients/$clientID/orders?limit=100" $null $managerSession)
    $vehicleHistory = Get-JSON (Invoke-TestRequest 'GET' "/vehicles/$vehicleID/orders?limit=100" $null $managerSession)
    if ($clientHistory.data.pagination.total -ne 4 -or $vehicleHistory.data.pagination.total -ne 4) { throw 'Client or vehicle history is incomplete' }
    $previousErrorPreference = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    $historyMutation = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -v ON_ERROR_STOP=1 -c "UPDATE order_status_history SET reason='tampered' WHERE order_id=$orderID;" 2>&1
    $historyMutationExit = $LASTEXITCODE
    $ErrorActionPreference = $previousErrorPreference
    if ($historyMutationExit -eq 0) { throw 'Append-only history accepted an update' }
    $historyCount = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -c "SELECT COUNT(*) FROM order_status_history WHERE order_id=$orderID;"
    if ($historyCount.Trim() -ne '4') { throw 'Order history count is inconsistent' }

    if ($VerifyPayments) {
        Assert-Status (Invoke-TestRequest 'GET' "/orders/$orderID/payments" $null $washOneSession) 403 'Wash staff accessed payment details'
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/payments" @{ version = $finishedData.version; method = 'CASH'; external_reference = ''; idempotency_key = 'wash_payment_attempt_0001'; deliver = $true } $washOneSession $washOneHeaders) 403 'Wash staff registered a payment'
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/payments" @{ version = $finishedData.version; method = 'CRYPTO'; external_reference = ''; idempotency_key = 'invalid_method_payment_01'; deliver = $true } $managerSession $managerHeaders) 422 'Unknown payment method was accepted'
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/payments" @{ version = $finishedData.version; method = 'QR'; external_reference = 'QR-100'; idempotency_key = 'protected_amount_payment_1'; deliver = $true; amount = '0.01' } $managerSession $managerHeaders) 400 'Client-controlled payment amount was accepted'
        $stillFinished = (Get-JSON (Invoke-TestRequest 'GET' "/orders/$orderID" $null $managerSession)).data
        if ($stillFinished.status -ne 'FINISHED') { throw 'Failed payment delivered the order' }

        $deliveryPaymentBody = @{ version = $finishedData.version; method = 'QR'; external_reference = 'QR-ORDEN-100'; idempotency_key = 'payment_delivery_order_100'; deliver = $true }
        $deliveryPayment = Invoke-TestRequest 'POST' "/orders/$orderID/payments" $deliveryPaymentBody $managerSession $managerHeaders
        Assert-Status $deliveryPayment 201 'Manager payment and delivery failed'
        $deliveryPaymentData = (Get-JSON $deliveryPayment).data
        if ($deliveryPaymentData.payment.amount -ne '40.00' -or $deliveryPaymentData.order.status -ne 'DELIVERED') { throw 'Payment amount or atomic delivery is inconsistent' }
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/payments" $deliveryPaymentBody $managerSession $managerHeaders) 200 'Idempotent payment retry failed'
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$orderID/payments" @{ version = $deliveryPaymentData.order.version; method = 'CASH'; external_reference = ''; idempotency_key = 'second_delivery_payment_01'; deliver = $false } $managerSession $managerHeaders) 409 'Second valid payment was accepted'

        $deliveredCorrection = Invoke-TestRequest 'POST' "/payments/$($deliveryPaymentData.payment.id)/correct" @{ version = $deliveryPaymentData.order.version; method = 'BANK_TRANSFER'; external_reference = 'TR-200'; idempotency_key = 'corrected_delivery_pay_01'; reason = 'Método registrado incorrectamente' } $managerSession $managerHeaders
        Assert-Status $deliveredCorrection 201 'Delivered payment correction failed'
        $deliveredCorrectionData = (Get-JSON $deliveredCorrection).data
        if ($deliveredCorrectionData.order.status -ne 'DELIVERED' -or $deliveredCorrectionData.payment.amount -ne '40.00') { throw 'Delivered correction broke payment/order consistency' }
        $deliveredPayments = @((Get-JSON (Invoke-TestRequest 'GET' "/orders/$orderID/payments" $null $managerSession)).data.payments)
        if (@($deliveredPayments | Where-Object { $_.status -eq 'VALID' }).Count -ne 1 -or $deliveredPayments.Count -ne 2) { throw 'Correction did not preserve exactly one valid payment' }

        $earlyPayment = Invoke-TestRequest 'POST' "/orders/$unassignedOrderID/payments" @{ version = 1; method = 'CASH'; external_reference = ''; idempotency_key = 'early_payment_order_001'; deliver = $false } $managerSession $managerHeaders
        Assert-Status $earlyPayment 201 'Early payment failed'
        $earlyPaymentData = (Get-JSON $earlyPayment).data
        if ($earlyPaymentData.order.status -ne 'PENDING' -or $earlyPaymentData.payment.amount -ne '14.90') { throw 'Early payment changed state or amount' }
        Assert-Status (Invoke-TestRequest 'PATCH' "/orders/$unassignedOrderID" @{ version = 1; service_ids = @($first.service_id); staff_ids = @($washOneID); estimated_delivery_at = $later } $managerSession $managerHeaders) 409 'Paid order services were changed'
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$unassignedOrderID/cancel" @{ version = 1; reason = 'Cliente solicita cancelar' } $managerSession $managerHeaders) 409 'Paid order was cancelled without voiding payment'
        $voidedEarly = Invoke-TestRequest 'POST' "/payments/$($earlyPaymentData.payment.id)/void" @{ version = 1; reason = 'Cobro registrado por error operativo' } $managerSession $managerHeaders
        Assert-Status $voidedEarly 200 'Manager could not void an early payment'
        $voidedEarlyData = (Get-JSON $voidedEarly).data
        if ($voidedEarlyData.payment.status -ne 'VOIDED' -or $voidedEarlyData.payment.review_status -ne 'PENDING_REVIEW' -or $voidedEarlyData.order.version -ne 2) { throw 'Manager void lacks pending review or version update' }
        Assert-Status (Invoke-TestRequest 'POST' "/orders/$cancelID/payments" @{ version = $cancelledData.version; method = 'CASH'; external_reference = ''; idempotency_key = 'cancelled_order_payment1'; deliver = $false } $managerSession $managerHeaders) 409 'Cancelled order accepted payment'

        $adminOrder = Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($second.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $adminSession $adminHeaders
        Assert-Status $adminOrder 201 'Admin payment fixture failed'; $adminOrderData = (Get-JSON $adminOrder).data
        $adminPayment = Invoke-TestRequest 'POST' "/orders/$($adminOrderData.id)/payments" @{ version = $adminOrderData.version; method = 'CARD'; external_reference = 'POS-300'; idempotency_key = 'admin_payment_order_001'; deliver = $false } $adminSession $adminHeaders
        Assert-Status $adminPayment 201 'Administrator payment failed'; $adminPaymentData = (Get-JSON $adminPayment).data
        if ($adminPaymentData.payment.amount -ne '14.90') { throw 'Administrator payment amount is wrong' }
        $adminVoid = Invoke-TestRequest 'POST' "/payments/$($adminPaymentData.payment.id)/void" @{ version = $adminPaymentData.order.version; reason = 'Prueba de anulación administrativa' } $adminSession $adminHeaders
        Assert-Status $adminVoid 200 'Administrator void failed'
        if ((Get-JSON $adminVoid).data.payment.review_status -ne 'REVIEWED') { throw 'Administrator void should be self-authorized' }

        $editableCurrent = (Get-JSON (Invoke-TestRequest 'GET' "/orders/$($editable.id)" $null $managerSession)).data
        $managerCookie = @($managerSession.Cookies.GetCookies($script:apiBase) | Where-Object { $_.Name -eq $cookieName })[0]
        $paymentCookieContainer = New-Object Net.CookieContainer; $paymentCookieContainer.Add((New-Object Net.Cookie($cookieName, $managerCookie.Value, '/api/v1', '127.0.0.1')))
        $paymentHandler = New-Object Net.Http.HttpClientHandler; $paymentHandler.CookieContainer = $paymentCookieContainer; $paymentClient = New-Object Net.Http.HttpClient($paymentHandler)
        try {
            $paymentTasks = @()
            foreach ($key in @('concurrent_payment_key_01', 'concurrent_payment_key_02')) {
                $paymentRequest = New-Object Net.Http.HttpRequestMessage([Net.Http.HttpMethod]::Post, "$script:apiBase/orders/$($editable.id)/payments")
                $paymentRequest.Headers.Add('Origin', 'http://localhost:3000'); $paymentRequest.Headers.Add('X-CSRF-Token', $managerHeaders['X-CSRF-Token'])
                $paymentRequest.Content = New-Object Net.Http.StringContent((@{ version = $editableCurrent.version; method = 'QR'; external_reference = 'QR-CONCURRENTE'; idempotency_key = $key; deliver = $false } | ConvertTo-Json -Compress), [Text.Encoding]::UTF8, 'application/json')
                $paymentTasks += $paymentClient.SendAsync($paymentRequest)
            }
            [Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]$paymentTasks)
            $paymentStatuses = @($paymentTasks | ForEach-Object { [int]$_.Result.StatusCode } | Sort-Object)
            if (($paymentStatuses -join ',') -ne '201,409') { throw "Concurrent payments returned $($paymentStatuses -join ',')" }
        } finally { $paymentClient.Dispose(); $paymentHandler.Dispose() }
        $concurrentList = @((Get-JSON (Invoke-TestRequest 'GET' "/orders/$($editable.id)/payments" $null $managerSession)).data.payments)
        $concurrentValid = @($concurrentList | Where-Object { $_.status -eq 'VALID' })
        if ($concurrentValid.Count -ne 1) { throw 'Concurrent requests created more than one valid payment' }
        $correctedConcurrent = Invoke-TestRequest 'POST' "/payments/$($concurrentValid[0].id)/correct" @{ version = $editableCurrent.version; method = 'CARD'; external_reference = 'POS-400'; idempotency_key = 'correct_concurrent_pay01'; reason = 'Corrección del método de cobro' } $managerSession $managerHeaders
        Assert-Status $correctedConcurrent 201 'Atomic payment replacement failed'

        Assert-Status (Invoke-TestRequest 'GET' '/admin/payment-reviews?limit=100' $null $managerSession) 403 'Manager accessed administrator review queue'
        $pendingReviews = Get-JSON (Invoke-TestRequest 'GET' '/admin/payment-reviews?limit=100' $null $adminSession)
        if ($pendingReviews.data.pagination.total -lt 3) { throw 'Pending review queue is incomplete' }
        foreach ($review in @($pendingReviews.data.payments)) {
            Assert-Status (Invoke-TestRequest 'POST' "/admin/payment-reviews/$($review.id)" @{ decision = 'REVIEWED'; note = 'Control administrativo conforme' } $adminSession $adminHeaders) 200 'Administrator review failed'
        }
        $remainingReviews = Get-JSON (Invoke-TestRequest 'GET' '/admin/payment-reviews?limit=100' $null $adminSession)
        if ($remainingReviews.data.pagination.total -ne 0) { throw 'Reviewed voids remained in the pending queue' }

        $previousErrorPreference = $ErrorActionPreference
        $ErrorActionPreference = 'Continue'
        $eventMutation = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -v ON_ERROR_STOP=1 -c "DELETE FROM payment_events WHERE payment_id=$($deliveredCorrectionData.payment.id);" 2>&1
        $eventMutationExit = $LASTEXITCODE
        $ErrorActionPreference = $previousErrorPreference
        if ($eventMutationExit -eq 0) { throw 'Append-only payment events accepted deletion' }
        if ($KeepEnvironment) {
            $visualOrder = Invoke-TestRequest 'POST' '/orders' @{ vehicle_id = $vehicleID; service_ids = @($second.service_id); staff_ids = @($washOneID); estimated_delivery_at = $future } $managerSession $managerHeaders
            Assert-Status $visualOrder 201 'Visual payment fixture failed'; $visualOrderData = (Get-JSON $visualOrder).data
            $visualPayment = Invoke-TestRequest 'POST' "/orders/$($visualOrderData.id)/payments" @{ version = $visualOrderData.version; method = 'QR'; external_reference = 'QR-VISUAL-500'; idempotency_key = 'visual_payment_review_01'; deliver = $false } $managerSession $managerHeaders
            Assert-Status $visualPayment 201 'Visual payment registration failed'; $visualPaymentData = (Get-JSON $visualPayment).data
            $visualVoidReason = 'Anulaci' + [char]0x00F3 + 'n pendiente para revisi' + [char]0x00F3 + 'n visual'
            Assert-Status (Invoke-TestRequest 'POST' "/payments/$($visualPaymentData.payment.id)/void" @{ version = $visualPaymentData.order.version; reason = $visualVoidReason } $managerSession $managerHeaders) 200 'Visual payment void failed'
            Write-Output "VISUAL_FIXTURES order_id=$($visualOrderData.id) review_payment_id=$($visualPaymentData.payment.id)"
        }
        Write-Output 'MODULE6_E2E_OK timing=dropoff+pickup exact_total=server methods=validated idempotency=201/200 concurrency=201/409 delivery=atomic void=manager_review_pending correction=atomic review=admin append_only=payment_events roles=protected'
    }

    if ($VerifyReports) {
        $today = Get-Date -Format 'yyyy-MM-dd'
        $range = "date_from=$today&date_to=$today"
        $watch = [Diagnostics.Stopwatch]::StartNew()
        $managerOperationalResponse = Invoke-TestRequest 'GET' "/reports/operations?$range" $null $managerSession
        $watch.Stop()
        Assert-Status $managerOperationalResponse 200 'Manager operational report failed'
        if ($watch.ElapsedMilliseconds -gt 2000) { throw "Operational report exceeded 2 seconds: $($watch.ElapsedMilliseconds)ms" }
        $managerOperational = Get-JSON $managerOperationalResponse
        if ((Get-Body $managerOperationalResponse) -match 'revenue|average_ticket|voided_amount') { throw 'Manager operational report exposed financial aggregates' }
        Assert-Status (Invoke-TestRequest 'GET' "/reports/operations?$range" $null $adminSession) 200 'Administrator operational report failed'
        Assert-Status (Invoke-TestRequest 'GET' "/reports/operations?$range" $null $washOneSession) 403 'Wash staff accessed aggregate operational reports'

        $financialResponse = Invoke-TestRequest 'GET' "/admin/reports/financial?$range" $null $adminSession
        Assert-Status $financialResponse 200 'Administrator financial report failed'
        $financial = Get-JSON $financialResponse
        Assert-Status (Invoke-TestRequest 'GET' "/admin/reports/financial?$range" $null $managerSession) 403 'Manager accessed financial aggregates'
        Assert-Status (Invoke-TestRequest 'GET' "/admin/dashboard?$range" $null $managerSession) 403 'Manager accessed administrator dashboard'
        $dashboard = Get-JSON (Invoke-TestRequest 'GET' "/admin/dashboard?$range" $null $adminSession)
        if ($dashboard.data.operational.summary.orders_received -ne $managerOperational.data.summary.orders_received) { throw 'Dashboard and operational report formulas differ' }

        $dailyReceived = ($managerOperational.data.daily | Measure-Object -Property received -Sum).Sum
        if ($dailyReceived -ne $managerOperational.data.summary.orders_received) { throw 'Daily operational totals duplicate or omit orders' }
        $dailyRevenue = [decimal](($financial.data.daily | ForEach-Object { [decimal]$_.amount } | Measure-Object -Sum).Sum)
        if ($dailyRevenue -ne [decimal]$financial.data.summary.revenue) { throw 'Daily financial totals duplicate or omit valid payments' }
        if ($financial.data.summary.payments -lt 2 -or [decimal]$financial.data.summary.revenue -le 0) { throw 'Financial report did not include valid payments' }
        if ($financial.data.summary.voided_count -lt 2 -or [decimal]$financial.data.summary.voided_amount -le 0) { throw 'Financial report did not audit voided payments separately' }
        if (@($financial.data.methods).Count -ne 4) { throw 'Financial report did not return the four approved payment methods' }

        Assert-Status (Invoke-TestRequest 'GET' '/reports/operations?date_from=20-08-2026&date_to=2026-08-20' $null $managerSession) 422 'Invalid report date was accepted'
        Assert-Status (Invoke-TestRequest 'GET' '/reports/operations?date_from=2026-08-21&date_to=2026-08-20' $null $managerSession) 422 'Inverted report range was accepted'
        Assert-Status (Invoke-TestRequest 'GET' '/reports/operations?date_from=2024-01-01&date_to=2026-01-02' $null $managerSession) 422 'Oversized report range was accepted'
        $reportIndexes = & $psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND indexname IN ('idx_orders_received_at','idx_order_status_history_status_time','idx_payments_status_paid_at','idx_payments_voided_at');"
        if ($reportIndexes.Trim() -ne '4') { throw 'Report query indexes are incomplete' }
        if ($KeepEnvironment) {
            $visualAdmin = Invoke-TestRequest 'POST' '/admin/users' @{ email = 'visual-admin-reports@example.test'; password = 'visual reports admin 2026!'; first_name = 'Ana'; last_name = 'Reportes'; role = 'admin'; status = 'active' } $adminSession $adminHeaders
            Assert-Status $visualAdmin 201 'Visual report administrator fixture failed'
            Write-Output 'VISUAL_REPORT_USERS admin=visual-admin-reports@example.test manager=manager-orders@example.test'
        }
        Write-Output "MODULE7_E2E_OK timezone=America/La_Paz operational=admin+manager financial=admin_only totals=no_duplicates voids=separate methods=4 dates=validated indexes=4 response_ms=$($watch.ElapsedMilliseconds)"
    }

    Write-Output 'E2E_OK migration_twice=ok atomic_create=201 total=40.00 snapshots=name+price mass_assignment=400 missing_price=422 duplicates=422 idor=404 transitions=valid roles=protected concurrency=200/409 history=append_only cancellation=reason+terminal edits=state_limited delivery=payment_blocked histories=client+vehicle'
} finally {
    if ((Get-Location).Path -eq $backendRoot) { Pop-Location }
    if ($KeepEnvironment -and $apiProcess -and -not $apiProcess.HasExited) {
        Write-Output "ENVIRONMENT_KEPT PID=$($apiProcess.Id) DATABASE=$databaseName PORT=$serverPort EXECUTABLE=$apiExecutable"
    } else {
        if ($apiProcess -and -not $apiProcess.HasExited) { $apiProcess.Kill(); $apiProcess.WaitForExit() }
        if ($databaseName -match '^lavadero_m5_e2e_[0-9]{14}$') { & $dropDb -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER --if-exists $databaseName | Out-Null; Write-Output "TEMP_DATABASE_REMOVED=$databaseName" }
        if (Test-Path -LiteralPath $apiExecutable) { Remove-Item -LiteralPath $apiExecutable -Force }
    }
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
