# Saracie Wallet Security

This guide is for the private mining wallet used by the first miner and by future users.

## Core Rule

Never publish:

- wallet files;
- mnemonic words;
- passphrases;
- `.saracie` data directories;
- `.env` files;
- private keys;
- screenshots that show mnemonic words.

## Recommended Mining Wallet

Use an encrypted wallet file:

```bash
saracie-wallet create --wallet saracie-mining.wallet
```

The command asks for the passphrase in the terminal without printing it.

Then open it to get the public payout address:

```bash
saracie-wallet open --wallet saracie-mining.wallet
```

Mine to that public address:

```bash
saracie-miner --datadir .saracie --address sar1...
```

The miner only needs the public address. It does not need the wallet file or private keys.

## Passphrase Handling

Preferred:

```bash
saracie-wallet create --wallet saracie.wallet
```

Avoid putting the passphrase directly in shell history:

```bash
saracie-wallet create --wallet saracie.wallet --passphrase "..."
```

That form is supported for automation, but it is not recommended for the main personal wallet.

## GitHub Safety

The repository ignores local wallet and chain data:

```text
*.wallet
*.seed
*.mnemonic
*.secret
*.key
*.pem
.env
.saracie*/
bin/
dist/
```

Before pushing public code, run:

```powershell
.\scripts\preflight-public.ps1
```

The preflight fails if a wallet, key, seed, local chain directory, or release artifact is accidentally visible to Git.

## Backup

Keep at least two offline backups of the encrypted wallet file and passphrase.

The wallet file without the passphrase should not be enough to spend funds, but losing either the wallet file or the passphrase can make funds unreachable.

## First Miner Setup

Recommended flow:

```powershell
.\bin\saracie-wallet.exe create --wallet saracie-mining.wallet
.\bin\saracie-wallet.exe open --wallet saracie-mining.wallet
.\bin\saracie-miner.exe --datadir .saracie --address sar1...
```

Do not commit `saracie-mining.wallet`.

Helper script:

```powershell
.\scripts\create-mining-wallet.ps1
```

The helper creates or opens `saracie-mining.wallet` and prints the public payout address.
