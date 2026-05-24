package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/ui"
)

func main() {
	dataDir := flag.String("datadir", ".saracie-ui", "data directory")
	listen := flag.String("listen", "127.0.0.1:7340", "local UI listen address")
	peers := flag.String("peers", "", "comma-separated peer status API URLs")
	flag.Parse()

	store, err := chain.Open(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	server := ui.New(store, *dataDir, parsePeers(*peers))
	fmt.Println("Saracie UI listening at", ui.LocalURL(*listen))
	if err := server.ListenAndServe(*listen); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
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
