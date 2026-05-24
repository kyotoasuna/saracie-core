$ErrorActionPreference = "Stop"

$names = @("saracied", "saracie-miner", "saracie-wallet", "saracie-ui")
$stopped = 0

foreach ($name in $names) {
    $processes = Get-Process -Name $name -ErrorAction SilentlyContinue
    foreach ($process in $processes) {
        Stop-Process -Id $process.Id -Force
        Write-Host "Stopped $name pid=$($process.Id)"
        $stopped++
    }
}

if ($stopped -eq 0) {
    Write-Host "No Saracie processes were running."
}
