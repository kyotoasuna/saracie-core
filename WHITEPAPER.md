# Saracie (SRCE)

## A Simple Mineable Digital Currency

Version: 1.0 draft  
Creator: Kyoto Asuna  
Ticker: SRCE  
Network: Saracie Mainnet  
Consensus: Proof-of-Work  

## Abstract

Saracie (SRCE) is an independent mineable digital currency with a fixed supply of 210,000 coins.

It is built around a simple monetary idea: every coin is created by Proof-of-Work, every transaction is verified by the network, and every user can run a node, mine blocks, hold coins, and send value without relying on a central issuer.

SRCE is not a token on another blockchain. It is its own network, with its own blocks, nodes, wallets, miners, addresses, and monetary schedule.

The design is intentionally simple:

```text
Fixed supply.
Proof-of-Work mining.
No premine.
No private allocation.
Six-month halvings.
Open-source node software.
Simple wallet model.
```

Saracie is designed to be easy to explain, easy to run, easy to mine, and easy to verify.

## 1. The Idea

Most digital currencies try to become platforms. Saracie starts from a smaller and clearer goal:

```text
A scarce digital coin that anyone can mine and anyone can verify.
```

The network is built for people who want a direct monetary system, not a complicated application stack.

Saracie focuses on:

- scarcity;
- open mining;
- transparent issuance;
- independent nodes;
- simple wallets;
- public rules.

The main question behind SRCE is simple:

```text
What if a new Proof-of-Work coin started with a much smaller supply,
a faster halving cycle, and a clean public launch?
```

Saracie is the answer to that question.

## 2. Monetary Parameters

Saracie uses a fixed monetary policy.

```text
Coin name:        Saracie
Ticker:           SRCE
Maximum supply:   210,000 SRCE
Decimals:         8
Smallest unit:    0.00000001 SRCE
Block target:     60 seconds
Halving period:   262,800 blocks
Approx. halving:  6 months
Retarget period:  60 blocks
Premine:          0 SRCE
Private sale:     none
Founder reserve:  none
```

The total supply is 100 times smaller than Bitcoin's 21,000,000 BTC supply.

Saracie does not create coins through staking, minting, governance votes, or central allocation. New SRCE enters circulation only through block rewards paid to miners.

## 3. Emission and Halving

Saracie uses a halving schedule.

Every 262,800 blocks, the block reward is reduced by half. With a target block time of 60 seconds, this is approximately every six months.

The first reward era distributes approximately half of the total supply:

```text
Blocks in first era:       262,800
First era issuance:        about 105,000 SRCE
Initial block reward:      about 0.39954337 SRCE
```

Estimated emission:

```text
Era 1: 105,000 SRCE
Era 2:  52,500 SRCE
Era 3:  26,250 SRCE
Era 4:  13,125 SRCE
Era 5:   6,562.5 SRCE
Era 6:   3,281.25 SRCE
```

The halving schedule makes the supply curve easy to understand. Early in the network, mining rewards are larger. Over time, new issuance becomes smaller, and the remaining supply becomes harder to obtain through mining.

The supply cap is enforced by the consensus rules of the network.

## 4. Proof-of-Work Mining

Saracie is secured by Proof-of-Work.

Miners collect pending transactions, build candidate blocks, and search for a valid block hash. A block is accepted by the network only when it satisfies the current difficulty target and follows all consensus rules.

When a miner finds a valid block, the miner receives:

```text
block reward + transaction fees
```

Mining performs three important functions:

- it creates new SRCE according to the emission schedule;
- it orders transactions into blocks;
- it protects the chain by making history expensive to rewrite.

The first implementation is planned as a Bitcoin-style Proof-of-Work chain. This keeps the system familiar and easy to implement while preserving the core idea of open mining.

Saracie adjusts mining difficulty every 60 blocks. If blocks are found too quickly, the next interval becomes harder. If blocks are found too slowly, the next interval becomes easier, within bounded adjustment limits.

## 5. Nodes

A Saracie node is a program that connects to the network and verifies the blockchain.

Nodes are important because they enforce the rules. A node checks:

- that blocks are valid;
- that Proof-of-Work is correct;
- that transactions do not spend the same coins twice;
- that block rewards are not too high;
- that the 210,000 SRCE supply cap is respected;
- that the halving schedule is followed.

Users do not need permission to run a node. A stronger network is created when more people run their own nodes.

## 6. Transactions

Saracie uses the UTXO model.

In this model, coins are tracked as unspent transaction outputs. A wallet balance is the sum of all UTXOs controlled by that wallet's private keys.

When a user sends SRCE:

1. The wallet selects enough UTXOs to cover the payment.
2. The wallet creates a transaction.
3. The wallet signs the transaction with the correct private keys.
4. The transaction is broadcast to the network.
5. Miners include the transaction in a block.
6. Nodes verify that the transaction follows the rules.

This model is simple, proven, and easy for full nodes to validate.

## 7. Wallets and Addresses

A Saracie wallet controls private keys.

Coins are not stored inside the wallet file itself. Coins exist on the blockchain as UTXOs. The wallet stores or derives the private keys that can spend those UTXOs.

The planned wallet model uses hierarchical deterministic key generation.

The basic process is:

```text
random entropy
-> master seed
-> master private key
-> child private keys
-> public keys
-> Saracie addresses
```

Each address is derived from a public key. A user can generate many receiving addresses from the same wallet.

Planned address format:

```text
sar1...
```

The `sar1` prefix makes Saracie addresses visually distinct from Bitcoin and other networks.

Wallet functions in the first release:

- create wallet;
- encrypt wallet;
- generate receive address;
- send SRCE;
- show balance;
- show transaction history;
- back up wallet;
- connect to Saracie nodes.

The goal is a wallet that normal users can understand while still giving technical users direct access through command-line and RPC tools.

## 8. Scarcity Clock

Saracie includes a simple public concept called the Scarcity Clock.

The Scarcity Clock shows the monetary state of the network:

- current block height;
- current reward era;
- current block reward;
- estimated time until next halving;
- SRCE already mined;
- SRCE left to mine;
- percentage of total supply issued.

The Scarcity Clock is calculated from the blockchain itself. It does not require trust in a company, website, or central server.

This makes Saracie's monetary policy visible to everyone.

## 9. Why Adopt Saracie

Saracie is designed to be adopted because it is simple.

Users can understand the coin:

```text
210,000 total supply.
Proof-of-Work mining.
Halving every six months.
No premine.
No private allocation.
```

Miners can understand the opportunity:

```text
Run mining software.
Find valid blocks.
Receive SRCE block rewards.
Help secure the network.
```

Node operators can understand their role:

```text
Run a node.
Verify the chain.
Relay blocks and transactions.
Protect the rules.
```

Developers can understand the base:

```text
UTXO model.
RPC interface.
Open-source software.
Simple monetary rules.
Room for tools, explorers, wallets, pools, and services.
```

Saracie does not ask users to understand a complex financial machine. It gives them a scarce mineable coin with rules they can verify.

## 10. Launch Model

Saracie mainnet is intended to launch as a public Proof-of-Work network.

The launch model:

```text
Open-source code.
Public mainnet parameters.
Public genesis block.
Public wallet instructions.
Public mining instructions.
Public node software.
No premine.
```

The first node opens the network. After mainnet is live, miners can connect, mine blocks, and compete for block rewards under the same consensus rules.

Every SRCE coin is created by mining.

## 11. Development Path

Saracie begins as a simple base currency and leaves room for future development.

Initial developer tools:

- full node;
- wallet;
- command-line interface;
- JSON-RPC interface;
- mining support;
- raw transaction tools;
- block and transaction data.

Future ecosystem tools may include:

- block explorer;
- mining pool software;
- lightweight wallet;
- mobile wallet;
- payment tools;
- public APIs;
- developer documentation.

The project can grow without making the base layer unnecessarily complicated.

## 12. Upgrades

Saracie is designed to be simple at launch, but not frozen forever.

Future changes can be proposed through public improvement proposals.

Recommended process:

```text
proposal
-> discussion
-> implementation
-> public release
-> node/miner adoption
-> activation if required
```

The base rule is conservative: changes that affect consensus should be rare, clear, and easy for the community to verify.

## 13. Summary

Saracie is a new independent Proof-of-Work coin with:

```text
210,000 SRCE maximum supply
60-second block target
6-month halving schedule
0 premine
0 private allocation
native mining
native wallet
independent nodes
open-source development
```

It is built for people who want to mine, run nodes, send transactions, and participate in a simple scarce digital currency from the beginning.

Saracie starts with one clear idea:

```text
Scarcity should be simple enough for anyone to verify.
```
