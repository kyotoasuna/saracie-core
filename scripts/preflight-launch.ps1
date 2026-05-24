param(
    [string]$DataDir = ".saracie",
    [int]$NodePort = 7339,
    [int]$UIPort = 7340
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Push-Location $Root
try {
    & (Join-Path $PSScriptRoot "preflight-public.ps1")

    $required = @(
        "bin\saracied.exe",
        "bin\saracie-miner.exe",
        "bin\saracie-wallet.exe",
        "bin\saracie-ui.exe"
    )

    foreach ($file in $required) {
        if (-not (Test-Path (Join-Path $Root $file))) {
            throw "Missing $file. Build binaries first."
        }
    }

    $ports = @($NodePort, $UIPort)
    foreach ($port in $ports) {
        $listening = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
        if ($listening) {
            throw "Port $port is already in use. Stop old Saracie processes first."
        }
    }

    $params = & .\bin\saracied.exe params | ConvertFrom-Json
    if ($params.ticker -ne "SRCE") {
        throw "Unexpected ticker: $($params.ticker)"
    }
    if ($params.max_supply_base_units -ne 21000000000000) {
        throw "Unexpected max supply: $($params.max_supply_base_units)"
    }
    if ($params.address_hrp -ne "sar") {
        throw "Unexpected address HRP: $($params.address_hrp)"
    }

    $genesis = & .\bin\saracied.exe genesis | ConvertFrom-Json
    if (-not $genesis.hash) {
        throw "Genesis hash missing."
    }

    Write-Host "Launch preflight passed." -ForegroundColor Green
    Write-Host "Genesis: $($genesis.hash)"
    Write-Host "Node port: $NodePort"
    Write-Host "UI port: $UIPort"
    Write-Host "Data dir: $DataDir"
}
finally {
    Pop-Location
}
