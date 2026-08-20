$ErrorActionPreference = 'Stop'

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendRoot = Join-Path $projectRoot 'backend'
$workspaceTemp = Join-Path $projectRoot '.tmp'
New-Item -ItemType Directory -Force -Path $workspaceTemp | Out-Null

Get-Content (Join-Path $projectRoot '.env') | ForEach-Object {
    if ($_ -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
    }
}

$databaseName = 'lavadero_m2_e2e_' + (Get-Date -Format 'yyyyMMddHHmmss')
if ($databaseName -notmatch '^lavadero_m2_e2e_[0-9]{14}$') { throw 'Unsafe temporary database name' }
$serverPort = 18082
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
    $apiExecutable = Join-Path $workspaceTemp 'm2-api.exe'
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
    $adminHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $adminLogin).data.csrf_token }

    $initialTypes = Get-JSON (Invoke-TestRequest 'GET' '/admin/vehicle-types?active=all&limit=100' $null $adminSession)
    $initialCategories = Get-JSON (Invoke-TestRequest 'GET' '/admin/service-categories?active=all&limit=100' $null $adminSession)
    $initialServices = Get-JSON (Invoke-TestRequest 'GET' '/admin/services?active=all&limit=100' $null $adminSession)
    if (@($initialTypes.data.vehicle_types).Count -ne 6) { throw 'Initial vehicle types were not loaded exactly once' }
    if (@($initialCategories.data.service_categories).Count -ne 2) { throw 'Initial categories were not loaded exactly once' }
    if (@($initialServices.data.services).Count -ne 6) { throw 'Initial services were not loaded exactly once' }
    $moto = @($initialTypes.data.vehicle_types | Where-Object { $_.name -eq 'Moto' })[0]
    if (-not $moto) { throw 'Initial Moto vehicle type is missing' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$($moto.id)" @{ name = 'Moto editada' } $adminSession $adminHeaders)) -ne 200) { throw 'An initial value could not be modified' }

    $managerPassword = 'correct horse manager 2026!'
    $washPassword = 'correct horse wash 2026!'
    foreach ($fixture in @(
        @{ email = 'manager-catalog@example.test'; password = $managerPassword; first_name = 'Mara'; last_name = 'Catálogo'; role = 'manager'; status = 'active' },
        @{ email = 'wash-catalog@example.test'; password = $washPassword; first_name = 'Luz'; last_name = 'Catálogo'; role = 'wash_staff'; status = 'active' }
    )) {
        if ((Get-Status (Invoke-TestRequest 'POST' '/admin/users' $fixture $adminSession $adminHeaders)) -ne 201) { throw 'Role fixture creation failed' }
    }

    $vehicle = Invoke-TestRequest 'POST' '/admin/vehicle-types' @{ name = 'Vehículo de prueba' } $adminSession $adminHeaders
    if ((Get-Status $vehicle) -ne 201) { throw 'Vehicle type creation failed' }
    $vehicleID = (Get-JSON $vehicle).data.id
    $vehicleDuplicate = Invoke-TestRequest 'POST' '/admin/vehicle-types' @{ name = '  VEHÍCULO   DE PRUEBA  ' } $adminSession $adminHeaders
    if ((Get-Status $vehicleDuplicate) -ne 409) { throw 'Normalized vehicle type duplicate was not rejected' }

    $categoryA = Invoke-TestRequest 'POST' '/admin/service-categories' @{ name = 'Categoría A' } $adminSession $adminHeaders
    $categoryB = Invoke-TestRequest 'POST' '/admin/service-categories' @{ name = 'Categoría B' } $adminSession $adminHeaders
    if ((Get-Status $categoryA) -ne 201 -or (Get-Status $categoryB) -ne 201) { throw 'Category creation failed' }
    $categoryAID = (Get-JSON $categoryA).data.id
    $categoryBID = (Get-JSON $categoryB).data.id
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/service-categories' @{ name = ' categoría   a ' } $adminSession $adminHeaders)) -ne 409) { throw 'Normalized category duplicate was not rejected' }

    $serviceA = Invoke-TestRequest 'POST' '/admin/services' @{ category_id = $categoryAID; name = 'Servicio de prueba' } $adminSession $adminHeaders
    if ((Get-Status $serviceA) -ne 201) { throw 'Service creation failed' }
    $serviceAID = (Get-JSON $serviceA).data.id
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/services' @{ category_id = $categoryAID; name = ' SERVICIO   DE PRUEBA ' } $adminSession $adminHeaders)) -ne 409) { throw 'Same-category normalized service duplicate was not rejected' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/services' @{ category_id = $categoryBID; name = 'Servicio de prueba' } $adminSession $adminHeaders)) -ne 201) { throw 'Same service name in another category was rejected' }

    $managerSession = New-WebSession
    if ((Get-Status (Invoke-TestRequest 'POST' '/auth/login' @{ email = 'manager-catalog@example.test'; password = $managerPassword } $managerSession $origin)) -ne 200) { throw 'Manager login failed' }
    $washSession = New-WebSession
    if ((Get-Status (Invoke-TestRequest 'POST' '/auth/login' @{ email = 'wash-catalog@example.test'; password = $washPassword } $washSession $origin)) -ne 200) { throw 'Wash-staff login failed' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/vehicle-types' @{ name = 'No permitido' } $managerSession)) -ne 403) { throw 'Manager modified catalogs' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/services/$serviceAID/status" @{ active = $false } $washSession)) -ne 403) { throw 'Wash staff modified catalogs' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/catalogs/vehicle-types' $null $managerSession)) -ne 200) { throw 'Manager could not read active catalogs' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/catalogs/services' $null $washSession)) -ne 200) { throw 'Wash staff could not read active catalogs' }

    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$vehicleID/status" @{ active = $false } $adminSession $adminHeaders)) -ne 200) { throw 'Vehicle type deactivation failed' }
    $activeVehicles = Get-JSON (Invoke-TestRequest 'GET' '/catalogs/vehicle-types' $null $managerSession)
    if (@($activeVehicles.data.vehicle_types | Where-Object { $_.id -eq $vehicleID }).Count -ne 0) { throw 'Inactive vehicle type remained operational' }
    $adminVehicles = Get-JSON (Invoke-TestRequest 'GET' '/admin/vehicle-types?active=false' $null $adminSession)
    if (@($adminVehicles.data.vehicle_types | Where-Object { $_.id -eq $vehicleID }).Count -ne 1) { throw 'Inactive vehicle type disappeared from administration' }

    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/service-categories/$categoryAID/status" @{ active = $false } $adminSession $adminHeaders)) -ne 200) { throw 'Category deactivation failed' }
    $activeServices = Get-JSON (Invoke-TestRequest 'GET' '/catalogs/services' $null $washSession)
    if (@($activeServices.data.services | Where-Object { $_.id -eq $serviceAID }).Count -ne 0) { throw 'Service from inactive category remained operational' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/services' @{ category_id = $categoryAID; name = 'No disponible' } $adminSession $adminHeaders)) -ne 422) { throw 'Service was created in an inactive category' }

    $persisted = & $psqlCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM vehicle_types WHERE id = $vehicleID;"
    if ($LASTEXITCODE -ne 0 -or [int]$persisted -ne 1) { throw 'Soft deactivation removed the vehicle type' }

    Write-Output 'E2E_OK migration_twice=ok initial_values=6/2/6 initial_edit=200 vehicle_duplicate=409 category_duplicate=409 service_duplicate=409 cross_category=201 manager_write=403 wash_write=403 active_reads=200 soft_delete=preserved inactive_hidden=ok'
} finally {
    if ((Get-Location).Path -eq $backendRoot) { Pop-Location }
    if ($apiProcess -and -not $apiProcess.HasExited) { $apiProcess.Kill(); $apiProcess.WaitForExit() }
    if ($databaseName -match '^lavadero_m2_e2e_[0-9]{14}$') {
        & $dropDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER --if-exists $databaseName | Out-Null
        Write-Output "TEMP_DATABASE_REMOVED=$databaseName"
    }
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
