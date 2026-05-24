# Saracie Mainnet Launch Runbook

This is the private operational runbook for starting Saracie mainnet.

## 1. Build

```powershell
.\scripts\build-release.ps1 -Version v0.1.0
```

## 2. Public Preflight

```powershell
.\scripts\preflight-public.ps1
```

Do not continue if this fails.

Then run launch preflight:

```powershell
.\scripts\preflight-launch.ps1
```

## 3. Create Private Mining Wallet

Create a fresh wallet at launch time:

```powershell
.\scripts\create-mining-wallet.ps1
```

Write down the public `sar1...` address. Keep the `.wallet` file private.

## 4. Start Seed Node

If you have a public IP and port forwarding for `7339`:

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "0.0.0.0:7339" -Self "http://YOUR_PUBLIC_IP:7339"
```

For local-only launch testing:

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7339" -Self "http://127.0.0.1:7339"
```

Check:

```powershell
Invoke-RestMethod http://127.0.0.1:7339/status
```

## 5. Start UI

```powershell
.\scripts\start-ui.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7340" -Peers "http://127.0.0.1:7339"
```

Open:

```text
http://127.0.0.1:7340
```

## 6. Start Mining

Use the public payout address from the private wallet:

```powershell
.\scripts\start-miner.ps1 -DataDir ".saracie" -Address "sar1..." -Peers "http://127.0.0.1:7339"
```

## 7. Watch First Blocks

```powershell
Invoke-RestMethod http://127.0.0.1:7339/status
```

The first mined block should move height from `0` to `1`.

## 8. Stop Local Processes

```powershell
.\scripts\stop-saracie.ps1
```

## Launch Rule

Publish source and release files before public mining starts. The private mining wallet is never committed, uploaded, screenshotted, or shared.
