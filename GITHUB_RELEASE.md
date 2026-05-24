# GitHub Release Guide

This guide prepares the public Saracie Core release.

## 1. Create Repository

Recommended repository:

```text
saracie-core
```

Recommended description:

```text
Saracie (SRCE) - independent Proof-of-Work digital currency
```

## 2. Push Code

```bash
git remote add origin https://github.com/kyotoasuna/saracie-core.git
git branch -M main
git push -u origin main
```

## 3. Create First Release Tag

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub release workflow builds:

```text
saracied
saracie-miner
saracie-wallet
saracie-ui
```

For:

```text
Windows amd64
Linux amd64
```

Each release artifact includes `SHA256SUMS.txt`.

## 4. Release Notes Template

```text
Saracie Core v0.1.0

Initial public foundation release.

Includes:
- Saracie consensus parameters
- genesis block
- local chain storage
- Proof-of-Work mining
- signed transactions
- UTXO validation
- local mempool
- peer sync and gossip
- encrypted wallet file
- local browser UI
- Windows/Linux binaries

Ticker: SRCE
Supply: 210,000 SRCE
Halving: every 262,800 blocks
Premine: 0 SRCE
Address prefix: sar1
```

## 5. Vercel Site

Deploy the repository to Vercel and use:

```text
Output directory: site
```

The free Vercel URL can be used as the first public website.

## GitHub Pages Site

The repository also includes a GitHub Pages workflow.

Expected URL:

```text
https://kyotoasuna.github.io/saracie-core/
```
