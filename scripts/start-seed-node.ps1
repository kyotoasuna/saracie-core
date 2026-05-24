param(
    [string]$DataDir = ".saracie",
    [string]$Listen = "0.0.0.0:7339",
    [string]$Self = "",
    [string]$Peers = "",
    [switch]$Foreground
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$NodeExe = Join-Path $Root "bin\saracied.exe"

if (-not (Test-Path $NodeExe)) {
    throw "Missing $NodeExe. Run .\scripts\build-release.ps1 or go build first."
}

$argsList = @("node", "--datadir", $DataDir, "--listen", $Listen)
if ($Self -ne "") {
    $argsList += @("--self", $Self)
}
if ($Peers -ne "") {
    $argsList += @("--peers", $Peers)
}

if ($Foreground) {
    & $NodeExe @argsList
    exit $LASTEXITCODE
}

$process = Start-Process -FilePath $NodeExe -ArgumentList $argsList -WindowStyle Hidden -PassThru
Write-Host "Saracie seed node started."
Write-Host "PID: $($process.Id)"
Write-Host "Listen: $Listen"
if ($Self -ne "") {
    Write-Host "Public URL: $Self"
}
