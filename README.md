# Saracie Core

Saracie (SRCE) is an independent Proof-of-Work digital currency created by Kyoto Asuna.

The network is designed around a small fixed supply, public mining, simple wallets, and rules that anyone can verify.

Public repository:

```text
https://github.com/kyotoasuna/saracie-core
```

Website:

```text
https://kyotoasuna.github.io/saracie-core/
```

Latest release:

```text
https://github.com/kyotoasuna/saracie-core/releases/tag/v0.1.1
```

```text
Name:            Saracie
Ticker:          SRCE
Maximum supply:  210,000 SRCE
Block target:    60 seconds
Halving:         every 262,800 blocks
Approx. halving: every 6 months
Retarget:        every 60 blocks
Premine:         0 SRCE
Address prefix:  sar1...
Consensus:       Proof-of-Work
Ledger:          UTXO
```

## Current Build Status

This repository currently contains the first Saracie Core foundation:

- consensus parameters;
- genesis block;
- deterministic wallet/address generation;
- encrypted wallet files;
- `sar1...` address encoding;
- signed transactions;
- UTXO validation;
- local mempool;
- local chain storage;
- CPU mining for local blocks;
- node status HTTP API;
- peer list endpoint;
- block and transaction gossip;
- periodic peer sync;
- local browser UI;
- separate node, UI, and miner data directories;
- difficulty retargeting;
- public whitepaper and launch docs.

Stronger production-grade peer discovery and more user-friendly installers are the next implementation milestones.

## Quick Start

For a Windows step-by-step launch/mining guide, read:

```text
QUICKSTART_WINDOWS.md
```

Install Go, then run:

```bash
go test ./...
go run ./cmd/saracied params
go run ./cmd/saracied wallet-new --index 0
```

Initialize a local Saracie chain:

```bash
go run ./cmd/saracied init --datadir .saracie
```

Mine a block to your address:

```bash
go run ./cmd/saracied mine --datadir .saracie --address sar1... --blocks 1
```

Create an encrypted wallet file:

```bash
go run ./cmd/saracie-wallet create --wallet saracie.wallet
```

Check balance:

```bash
go run ./cmd/saracie-wallet balance --datadir .saracie --address sar1...
```

Send from an encrypted wallet file:

```bash
go run ./cmd/saracie-wallet send-file --datadir .saracie --wallet saracie.wallet --to sar1... --amount 0.1 --fee 0.00001
```

Run the status API:

```bash
go run ./cmd/saracied node --datadir .saracie --listen :7339
```

Then open:

```text
http://localhost:7339/status
```

Run a node with peers:

```bash
go run ./cmd/saracied node --datadir .saracie --listen :7339 --self http://your-ip:7339 --peers http://peer-ip:7339
```

Run the local UI:

```bash
go run ./cmd/saracie-ui --datadir .saracie-ui --listen 127.0.0.1:7340 --peers http://127.0.0.1:7339
```

## Commands

```text
saracied params
saracied genesis
saracied init --datadir .saracie
saracied status --datadir .saracie
saracied wallet-new --index 0
saracied wallet-address --mnemonic "words..." --index 0
saracied send --datadir .saracie --mnemonic "words..." --to sar1... --amount 1 --fee 0.00001
saracied mempool --datadir .saracie
saracied mine --datadir .saracie --address sar1... --blocks 1
saracied node --datadir .saracie --listen :7339 --self http://host:7339 --peers http://peer:7339
saracied sync --datadir .saracie --peers http://host:7339
```

## Documents

- [Whitepaper](WHITEPAPER.md)
- [Mining Guide](MINING.md)
- [Wallet Guide](WALLET.md)
- [Wallet Security](WALLET_SECURITY.md)
- [Local UI](UI.md)
- [Launch Notes](LAUNCH.md)
- [Launch Runbook](LAUNCH_RUNBOOK.md)
- [GitHub Release Guide](GITHUB_RELEASE.md)

## Build Release

From PowerShell:

```powershell
.\scripts\build-release.ps1 -Version v0.1.0
```

This creates Windows/Linux binaries and `SHA256SUMS.txt` under `dist/`.

## Public Preflight

Before pushing to GitHub:

```powershell
.\scripts\preflight-public.ps1
```

This checks that wallet files, seeds, private keys, local chain data, and build artifacts are not visible to Git.

Before starting a launch stack:

```powershell
.\scripts\preflight-launch.ps1
```

## Website Hosting

The official free website is currently published with GitHub Pages:

```text
https://kyotoasuna.github.io/saracie-core/
```

`vercel.json` is optional. It is only kept as a ready-to-use fallback if the site is later deployed to Vercel.

## Private Mining Wallet

Create your personal mining wallet locally:

```powershell
.\scripts\create-mining-wallet.ps1
```

The resulting `saracie-mining.wallet` file is ignored by Git and must stay private.

## Start Local Launch Stack

```powershell
.\scripts\start-seed-node.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7339" -Self "http://127.0.0.1:7339"
.\scripts\start-ui.ps1 -DataDir ".saracie" -Listen "127.0.0.1:7340" -Peers "http://127.0.0.1:7339"
.\scripts\start-miner.ps1 -DataDir ".saracie" -Address "sar1..." -Peers "http://127.0.0.1:7339"
```

Stop everything:

```powershell
.\scripts\stop-saracie.ps1
```

## Development Direction

The first public release should include:

- full node;
- wallet;
- encrypted wallet file;
- local miner;
- local browser UI;
- separate miner binary;
- separate wallet binary;
- signed transactions;
- UTXO validation;
- peer gossip;
- public source code;
- Windows/Linux builds;
- simple website;
- public launch instructions.

Saracie starts simple on purpose: a scarce mineable coin with clear rules and open software.

## License

MIT. See [LICENSE](LICENSE).
