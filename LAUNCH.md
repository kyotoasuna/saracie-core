# Saracie Launch Notes

Saracie mainnet should launch publicly with source code, wallet instructions, mining instructions, and binaries available before public mining starts.

## Launch Identity

```text
Name:     Saracie
Ticker:   SRCE
Creator:  Kyoto Asuna
Network:  Saracie Mainnet
Prefix:   sar1
```

## Public Launch Order

1. Publish source code.
2. Publish whitepaper.
3. Publish mining guide.
4. Publish wallet guide.
5. Publish release binaries.
6. Publish checksums.
7. Start initial seed/status node.
8. Announce that mainnet is live.
9. Start public mining.
10. Track the first blocks publicly.

## Zero-Budget Hosting

The launch can use:

```text
GitHub Releases     source code and binaries
GitHub repository   whitepaper and docs
Vercel              free static website on .vercel.app
Local PC            first node and first miner
```

The website can be deployed from the `site` directory.

## Mainnet Readiness Checklist

```text
[x] HTTP peer block sync implemented
[x] HTTP transaction propagation implemented
[x] basic peer discovery implemented
[x] signed user transactions implemented
[x] UTXO validation implemented
[x] local mempool implemented
[x] difficulty retarget implemented
[x] cumulative work chain selection implemented
[x] encrypted wallet file implemented
[ ] wallet backup UX implemented
[x] miner binary separated
[x] wallet binary separated
[x] local UI binary created
[x] release build script created
[x] GitHub release workflow created
[x] public preflight script created
[x] launch preflight script created
[x] launch runbook created
[x] startup scripts created
[ ] Windows build created for public release
[ ] Linux build created for public release
[ ] release checksums created for public release
[ ] seed node address published
[ ] website deployed
[x] Vercel static site config created
[x] GitHub Pages workflow created
[ ] launch time announced
```

## Current Technical State

The repository currently supports local wallet generation, encrypted wallet files, local chain initialization, signed transactions, UTXO validation, local mempool, status API, peer HTTP sync, peer discovery, block/transaction gossip, local mining, and a local browser UI.

The next required step before a real public mainnet is stronger production-grade P2P networking and release packaging.
