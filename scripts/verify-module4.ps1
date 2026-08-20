$ErrorActionPreference = 'Stop'

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendRoot = Join-Path $projectRoot 'backend'
$workspaceTemp = Join-Path $projectRoot '.tmp'
New-Item -ItemType Directory -Force -Path $workspaceTemp | Out-Null

Get-Content (Join-Path $projectRoot '.env') | ForEach-Object {
    if ($_ -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') { [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process') }
}

$databaseName = 'lavadero_m4_e2e_' + (Get-Date -Format 'yyyyMMddHHmmss')
if ($databaseName -notmatch '^lavadero_m4_e2e_[0-9]{14}$') { throw 'Unsafe temporary database name' }
$serverPort = 18084
if (Get-NetTCPConnection -LocalPort $serverPort -State Listen -ErrorAction SilentlyContinue) { throw "Port $serverPort is already in use" }

$createDbCommand = (Get-Command createdb.exe -ErrorAction Stop).Source
$psqlCommand = (Get-Command psql.exe -ErrorAction Stop).Source
$dropDbCommand = (Get-Command dropdb.exe -ErrorAction Stop).Source
$env:PGPASSWORD = $env:DB_PASSWORD
$env:GOCACHE = Join-Path $workspaceTemp 'go-build'
$apiProcess = $null
$apiExecutable = Join-Path $workspaceTemp 'm4-api.exe'

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
        try { if ([int](Invoke-WebRequest -Uri "http://127.0.0.1:$serverPort/health" -UseBasicParsing).StatusCode -eq 200) { $healthy = $true; break } } catch {}
    }
    if (-not $healthy) { throw 'E2E API did not become healthy' }

    $origin = @{ Origin = 'http://localhost:3000' }
    $adminSession = New-WebSession
    $adminLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $env:DEV_ADMIN_EMAIL; password = $env:DEV_ADMIN_PASSWORD } $adminSession $origin
    if ((Get-Status $adminLogin) -ne 200) { throw 'Administrator login failed' }
    $adminHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $adminLogin).data.csrf_token }

    $missingCI = Invoke-TestRequest 'POST' '/clients' @{ full_name = 'Cliente Sin Documento'; ci = $null; phone = '+591 700-10001' } $adminSession $adminHeaders
    if ((Get-Status $missingCI) -ne 201) { throw 'Client without CI was rejected' }
    $missingData = (Get-JSON $missingCI).data
    if ($missingData.ci -ne '00000000' -or $missingData.ci_provided -ne $false) { throw 'Missing CI representation is incorrect' }
    $clientID = $missingData.id
    if ((Get-Status (Invoke-TestRequest 'POST' '/clients' @{ full_name = 'Segundo Sin Documento'; ci = ''; phone = '70010002' } $adminSession $adminHeaders)) -ne 201) { throw 'A second client without CI was rejected' }

    $realCI = Invoke-TestRequest 'POST' '/clients' @{ full_name = 'Cliente Con Documento'; ci = '12-34 LP'; phone = '70010003' } $adminSession $adminHeaders
    if ((Get-Status $realCI) -ne 201) { throw 'Client with real CI was not created' }
    $realClientID = (Get-JSON $realCI).data.id
    if ((Get-Status (Invoke-TestRequest 'POST' '/clients' @{ full_name = 'CI Repetido'; ci = '1234lp'; phone = '70010004' } $adminSession $adminHeaders)) -ne 409) { throw 'Normalized duplicate CI was accepted' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/clients' @{ full_name = 'CI Reservado'; ci = '00000000'; phone = '70010005' } $adminSession $adminHeaders)) -ne 422) { throw 'Reserved CI placeholder was accepted as a real CI' }
    if ((Get-Status (Invoke-TestRequest 'POST' '/clients' @{ full_name = 'A'; ci = $null; phone = '123' } $adminSession $adminHeaders)) -ne 422) { throw 'Invalid client fields were accepted' }

    $clientList = Get-JSON (Invoke-TestRequest 'GET' '/clients?search=Cliente&limit=20' $null $adminSession)
    $listedReal = @($clientList.data.clients | Where-Object { $_.id -eq $realClientID })[0]
    if (-not $listedReal -or $listedReal.ci -eq '12-34 LP' -or $listedReal.phone -eq '70010003') { throw 'Client list exposed full personal identifiers' }
    $clientDetail = Get-JSON (Invoke-TestRequest 'GET' "/clients/$realClientID" $null $adminSession)
    if ($clientDetail.data.ci -ne '12-34 LP' -or $clientDetail.data.phone -ne '70010003') { throw 'Authorized client detail did not expose the full record' }
    $escapedSearch = Get-JSON (Invoke-TestRequest 'GET' '/clients?search=%25' $null $adminSession)
    if ($escapedSearch.data.pagination.total -ne 0) { throw 'LIKE wildcard was not escaped in client search' }

    $catalog = Get-JSON (Invoke-TestRequest 'GET' '/catalogs/vehicle-types' $null $adminSession)
    $vehicleTypeID = @($catalog.data.vehicle_types)[0].id
    if (-not $vehicleTypeID) { throw 'No active vehicle type was available' }
    $vehicleOne = Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = $null; brand = 'Competición'; model = 'Prototipo'; color = 'Negro' } $adminSession $adminHeaders
    if ((Get-Status $vehicleOne) -ne 201) { throw 'Vehicle without plate was rejected' }
    $vehicleID = (Get-JSON $vehicleOne).data.id
    if ((Get-Status (Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = ''; brand = ''; model = ''; color = '' } $adminSession $adminHeaders)) -ne 201) { throw 'A second vehicle without plate was rejected' }
    if ((Get-Status (Invoke-TestRequest 'POST' "/clients/$realClientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = '1234-ABC'; brand = ''; model = ''; color = '' } $adminSession $adminHeaders)) -ne 201) { throw 'Vehicle with plate was not created' }
    if ((Get-Status (Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = '1234abc'; brand = ''; model = ''; color = '' } $adminSession $adminHeaders)) -ne 409) { throw 'Normalized duplicate plate was accepted' }

    $ownerChange = Invoke-TestRequest 'PATCH' "/vehicles/$vehicleID" @{ client_id = $realClientID; vehicle_type_id = $vehicleTypeID; plate = $null; brand = 'Cambio'; model = ''; color = '' } $adminSession $adminHeaders
    if ((Get-Status $ownerChange) -ne 400) { throw 'Vehicle owner change was not rejected by the API contract' }
    $vehicleAfter = Get-JSON (Invoke-TestRequest 'GET' "/vehicles/$vehicleID" $null $adminSession)
    if ($vehicleAfter.data.client_id -ne $clientID) { throw 'Vehicle owner changed unexpectedly' }

    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$vehicleTypeID/status" @{ active = $false } $adminSession $adminHeaders)) -ne 200) { throw 'Vehicle type deactivation failed' }
    if ((Get-Status (Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = $null; brand = ''; model = ''; color = '' } $adminSession $adminHeaders)) -ne 422) { throw 'New vehicle accepted an inactive type' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/vehicles/$vehicleID" @{ vehicle_type_id = $vehicleTypeID; plate = $null; brand = 'Competición'; model = 'Actualizado'; color = 'Negro' } $adminSession $adminHeaders)) -ne 200) { throw 'Existing vehicle could not retain its inactive historical type' }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/admin/vehicle-types/$vehicleTypeID/status" @{ active = $true } $adminSession $adminHeaders)) -ne 200) { throw 'Vehicle type reactivation failed' }

    $managerPassword = 'correct horse manager clients 2026!'
    $washPassword = 'correct horse wash clients 2026!'
    foreach ($fixture in @(
        @{ email = 'manager-clients@example.test'; password = $managerPassword; first_name = 'Mara'; last_name = 'Clientes'; role = 'manager'; status = 'active' },
        @{ email = 'wash-clients@example.test'; password = $washPassword; first_name = 'Luz'; last_name = 'Clientes'; role = 'wash_staff'; status = 'active' }
    )) {
        if ((Get-Status (Invoke-TestRequest 'POST' '/admin/users' $fixture $adminSession $adminHeaders)) -ne 201) { throw 'Role fixture creation failed' }
    }
    $managerSession = New-WebSession
    $managerLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = 'manager-clients@example.test'; password = $managerPassword } $managerSession $origin
    if ((Get-Status $managerLogin) -ne 200) { throw 'Manager login failed' }
    $managerHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = (Get-JSON $managerLogin).data.csrf_token }
    if ((Get-Status (Invoke-TestRequest 'PATCH' "/clients/$clientID" @{ full_name = 'Cliente Actualizado'; ci = $null; phone = '70010009' } $managerSession $managerHeaders)) -ne 200) { throw 'Manager could not update a client' }
    if ((Get-Status (Invoke-TestRequest 'POST' "/clients/$clientID/vehicles" @{ vehicle_type_id = $vehicleTypeID; plate = $null; brand = ''; model = ''; color = '' } $managerSession $managerHeaders)) -ne 201) { throw 'Manager could not add a vehicle' }
    $washSession = New-WebSession
    if ((Get-Status (Invoke-TestRequest 'POST' '/auth/login' @{ email = 'wash-clients@example.test'; password = $washPassword } $washSession $origin)) -ne 200) { throw 'Wash-staff login failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/clients' $null $washSession)) -ne 403) { throw 'Wash staff accessed clients' }
    $deleteStatus = Get-Status (Invoke-TestRequest 'DELETE' "/vehicles/$vehicleID" $null $adminSession $adminHeaders)
    if ($deleteStatus -notin @(404, 405)) { throw 'A physical vehicle delete endpoint is exposed' }

    $nullCounts = & $psqlCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -X -t -A -v ON_ERROR_STOP=1 -c "SELECT (SELECT COUNT(*) FROM clients WHERE ci IS NULL)::TEXT || '/' || (SELECT COUNT(*) FROM vehicles WHERE plate IS NULL)::TEXT;"
    if ($LASTEXITCODE -ne 0 -or $nullCounts.Trim() -ne '2/3') { throw "Missing identifiers were not stored as SQL NULL: $nullCounts" }
    Write-Output "E2E_OK migration_twice=ok ci_null=2 ci_display=00000000 ci_reserved=422 ci_unique=409 privacy=list_masked detail_full vehicle_plate_null=3 plate_unique=409 owner_change=400 manager_write=200 wash_access=403 delete=$deleteStatus"
} finally {
    if ((Get-Location).Path -eq $backendRoot) { Pop-Location }
    if ($apiProcess -and -not $apiProcess.HasExited) { $apiProcess.Kill(); $apiProcess.WaitForExit() }
    if ($databaseName -match '^lavadero_m4_e2e_[0-9]{14}$') {
        & $dropDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER --if-exists $databaseName | Out-Null
        Write-Output "TEMP_DATABASE_REMOVED=$databaseName"
    }
    if (Test-Path -LiteralPath $apiExecutable) { Remove-Item -LiteralPath $apiExecutable -Force }
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
