package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
	"github.com/kyotoasuna/saracie-core/internal/node"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "params":
		err = printJSON(consensus.Mainnet)
	case "genesis":
		err = genesis()
	case "init":
		err = initChain(os.Args[2:])
	case "status":
		err = status(os.Args[2:])
	case "balance":
		err = balance(os.Args[2:])
	case "wallet-new":
		err = walletNew(os.Args[2:])
	case "wallet-address":
		err = walletAddress(os.Args[2:])
	case "send":
		err = send(os.Args[2:])
	case "mempool":
		err = mempool(os.Args[2:])
	case "mine":
		err = mine(os.Args[2:])
	case "node":
		err = runNode(os.Args[2:])
	case "sync":
		err = syncPeers(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Saracie Core

Usage:
  saracied params
  saracied genesis
  saracied init --datadir .saracie
  saracied status --datadir .saracie
  saracied balance --datadir .saracie --address sar1...
  saracied wallet-new --index 0
  saracied wallet-address --mnemonic "words..." --index 0
  saracied send --datadir .saracie --mnemonic "words..." --to sar1... --amount 1 --fee 0.00001
  saracied mempool --datadir .saracie
  saracied mine --datadir .saracie --address sar1... --blocks 1 --peers http://host:7339
  saracied node --datadir .saracie --listen :7339 --self http://host:7339 --peers http://host:7339
  saracied sync --datadir .saracie --peers http://host:7339`)
}

func genesis() error {
	block := chain.NewGenesis()
	return printJSON(struct {
		Hash  string      `json:"hash"`
		Block chain.Block `json:"block"`
	}{
		Hash:  block.Hash(),
		Block: block,
	})
}

func initChain(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	_ = fs.Parse(args)

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	return printJSON(store.Status())
}

func status(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	_ = fs.Parse(args)

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	return printJSON(store.Status())
}

func balance(args []string) error {
	fs := flag.NewFlagSet("balance", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	address := fs.String("address", "", "address")
	_ = fs.Parse(args)
	if *address == "" {
		return fmt.Errorf("--address is required")
	}

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	bal, err := store.Balance(*address)
	if err != nil {
		return err
	}
	return printJSON(bal)
}

func walletNew(args []string) error {
	fs := flag.NewFlagSet("wallet-new", flag.ExitOnError)
	index := fs.Uint("index", 0, "address index")
	_ = fs.Parse(args)

	account, err := wallet.NewAccount(uint32(*index))
	if err != nil {
		return err
	}
	return printJSON(account)
}

func walletAddress(args []string) error {
	fs := flag.NewFlagSet("wallet-address", flag.ExitOnError)
	mnemonic := fs.String("mnemonic", "", "BIP39 mnemonic")
	index := fs.Uint("index", 0, "address index")
	_ = fs.Parse(args)
	if *mnemonic == "" {
		return fmt.Errorf("--mnemonic is required")
	}

	account, err := wallet.Derive(*mnemonic, uint32(*index))
	if err != nil {
		return err
	}
	return printJSON(account)
}

func send(args []string) error {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	mnemonic := fs.String("mnemonic", "", "sender BIP39 mnemonic")
	index := fs.Uint("index", 0, "sender address index")
	to := fs.String("to", "", "recipient address")
	amountRaw := fs.String("amount", "", "amount in SRCE")
	feeRaw := fs.String("fee", "0.00001000", "fee in SRCE")
	peers := fs.String("peers", "", "comma-separated peer status API URLs")
	_ = fs.Parse(args)

	if *mnemonic == "" {
		return fmt.Errorf("--mnemonic is required")
	}
	if *to == "" {
		return fmt.Errorf("--to is required")
	}
	if *amountRaw == "" {
		return fmt.Errorf("--amount is required")
	}

	amount, err := consensus.ParseAmount(*amountRaw)
	if err != nil {
		return err
	}
	fee, err := consensus.ParseAmount(*feeRaw)
	if err != nil {
		return err
	}

	keyPair, err := wallet.DeriveKeyPair(*mnemonic, uint32(*index))
	if err != nil {
		return err
	}
	toPubKeyHash, err := wallet.AddressToPubKeyHash(*to)
	if err != nil {
		return err
	}

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	utxos, err := store.SpendableUTXOs(keyPair.PubKeyHash)
	if err != nil {
		return err
	}

	tx, err := chain.NewSignedTransaction(utxos, keyPair.PrivateKey, keyPair.PublicKey, toPubKeyHash, keyPair.PubKeyHash, amount, fee)
	if err != nil {
		return err
	}
	if err := store.AddMempoolTx(tx); err != nil {
		return err
	}

	for _, peer := range parsePeers(*peers) {
		if err := node.SubmitTransaction(peer, tx); err != nil {
			fmt.Fprintf(os.Stderr, "peer transaction submit failed for %s: %v\n", peer, err)
		}
	}

	return printJSON(tx)
}

func mempool(args []string) error {
	fs := flag.NewFlagSet("mempool", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	_ = fs.Parse(args)

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	return printJSON(store.Mempool)
}

func mine(args []string) error {
	fs := flag.NewFlagSet("mine", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	address := fs.String("address", "", "payout address")
	blocks := fs.Uint("blocks", 1, "number of blocks to mine")
	peers := fs.String("peers", "", "comma-separated peer status API URLs")
	_ = fs.Parse(args)
	if *address == "" {
		return fmt.Errorf("--address is required")
	}

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	peerList := parsePeers(*peers)
	for i := uint(0); i < *blocks; i++ {
		syncBeforeMining(store, peerList)

		block, err := store.MineNext(ctx, *address)
		if err != nil {
			return err
		}
		fmt.Printf("mined height=%d hash=%s reward=%s\n",
			block.Header.Height,
			block.Hash(),
			consensus.FormatAmount(coinbaseValue(block)),
		)
		for _, peer := range peerList {
			if err := node.SubmitBlock(peer, block); err != nil {
				fmt.Fprintf(os.Stderr, "peer submit failed for %s: %v\n", peer, err)
			}
		}
	}

	return nil
}

func syncBeforeMining(store *chain.Store, peers []string) {
	if len(peers) == 0 {
		return
	}
	if _, err := node.SyncFromPeers(store, peers); err != nil {
		fmt.Fprintf(os.Stderr, "peer chain sync warning: %v\n", err)
	}
	if _, err := node.SyncMempoolFromPeers(store, peers); err != nil {
		fmt.Fprintf(os.Stderr, "peer mempool sync warning: %v\n", err)
	}
}

func coinbaseValue(block chain.Block) int64 {
	if len(block.Transactions) == 0 {
		return 0
	}
	var total int64
	for _, out := range block.Transactions[0].Outputs {
		total += out.Value
	}
	return total
}

func runNode(args []string) error {
	fs := flag.NewFlagSet("node", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	listen := fs.String("listen", ":7339", "HTTP listen address")
	self := fs.String("self", "", "public URL for this node, announced to peers")
	peers := fs.String("peers", "", "comma-separated peer status API URLs")
	syncEvery := fs.Duration("sync-every", 30*time.Second, "peer sync interval, 0 disables periodic sync")
	_ = fs.Parse(args)

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}

	server := node.NewWithPeers(store, *self, parsePeers(*peers))
	server.DiscoverPeers()

	if synced, txs, err := server.SyncNetwork(); err != nil {
		fmt.Fprintf(os.Stderr, "peer sync failed: %v\n", err)
	} else if synced > 0 || txs > 0 {
		fmt.Printf("synced chains=%d transactions=%d\n", synced, txs)
	}

	if *syncEvery > 0 {
		go func() {
			ticker := time.NewTicker(*syncEvery)
			defer ticker.Stop()
			for range ticker.C {
				server.DiscoverPeers()
				if synced, txs, err := server.SyncNetwork(); err != nil {
					fmt.Fprintf(os.Stderr, "peer sync failed: %v\n", err)
				} else if synced > 0 || txs > 0 {
					fmt.Printf("synced chains=%d transactions=%d\n", synced, txs)
				}
			}
		}()
	}

	fmt.Printf("Saracie node status API listening on %s\n", *listen)
	fmt.Println("Endpoints: /status /params /blocks /transactions /mempool /peers /scarcity")
	return server.ListenAndServe(*listen)
}

func syncPeers(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	peers := fs.String("peers", "", "comma-separated peer status API URLs")
	_ = fs.Parse(args)

	store, err := chain.Open(*dataDir)
	if err != nil {
		return err
	}
	if _, err := node.SyncFromPeers(store, parsePeers(*peers)); err != nil {
		return err
	}
	if _, err := node.SyncMempoolFromPeers(store, parsePeers(*peers)); err != nil {
		return err
	}
	return printJSON(store.Status())
}

func parsePeers(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	peers := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			peers = append(peers, part)
		}
	}
	return peers
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
