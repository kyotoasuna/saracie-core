# Saracie Wallet Guide

A Saracie wallet controls private keys. The coins themselves exist on the blockchain as unspent transaction outputs.

## Address Format

Saracie addresses use the `sar1` prefix:

```text
sar1...
```

This makes SRCE addresses visually different from Bitcoin and other networks.

## How Wallets Are Calculated

The first wallet model is deterministic.

```text
random entropy
-> mnemonic words
-> seed
-> master private key
-> child private key
-> public key
-> public key hash
-> sar1 address
```

The derivation path used by the current implementation is:

```text
m/84'/7331'/0'/0/index
```

`7331` is the Saracie coin type used by this implementation.

## Create a Wallet

```bash
go run ./cmd/saracied wallet-new --index 0
```

Or use the separate wallet command:

```bash
go run ./cmd/saracie-wallet new --index 0
```

The command returns:

- mnemonic;
- derivation path;
- public address;
- public key;
- public key hash.

Save the mnemonic offline. Anyone with the mnemonic can derive the same wallet.

## Create an Encrypted Wallet File

```bash
go run ./cmd/saracie-wallet create --wallet saracie.wallet
```

The encrypted wallet file stores the mnemonic using:

```text
scrypt key derivation
AES-256-GCM encryption
```

You can also pass the passphrase through an environment variable:

```bash
SARACIE_WALLET_PASSPHRASE="choose-a-passphrase" go run ./cmd/saracie-wallet create --wallet saracie.wallet
```

Open the encrypted wallet metadata:

```bash
go run ./cmd/saracie-wallet open --wallet saracie.wallet
```

## Recreate an Address

```bash
go run ./cmd/saracied wallet-address --mnemonic "words..." --index 0
```

Or:

```bash
go run ./cmd/saracie-wallet address --mnemonic "words..." --index 0
```

## Check Balance

```bash
go run ./cmd/saracie-wallet balance --datadir .saracie --address sar1...
```

## Send SRCE

```bash
go run ./cmd/saracie-wallet send --datadir .saracie --mnemonic "words..." --to sar1... --amount 0.1 --fee 0.00001
```

Send from an encrypted wallet file:

```bash
go run ./cmd/saracie-wallet send-file --datadir .saracie --wallet saracie.wallet --to sar1... --amount 0.1 --fee 0.00001
```

The send command creates a signed transaction and places it in the local mempool. A miner includes pending transactions in the next block.

For personal funds and mining rewards, prefer encrypted wallet files over commands that pass mnemonic words directly.

Use a different index for more receiving addresses:

```bash
go run ./cmd/saracied wallet-address --mnemonic "words..." --index 1
```

## Wallet Roadmap

The current release foundation supports deterministic address generation, encrypted wallet files, balances, signed transactions, and local mining payouts.

The next wallet milestones are:

- encrypted wallet file improvements;
- send/receive screen;
- transaction history;
- backup workflow;
- desktop wallet UI;
- later browser extension wallet.

The desktop wallet comes before the browser extension because it is simpler and safer for the first mainnet release.
