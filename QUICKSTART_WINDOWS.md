# Saracie Windows Quickstart

This guide starts a local wallet, node, UI, and miner.

## 1. Open PowerShell

```powershell
cd C:\path\to\saracie-core
```

## 2. Create a Private Mining Wallet

```powershell
.\scripts\create-mining-wallet.ps1
```

Save the displayed `sar1...` address. Keep the wallet file and passphrase private.

## 3. Start a Local Seed Node

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7339" -Self "http://127.0.0.1:7339"
```

For a public seed node, use `0.0.0.0:7339` and a public IP or forwarded port.

## 4. Start the Local UI

```powershell
.\scripts\start-ui.ps1 -DataDir ".saracie-ui" -Listen "127.0.0.1:7340" -Peers "http://127.0.0.1:7339"
```

Open:

```text
http://127.0.0.1:7340
```

## 5. Start Mining

Replace `sar1...` with your payout address:

```powershell
.\scripts\start-miner.ps1 -DataDir ".saracie-miner" -Address "sar1..." -Peers "http://127.0.0.1:7339"
```

## 6. Check Status

```powershell
Invoke-RestMethod http://127.0.0.1:7339/status
```

Check balance:

```powershell
.\bin\saracie-wallet.exe balance --datadir ".saracie" --address "sar1..."
```

## 7. Stop

```powershell
.\scripts\stop-saracie.ps1
```

## Data Directories

```text
.saracie        seed node chain data
.saracie-ui     local UI sync data
.saracie-miner  local miner sync data
*.wallet        private encrypted wallet files
```

Do not publish wallet files, passphrases, or private mnemonics.
