# Saracie Local UI

Saracie Local UI is a browser-based desktop-style control panel served by a local Go binary.

It runs on the user's own machine and connects to the local Saracie data directory.

## Start

```bash
go run ./cmd/saracie-ui --datadir .saracie --listen 127.0.0.1:7340
```

Open:

```text
http://127.0.0.1:7340
```

## Current UI

- chain height;
- block reward;
- mempool count;
- encrypted wallet create/open;
- balance lookup;
- send from encrypted wallet file;
- miner start/stop;
- last mined block;
- last mining reward.

Wallet passphrases are sent only to the local UI server running on the user's machine. Keep the UI bound to `127.0.0.1` unless you know exactly why you are changing it.

## With Peers

```bash
go run ./cmd/saracie-ui --datadir .saracie --listen 127.0.0.1:7340 --peers http://peer-ip:7339
```

The UI miner submits mined blocks to configured peers.

## Local-Only Default

The recommended default listen address is:

```text
127.0.0.1:7340
```

This keeps wallet and mining controls local to the user's machine.
