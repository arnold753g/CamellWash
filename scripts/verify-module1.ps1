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

$databaseName = 'lavadero_m1_e2e_' + (Get-Date -Format 'yyyyMMddHHmmss')
if ($databaseName -notmatch '^lavadero_m1_e2e_[0-9]{14}$') {
    throw 'Unsafe temporary database name'
}

$serverPort = 18081
if (Get-NetTCPConnection -LocalPort $serverPort -State Listen -ErrorAction SilentlyContinue) {
    throw "Port $serverPort is already in use"
}

$createDbCommand = (Get-Command createdb.exe -ErrorAction Stop).Source
$psqlCommand = (Get-Command psql.exe -ErrorAction Stop).Source
$dropDbCommand = (Get-Command dropdb.exe -ErrorAction Stop).Source
$env:PGPASSWORD = $env:DB_PASSWORD
$env:GOCACHE = Join-Path $workspaceTemp 'go-build'
$apiProcess = $null

function New-WebSession {
    New-Object Microsoft.PowerShell.Commands.WebRequestSession
}

function Invoke-TestRequest($method, $path, $body, $session, $headers = @{}) {
    try {
        $arguments = @{
            Uri = "$script:apiBase$path"
            Method = $method
            WebSession = $session
            Headers = $headers
            UseBasicParsing = $true
        }
        if ($null -ne $body) {
            $arguments.ContentType = 'application/json'
            $arguments.Body = $body | ConvertTo-Json -Compress
        }
        Invoke-WebRequest @arguments
    } catch {
        $_.Exception.Response
    }
}

function Get-Status($response) {
    [int]$response.StatusCode
}

function Get-JSON($response) {
    $response.Content | ConvertFrom-Json
}

& $createDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER $databaseName
if ($LASTEXITCODE -ne 0) {
    throw 'Could not create the temporary E2E database'
}

try {
    $env:DB_NAME = $databaseName
    $env:APP_ENV = 'test'
    $env:DB_AUTO_MIGRATE = 'false'
    $env:SERVER_HOST = '127.0.0.1'
    $env:SERVER_PORT = [string]$serverPort
    $env:AUTH_COOKIE_SECURE = 'false'
    $env:SEED_DEVELOPMENT = 'true'
    $env:AUTH_RATE_LIMIT_PER_MINUTE = '100'
    $env:LOGIN_MAX_ATTEMPTS = '5'

    Push-Location $backendRoot
    go run ./cmd/migrate
    if ($LASTEXITCODE -ne 0) { throw 'E2E migration failed' }
    go run ./cmd/seed
    if ($LASTEXITCODE -ne 0) { throw 'E2E administrator seed failed' }
    $apiExecutable = Join-Path $workspaceTemp 'm1-api.exe'
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
            $health = Invoke-WebRequest -Uri "http://127.0.0.1:$serverPort/health" -UseBasicParsing
            if ([int]$health.StatusCode -eq 200) {
                $healthy = $true
                break
            }
        } catch {}
    }
    if (-not $healthy) { throw 'E2E API did not become healthy' }

    $origin = @{ Origin = 'http://localhost:3000' }
    $registration = Invoke-TestRequest 'POST' '/auth/register' @{ email = 'x@example.test' } (New-WebSession) $origin
    if ((Get-Status $registration) -ne 404) { throw 'Public registration route still exists' }

    $badLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $env:DEV_ADMIN_EMAIL; password = 'wrong password value' } (New-WebSession) $origin
    if ((Get-Status $badLogin) -ne 401) { throw 'Incorrect login did not return 401' }

    $adminSession = New-WebSession
    $adminLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $env:DEV_ADMIN_EMAIL; password = $env:DEV_ADMIN_PASSWORD } $adminSession $origin
    if ((Get-Status $adminLogin) -ne 200) { throw 'Administrator login failed' }
    $csrfToken = (Get-JSON $adminLogin).data.csrf_token
    $adminHeaders = @{ Origin = 'http://localhost:3000'; 'X-CSRF-Token' = $csrfToken }

    $managerBody = @{
        email = 'manager@example.test'
        password = 'correct horse battery 2026!'
        first_name = 'Mara'
        last_name = 'Encargada'
        role = 'manager'
        status = 'active'
    }
    $managerCreated = Invoke-TestRequest 'POST' '/admin/users' $managerBody $adminSession $adminHeaders
    if ((Get-Status $managerCreated) -ne 201) { throw 'Manager creation failed' }
    if ($managerCreated.Content -match 'password|hash|correct horse') { throw 'Create response leaked password material' }
    $managerID = (Get-JSON $managerCreated).data.id

    $managerUpdated = Invoke-TestRequest 'PATCH' "/admin/users/$managerID" @{
        email = $managerBody.email
        first_name = 'María'
        last_name = 'Encargada'
    } $adminSession $adminHeaders
    if ((Get-Status $managerUpdated) -ne 200) { throw 'Manager profile update failed' }

    $duplicateBody = $managerBody.Clone()
    $duplicateBody.email = ' MANAGER@EXAMPLE.TEST '
    $duplicateBody.role = 'wash_staff'
    $duplicate = Invoke-TestRequest 'POST' '/admin/users' $duplicateBody $adminSession $adminHeaders
    if ((Get-Status $duplicate) -ne 409) { throw 'Normalized duplicate email was not rejected' }

    $washBody = @{
        email = 'wash@example.test'
        password = 'correct horse battery 2026!'
        first_name = 'Luz'
        last_name = 'Lavado'
        role = 'wash_staff'
        status = 'active'
    }
    $washCreated = Invoke-TestRequest 'POST' '/admin/users' $washBody $adminSession $adminHeaders
    if ((Get-Status $washCreated) -ne 201) { throw 'Wash-staff creation failed' }

    $managerSession = New-WebSession
    $managerLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $managerBody.email; password = $managerBody.password } $managerSession $origin
    if ((Get-Status $managerLogin) -ne 200) { throw 'Manager login failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/admin/users' $null $managerSession)) -ne 403) { throw 'Manager reached the admin API' }

    $washSession = New-WebSession
    $washLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $washBody.email; password = $washBody.password } $washSession $origin
    if ((Get-Status $washLogin) -ne 200) { throw 'Wash-staff login failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/admin/users' $null $washSession)) -ne 403) { throw 'Wash staff reached the admin API' }
    & $psqlCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $databaseName -v ON_ERROR_STOP=1 -c "UPDATE auth_sessions SET expires_at = NOW() - INTERVAL '1 minute' WHERE user_id = (SELECT id FROM users WHERE email = 'wash@example.test');" | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Could not expire the wash-staff session fixture' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/me' $null $washSession)) -ne 401) { throw 'Expired wash-staff session retained access' }

    $deactivated = Invoke-TestRequest 'PATCH' "/admin/users/$managerID/status" @{ status = 'inactive' } $adminSession $adminHeaders
    if ((Get-Status $deactivated) -ne 200) { throw 'Manager deactivation failed' }
    if ((Get-Status (Invoke-TestRequest 'GET' '/me' $null $managerSession)) -ne 401) { throw 'Deactivated manager retained an active session' }
    $inactiveLogin = Invoke-TestRequest 'POST' '/auth/login' @{ email = $managerBody.email; password = $managerBody.password } (New-WebSession) $origin
    if ((Get-Status $inactiveLogin) -ne 403) { throw 'Inactive manager could still log in' }

    $lockBody = @{
        email = 'lock@example.test'
        password = 'correct horse battery 2026!'
        first_name = 'Cuenta'
        last_name = 'Bloqueo'
        role = 'wash_staff'
        status = 'active'
    }
    if ((Get-Status (Invoke-TestRequest 'POST' '/admin/users' $lockBody $adminSession $adminHeaders)) -ne 201) { throw 'Lockout fixture creation failed' }
    for ($attempt = 0; $attempt -lt 5; $attempt++) {
        $failed = Invoke-TestRequest 'POST' '/auth/login' @{ email = $lockBody.email; password = 'incorrect password value' } (New-WebSession) $origin
        if ((Get-Status $failed) -ne 401) { throw 'Failed login did not return 401 before lockout' }
    }
    $locked = Invoke-TestRequest 'POST' '/auth/login' @{ email = $lockBody.email; password = $lockBody.password } (New-WebSession) $origin
    if ((Get-Status $locked) -ne 429) { throw 'Login lockout was not enforced' }

    Write-Output 'E2E_OK registration=404 bad_login=401 admin_create=201 admin_edit=200 duplicate=409 manager_admin=403 wash_admin=403 expired=401 revoked=401 inactive_login=403 lockout=429'
} finally {
    if ((Get-Location).Path -eq $backendRoot) { Pop-Location }
    if ($apiProcess -and -not $apiProcess.HasExited) {
        $apiProcess.Kill()
        $apiProcess.WaitForExit()
    }
    if ($databaseName -match '^lavadero_m1_e2e_[0-9]{14}$') {
        & $dropDbCommand -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER --if-exists $databaseName | Out-Null
        Write-Output "TEMP_DATABASE_REMOVED=$databaseName"
    }
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
