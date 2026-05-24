param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Dist = Join-Path $Root "dist"
$Release = Join-Path $Dist "saracie-$Version"

New-Item -ItemType Directory -Force -Path $Release | Out-Null

Push-Location $Root
try {
    go test ./...

    $targets = @(
        @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
        @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" }
    )

    foreach ($target in $targets) {
        $env:GOOS = $target.GOOS
        $env:GOARCH = $target.GOARCH
        $outDir = Join-Path $Release "$($target.GOOS)-$($target.GOARCH)"
        New-Item -ItemType Directory -Force -Path $outDir | Out-Null

        go build -o (Join-Path $outDir "saracied$($target.Ext)") ./cmd/saracied
        go build -o (Join-Path $outDir "saracie-miner$($target.Ext)") ./cmd/saracie-miner
        go build -o (Join-Path $outDir "saracie-wallet$($target.Ext)") ./cmd/saracie-wallet
        go build -o (Join-Path $outDir "saracie-ui$($target.Ext)") ./cmd/saracie-ui
    }

    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

    Get-ChildItem -Path $Release -Recurse -File |
        Get-FileHash -Algorithm SHA256 |
        ForEach-Object { "$($_.Hash.ToLower())  $($_.Path.Substring($Release.Length + 1).Replace('\', '/'))" } |
        Set-Content -Path (Join-Path $Release "SHA256SUMS.txt")

    Write-Host "Release built at $Release"
}
finally {
    Pop-Location
}
