# Saracie Mining Guide

Saracie is a Proof-of-Work network. Miners produce blocks, secure the chain, and receive SRCE block rewards.

## Mining Parameters

```text
Block target:       60 seconds
Initial reward:     0.39954337 SRCE
Halving interval:   262,800 blocks
Approx. halving:    6 months
Difficulty retarget: every 60 blocks
Maximum supply:     210,000 SRCE
Premine:            0 SRCE
```

## Generate a Payout Address

```bash
go run ./cmd/saracied wallet-new --index 0
```

Save the mnemonic offline. The miner only needs the public address.

Example address:

```text
sar1...
```

## Initialize the Chain

```bash
go run ./cmd/saracied init --datadir .saracie
```

## Mine Blocks

```bash
go run ./cmd/saracied mine --datadir .saracie --address sar1... --blocks 1
```

Or use the separate miner command:

```bash
go run ./cmd/saracie-miner --datadir .saracie-miner --address sar1... --blocks 1
```

To submit mined blocks to a running node API:

```bash
go run ./cmd/saracie-miner --datadir .saracie-miner --address sar1... --blocks 1 --peers http://127.0.0.1:7339
```

The miner syncs from configured peers before mining, builds a candidate block, includes valid mempool transactions, searches for a valid Proof-of-Work hash, writes the block to the local miner chain, and submits the block to peers.

## Check Status

```bash
go run ./cmd/saracied status --datadir .saracie
```

## Run a Local Status Node

```bash
go run ./cmd/saracied node --datadir .saracie --listen :7339
```

Open:

```text
http://localhost:7339/status
```

## Mining Transactions

When a wallet creates a transaction, it enters the local mempool.

```bash
go run ./cmd/saracied mempool --datadir .saracie
```

The next mined block includes valid mempool transactions and pays transaction fees to the miner.

## Mining With Peers

Run a node:

```bash
go run ./cmd/saracied node --datadir .saracie --listen :7339 --self http://your-ip:7339 --peers http://peer-ip:7339
```

Mine and submit blocks to that node:

```bash
go run ./cmd/saracie-miner --datadir .saracie-miner --address sar1... --peers http://127.0.0.1:7339
```

The node will gossip accepted blocks and transactions to known peers.

The miner should not need private keys. It only needs a payout address.
