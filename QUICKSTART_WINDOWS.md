# Saracie Windows Quickstart

This guide starts a Saracie wallet, local UI, and miner from the Windows release zip.

## 1. Download and Extract

Download `saracie-windows-amd64.zip` from the latest release and extract it to a normal folder.

Do not run the tools directly from inside the zip preview.

## 2. Open PowerShell in the Extracted Folder

In File Explorer, open the extracted `saracie-windows-amd64` folder, click the address bar, type `powershell`, and press Enter.

If PowerShell blocks local scripts for this session, run:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```

## 3. Create a Private Mining Wallet

```powershell
.\scripts\create-mining-wallet.ps1
```

Save the displayed `sar1...` address. Keep the wallet file and passphrase private.

## 4. Start the Local UI

```powershell
.\scripts\start-ui.ps1 -DataDir ".saracie-ui" -Listen "127.0.0.1:7340" -Peers "http://SEED_NODE:7339"
```

Replace `http://SEED_NODE:7339` with a public Saracie seed node URL.

Open:

```text
http://127.0.0.1:7340
```

## 5. Start Mining

Replace `sar1...` with your payout address:

```powershell
.\scripts\start-miner.ps1 -DataDir ".saracie-miner" -Address "sar1..." -Peers "http://SEED_NODE:7339"
```

## 6. Check Status

```powershell
Invoke-RestMethod http://SEED_NODE:7339/status
```

Check balance:

```powershell
.\saracie-wallet.exe balance --datadir ".saracie-ui" --address "sar1..."
```

## Optional: Run a Local Seed Node

If you also want to run your own local node:

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7339" -Self "http://127.0.0.1:7339" -Peers "http://SEED_NODE:7339"
```

If you have a public IP and port forwarding, you can run a public seed node:

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "0.0.0.0:7339" -Self "http://YOUR_PUBLIC_IP:7339" -Peers "http://SEED_NODE:7339"
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
