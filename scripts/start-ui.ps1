param(
    [string]$DataDir = ".saracie-ui",
    [string]$Listen = "127.0.0.1:7340",
    [string]$Peers = "",
    [switch]$Foreground
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$UIExe = Join-Path $Root "bin\saracie-ui.exe"
if (-not (Test-Path $UIExe)) {
    $UIExe = Join-Path $Root "saracie-ui.exe"
}

if (-not (Test-Path $UIExe)) {
    throw "Missing $UIExe. Run .\scripts\build-release.ps1 or go build first."
}

$argsList = @("--datadir", $DataDir, "--listen", $Listen)
if ($Peers -ne "") {
    $argsList += @("--peers", $Peers)
}

if ($Foreground) {
    & $UIExe @argsList
    exit $LASTEXITCODE
}

$process = Start-Process -FilePath $UIExe -ArgumentList $argsList -WindowStyle Hidden -PassThru
Write-Host "Saracie UI started."
Write-Host "PID: $($process.Id)"
Write-Host "URL: http://$Listen"
