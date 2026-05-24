param(
    [Parameter(Mandatory = $true)]
    [string]$Address,

    [string]$DataDir = ".saracie",
    [string]$Peers = "",
    [switch]$Foreground
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$MinerExe = Join-Path $Root "bin\saracie-miner.exe"

if (-not (Test-Path $MinerExe)) {
    throw "Missing $MinerExe. Run .\scripts\build-release.ps1 or go build first."
}

if (-not $Address.StartsWith("sar1")) {
    throw "Mining address must start with sar1."
}

$argsList = @("--datadir", $DataDir, "--address", $Address)
if ($Peers -ne "") {
    $argsList += @("--peers", $Peers)
}

if ($Foreground) {
    & $MinerExe @argsList
    exit $LASTEXITCODE
}

$process = Start-Process -FilePath $MinerExe -ArgumentList $argsList -WindowStyle Hidden -PassThru
Write-Host "Saracie miner started."
Write-Host "PID: $($process.Id)"
Write-Host "Payout: $Address"
if ($Peers -ne "") {
    Write-Host "Peers: $Peers"
}
