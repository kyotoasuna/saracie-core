param(
    [string]$Wallet = "saracie-mining.wallet"
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Push-Location $Root
try {
    $walletExe = Join-Path $Root "bin\saracie-wallet.exe"
    $useBinary = Test-Path $walletExe

    if (Test-Path $Wallet) {
        Write-Host "Opening existing encrypted mining wallet: $Wallet"
        if ($useBinary) {
            & $walletExe open --wallet $Wallet
        } else {
            go run ./cmd/saracie-wallet open --wallet $Wallet
        }
        return
    }

    Write-Host "Creating encrypted mining wallet: $Wallet"
    if ($useBinary) {
        & $walletExe create --wallet $Wallet
    } else {
        go run ./cmd/saracie-wallet create --wallet $Wallet
    }

    Write-Host ""
    Write-Host "Use the displayed address as your mining payout address."
    Write-Host "The wallet file is ignored by Git and must stay private."
}
finally {
    Pop-Location
}
