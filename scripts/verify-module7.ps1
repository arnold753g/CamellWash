$ErrorActionPreference = 'Stop'
$scriptPath = Join-Path $PSScriptRoot 'verify-module5.ps1'
& powershell.exe -NoProfile -ExecutionPolicy Bypass -File $scriptPath -VerifyPayments -VerifyReports
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
