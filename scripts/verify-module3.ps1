$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Net.Http

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendRoot = Join-Path $projectRoot 'backend'
$workspaceTemp = Join-Path $projectRoot '.tmp'
New-Item -ItemType Directory -Force -Path $workspaceTemp | Out-Null

Get-Content (Join-Path $projectRoot '.env') | ForEach-Object {
    if ($_ -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
    }
}

$databaseName = 'lavadero_m3_e2e_' + (Get-Date -Format 'yyyyMMddHHmmss')
if ($databaseName -notmatch '^lavadero_m3_e2e_[0-9]{14}$') { throw 'Unsafe temporary database name' }
$serverPort = 18083
if (Get-NetTCPConnection -LocalPort $serverPort -State Listen -ErrorAction SilentlyContinue) { throw "Port $serverPort is already in use" }

$createDbCommand = (Get-Command createdb.exe -ErrorAction Stop).Source
$psqlCommand = (Get-Command psql.exe -ErrorAction Stop).Source
$dropDbCommand = (Get-Command dropdb.exe -ErrorAction Stop).Source
$env:PGPASSWORD = $env:DB_PASSWORD
$env:GOCACHE = Join-Path $workspaceTemp 'go-build'
$apiProcess = $null

function New-WebSession { New-Object Microsoft.PowerShell.Commands.WebRequestSession }
function Invoke-TestRequest($method, $path, $body, $session, $headers = @{}) {
    try {
        $arguments = @{ Uri = "$script:apiBase$path"; Method = $method; WebSession = $session; Headers = $headers; UseBasicParsing = $true }
        if ($null -ne $body) { $arguments.ContentType = 'application/json'; $arguments.Body = $body | ConvertTo-Json -Compress }
        Invoke-WebRequest @arguments
    } catch { $_.Exception.Response }
}
function Get-Status($response) { [int]$response.StatusCode }
function Get-JSON($response) { $response.Content | ConvertFrom-Json }

& $createDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER $databaseName
if ($LASTEXITCODE -ne 0) { throw 'Could not create the temporary E2E database' }

try {
    $env:DB_NAME = $databaseName
    $env:APP_ENV = 'test'
    $env:DB_AUTO_MIGRATE = 'false'
    $env:SERVER_HOST = '127.0.0.1'
    $env:SERVER_PORT = [string]$serverPort
    $env:AUTH_COOKIE_SECURE = 'false'
    $env:SEED_DEVELOPMENT = 'true'
    $env:AUTH_RATE_LIMIT_PER_MINUTE = '100'

    Push-Location $backendRoot
    go run ./cmd/migrate
    if ($LASTEXITCODE -ne 0) { throw 'E2E migration failed' }
    go run ./cmd/migrate
    if ($LASTEXITCODE -ne 0) { throw 'E2E migration was not idempotent' }
    go run ./cmd/seed
    if ($LASTEXITCODE -ne 0) { throw 'E2E administrator seed failed' }
    $apiExecutable = Join-Path $workspaceTemp 'm3-api.exe'
    go build -o $apiExecutable ./cmd/api
    if ($LASTEXITCODE -ne 0) { throw 'E2E API build failed' }
    Pop-Location

    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $apiExecutable
    $startInfo.WorkingDirectory = $backendRoot
    $startInfo.UseShellExecute = $true
    $startInfo.WindowStyle = [System.Diagnostics.ProcessWindowStyle]::Hidden
    $apiProcess = [System.Diagnostics.Process]::Start($startInfo)
    $script:apiBase = "http://127.0.0.1:$serverPort/api/v1"
    $healthy = $false
    for ($attempt = 0; $attempt -lt 20; $attempt++) {
        Start-Sleep -Milliseconds 500
        try {
            if ([int](Invoke-WebRequest -Uri "http://127.0.0.1:$serverPort/health" -UseBasicParsing).StatusCode -eq 200) { $healthy = $true; break }
        } catch {}
    }
    if (-not $healthy) { throw 'E2E API did not become healthy' }

    $origin = @{ Origin = 'http://localhost:3000' }
    $adminSession = New-WebSession
    $adminLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $env:DEV_ADMIN_EMAIL; password = $env:DEV_ADMIN_PASSWORD } $adminSession $origin
    if ((Get-Status $adminLogin) -ne 200) { throw 'Administrator login failed' }
    $adminCSRF = (Get-JSON $adminLogin).data.csrf_token
    $adminHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = $adminCSRF }

    $matrix = Get-JSON (Invoke-TestRequest 'GET' '/admin/service-prices/matrix' $null $adminSession)
    if ($matrix.data.summary.total -ne 36 -or $matrix.data.summary.configured -ne 0 -or $matrix.data.summary.missing -ne 36) {
        throw 'Initial empty matrix did not expose all 36 missing combinations'
    }
    $firstCell = @($matrix.data.cells)[0]
    $secondCell = @($matrix.data.cells)[1]
    $thirdCell = @($matrix.data.cells)[2]

    $created = Invoke-TestRequest 'POST' '/admin/service-prices' @{
        service_id = $firstCell.service_id; vehicle_type_id = $firstCell.vehicle_type_id; price = '9999999999.99'
    } $adminSession $adminHeaders
    if ((Get-Status $created) -ne 201 -or (Get-JSON $created).data.price -ne '9999999999.99') { throw 'Exact decimal price creation failed' }

    $current = Invoke-TestRequest 'GET' "/pricing/current?service_id=$($firstCell.service_id)&vehicle_type_id=$($firstCell.vehicle_type_id)" $null $adminSession
    if ((Get-Status $current) -ne 200 -or (Get-JSON $current).data.price -ne '9999999999.99') { throw 'Current price lost decimal precision' }

    $updated = Invoke-TestRequest 'PATCH' "/admin/service-prices/$($firstCell.service_id)/$($firstCell.vehicle_type_id)" @{ price = '0.1' } $adminSession $adminHeaders
    if ((Get-Status $updated) -ne 200 -or (Get-JSON $updated).data.price -ne '0.10') { throw 'Price update did not preserve canonical precision' }

    foreach ($invalid in @('0', '-1', '1.234', '1e2')) {
        $invalidResult = Invoke-TestRequest 'POST' '/admin/service-prices' @{
            service_id = $secondCell.service_id; vehicle_type_id = $secondCell.vehicle_type_id; price = $invalid
        } $adminSession $adminHeaders
        if ((Get-Status $invalidResult) -ne 422) { throw "Invalid price '$invalid' was not rejected" }
    }

    $managerPassword = 'correct horse manager price 2026!'
    $washPassword = 'correct horse wash price 2026!'
    foreach ($fixture in @(
        @{ email = 'manager-price@example.test'; password = $managerPassword; first_name = 'Mara'; last_name = 'Precio'; role = 'manager'; status = 'active' },
        @{ email = 'wash-price@example.test'; password = $washPassword; first_name = 'Luz'; last_name = 'Precio'; role = 'wash_staff'; status = 'active' }
    )) {
        if ((Get-Status (Invoke-TestRequest 'POST' '/admin/users' $fixture $adminSession $adminHeaders)) -ne 201) { throw 'Role fixture creation failed' }
    }
    $managerSession = New-WebSession
    if ((Get-Status (Invoke-TestRequest 'POST' '/auth/login' @{ email = 'manager-price@example.test'; password = $managerPassword } $managerSession $origin)) -ne 200) { throw 'Manager login failed' }
    $washSession = New-WebSession
    if ((Get-Status (Invoke-TestRequest 'POST' '/auth/login' @{ email = 'wash-price@example.test'; password = $washPassword } $washSession $origin)) -ne 200) { throw 'Wash-staff login failed' }

    if ((Get-Status (Invoke-TestRequest 'GET' "/pricing/current?service_id=$($firstCell.service_id)&vehicle_type_id=$($firstCell.vehicle_type_id)" $null $managerSession)) -ne 200) { throw 'Manager could not consult the current price' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/service-prices' @{ service_id = $secondCell.service_id; vehicle_type_id = $secondCell.vehicle_type_id; price = '20.00' } $managerSession)) -ne 403) { throw 'Manager modified prices' }
    if ((Get-Status (Invoke-TestRequest 'GET' "/pricing/current?service_id=$($firstCell.service_id)&vehicle_type_id=$($firstCell.vehicle_type_id)" $null $washSession)) -ne 403) { throw 'Wash staff accessed pricing information' }

    $cookieName = 'starter_session'
    $adminCookie = @($adminSession.Cookies.GetCookies($script:apiBase) | Where-Object { $_.Name -eq $cookieName })[0]
    if (-not $adminCookie) { throw 'Administrator cookie was not available for concurrency test' }
    $cookieContainer = New-Object System.Net.CookieContainer
    $concurrencyCookie = New-Object System.Net.Cookie($cookieName, $adminCookie.Value, '/api/v1', '127.0.0.1')
    $cookieContainer.Add($concurrencyCookie)
    $httpHandler = New-Object System.Net.Http.HttpClientHandler
    $httpHandler.CookieContainer = $cookieContainer
    $httpClient = New-Object System.Net.Http.HttpClient($httpHandler)
    try {
        $requests = @()
        foreach ($price in @('31.00', '32.00')) {
            $request = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Post, "$script:apiBase/admin/service-prices")
            $request.Headers.Add('Origin', 'http://localhost:3000')
            $request.Headers.Add('X-CSRF-Token', $adminCSRF)
            $body = @{ service_id = $thirdCell.service_id; vehicle_type_id = $thirdCell.vehicle_type_id; price = $price } | ConvertTo-Json -Compress
            $request.Content = New-Object System.Net.Http.StringContent($body, [Text.Encoding]::UTF8, 'application/json')
            $requests += $httpClient.SendAsync($request)
        }
        [System.Threading.Tasks.Task]::WaitAll([System.Threading.Tasks.Task[]]$requests)
        $concurrentStatuses = @($requests | ForEach-Object { [int]$_.Result.StatusCode } | Sort-Object)
        if (($concurrentStatuses -join ',') -ne '201,409') { throw "Concurrent duplicate results were $($concurrentStatuses -join ',')" }
    } finally { $httpClient.Dispose(); $httpHandler.Dispose() }

    $newService = Invoke-TestRequest 'POST' '/admin/services' @{ category_id = $firstCell.category_id; name = 'Servicio matriz temporal' } $adminSession $adminHeaders
    if ((Get-Status $newService) -ne 201) { throw 'Could not create a service for missing-cell test' }
    $newServiceID = (Get-JSON $newService).data.id
    $expanded = Get-JSON (Invoke-TestRequest 'GET' "/admin/service-prices/matrix?service_id=$newServiceID" $null $adminSession)
    if ($expanded.data.summary.total -ne 6 -or $expanded.data.summary.missing -ne 6) { throw 'New service did not immediately expose missing combinations' }

    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/services/$newServiceID/status" @{ active = $false } $adminSession $adminHeaders)) -ne 200) { throw 'Service deactivation failed' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/service-prices' @{ service_id = $newServiceID; vehicle_type_id = $firstCell.vehicle_type_id; price = '25.00' } $adminSession $adminHeaders)) -ne 422) { throw 'A price was created for an inactive service' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/services/$newServiceID/status" @{ active = $true } $adminSession $adminHeaders)) -ne 200) { throw 'Service reactivation failed' }
    $reactivated = Get-JSON (Invoke-TestRequest 'GET' "/admin/service-prices/matrix?service_id=$newServiceID" $null $adminSession)
    if ($reactivated.data.summary.missing -ne 6) { throw 'Reactivated service did not restore its missing combinations' }

    $duplicateCount = & $psqlCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM service_prices WHERE service_id = $($thirdCell.service_id) AND vehicle_type_id = $($thirdCell.vehicle_type_id);"
    if ($LASTEXITCODE -ne 0 -or [int]$duplicateCount -ne 1) { throw 'Concurrent writes created duplicate rows' }

    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$($firstCell.vehicle_type_id)/status" @{ active = $false } $adminSession $adminHeaders)) -ne 200) { throw 'Vehicle type deactivation failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' "/pricing/current?service_id=$($firstCell.service_id)&vehicle_type_id=$($firstCell.vehicle_type_id)" $null $managerSession)) -ne 404) { throw 'Inactive combination remained current' }
    $preservedPrice = & $psqlCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -v ON_ERROR_STOP=1 -c "SELECT price::TEXT FROM service_prices WHERE service_id = $($firstCell.service_id) AND vehicle_type_id = $($firstCell.vehicle_type_id);"
    if ($LASTEXITCODE -ne 0 -or $preservedPrice.Trim() -ne '0.10') { throw 'Inactive combination lost its stored price' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$($firstCell.vehicle_type_id)/status" @{ active = $true } $adminSession $adminHeaders)) -ne 200) { throw 'Vehicle type reactivation failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' "/pricing/current?service_id=$($firstCell.service_id)&vehicle_type_id=$($firstCell.vehicle_type_id)" $null $managerSession)) -ne 200) { throw 'Reactivated combination did not recover its preserved current price' }

    Write-Output 'E2E_OK migration_twice=ok initial_matrix=0/36 create=201 update=200 precision=exact invalid=422 manager_read=200 manager_write=403 wash_read=403 concurrent=201/409 missing_dynamic=ok inactive_hidden=ok history=preserved'
} finally {
    if ((Get-Location).Path -eq $backendRoot) { Pop-Location }
    if ($apiProcess -and -not $apiProcess.HasExited) { $apiProcess.Kill(); $apiProcess.WaitForExit() }
    if ($databaseName -match '^lavadero_m3_e2e_[0-9]{14}$') {
        & $dropDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER --if-exists $databaseName | Out-Null
        Write-Output "TEMP_DATABASE_REMOVED=$databaseName"
    }
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
