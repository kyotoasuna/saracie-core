# Saracie Core v0.1.1

This release improves the first public mining flow.

## Changes

- Miner syncs chain and mempool data from peers before mining.
- Local UI syncs chain and mempool data from configured peers.
- UI default data directory is now `.saracie-ui`.
- Miner script default data directory is now `.saracie-miner`.
- Windows quickstart added for wallet, node, UI, and mining.
- Launch docs now recommend separate data directories for node, UI, and miner.

## Safety

Wallet files, local chain directories, build output, and release output remain ignored by Git.
