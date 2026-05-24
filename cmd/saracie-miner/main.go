package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
	"github.com/kyotoasuna/saracie-core/internal/node"
)

func main() {
	dataDir := flag.String("datadir", ".saracie", "data directory")
	address := flag.String("address", "", "SRCE payout address")
	blocks := flag.Uint("blocks", 0, "number of blocks to mine, 0 means mine until stopped")
	peers := flag.String("peers", "", "comma-separated peer status API URLs")
	flag.Parse()

	if *address == "" {
		fmt.Fprintln(os.Stderr, "error: --address is required")
		os.Exit(1)
	}

	store, err := chain.Open(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("Saracie Miner")
	fmt.Println("payout:", *address)

	peerList := parsePeers(*peers)
	var mined uint
	for {
		if *blocks > 0 && mined >= *blocks {
			return
		}

		syncBeforeMining(store, peerList)

		block, err := store.MineNext(ctx, *address)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		mined++
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
