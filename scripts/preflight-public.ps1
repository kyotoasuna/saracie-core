$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Push-Location $Root
try {
    $blockedPatterns = @(
        '\.wallet$',
        '\.wallet\.json$',
        '\.seed$',
        '\.mnemonic$',
        '\.secret$',
        '\.key$',
        '\.pem$',
        '\.p12$',
        '\.pfx$',
        '(^|/)\.env(\.|$)',
        '(^|/)\.saracie',
        '(^|/)bin/',
        '(^|/)dist/'
    )

    $tracked = git ls-files
    $candidates = git ls-files --others --exclude-standard
    $all = @($tracked) + @($candidates)

    $bad = @()
    foreach ($file in $all) {
        $normalized = $file -replace '\\', '/'
        foreach ($pattern in $blockedPatterns) {
            if ($normalized -match $pattern) {
                $bad += $normalized
                break
            }
        }
    }

    if ($bad.Count -gt 0) {
        Write-Host "Blocked files would be public if committed:" -ForegroundColor Red
        $bad | Sort-Object -Unique | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
        exit 1
    }

    $ignoredSensitive = git status --short --ignored |
        Where-Object { $_ -match '!! ' } |
        Where-Object {
            $_ -match '\.wallet|\.seed|\.mnemonic|\.secret|\.key|\.pem|\.env|\.saracie|bin/|dist/'
        }

    Write-Host "Public preflight passed." -ForegroundColor Green
    if ($ignoredSensitive) {
        Write-Host "Sensitive/local files are ignored and will not be committed:" -ForegroundColor Yellow
        $ignoredSensitive | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
    }
}
finally {
    Pop-Location
}
