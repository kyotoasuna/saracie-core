package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
	"github.com/kyotoasuna/saracie-core/internal/node"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "new":
		err = newWallet(os.Args[2:])
	case "create":
		err = createWalletFile(os.Args[2:])
	case "open":
		err = openWalletFile(os.Args[2:])
	case "address":
		err = deriveAddress(os.Args[2:])
	case "balance":
		err = balance(os.Args[2:])
	case "send":
		err = send(os.Args[2:])
	case "send-file":
		err = sendFromFile(os.Args[2:])
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
	fmt.Println(`Saracie Wallet

Usage:
  saracie-wallet new --index 0
  saracie-wallet create --wallet saracie.wallet
  saracie-wallet open --wallet saracie.wallet
  saracie-wallet address --mnemonic "words..." --index 0
  saracie-wallet balance --datadir .saracie --address sar1...
  saracie-wallet send --datadir .saracie --mnemonic "words..." --to sar1... --amount 1 --fee 0.00001
  saracie-wallet send-file --datadir .saracie --wallet saracie.wallet --to sar1... --amount 1 --fee 0.00001`)
}

func newWallet(args []string) error {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	index := fs.Uint("index", 0, "address index")
	_ = fs.Parse(args)

	account, err := wallet.NewAccount(uint32(*index))
	if err != nil {
		return err
	}
	return printJSON(account)
}

func createWalletFile(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	walletPath := fs.String("wallet", "saracie.wallet", "encrypted wallet file")
	passphrase := fs.String("passphrase", "", "wallet passphrase, or use SARACIE_WALLET_PASSPHRASE")
	index := fs.Uint("index", 0, "address index")
	_ = fs.Parse(args)

	pass, err := readPassphrase(*passphrase, true)
	if err != nil {
		return err
	}
	info, err := wallet.CreateEncryptedWalletFile(*walletPath, pass, uint32(*index))
	if err != nil {
		return err
	}
	return printJSON(info)
}

func openWalletFile(args []string) error {
	fs := flag.NewFlagSet("open", flag.ExitOnError)
	walletPath := fs.String("wallet", "saracie.wallet", "encrypted wallet file")
	passphrase := fs.String("passphrase", "", "wallet passphrase, or use SARACIE_WALLET_PASSPHRASE")
	_ = fs.Parse(args)

	pass, err := readPassphrase(*passphrase, false)
	if err != nil {
		return err
	}
	keyPair, file, err := wallet.LoadEncryptedKeyPair(*walletPath, pass)
	if err != nil {
		return err
	}
	return printJSON(wallet.WalletFileInfo{
		Wallet:    *walletPath,
		Address:   keyPair.Account.Address,
		Path:      keyPair.Account.Path,
		Created:   file.CreatedAt,
		Encrypted: true,
	})
}

func deriveAddress(args []string) error {
	fs := flag.NewFlagSet("address", flag.ExitOnError)
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

func sendFromFile(args []string) error {
	fs := flag.NewFlagSet("send-file", flag.ExitOnError)
	dataDir := fs.String("datadir", ".saracie", "data directory")
	walletPath := fs.String("wallet", "saracie.wallet", "encrypted wallet file")
	passphrase := fs.String("passphrase", "", "wallet passphrase, or use SARACIE_WALLET_PASSPHRASE")
	to := fs.String("to", "", "recipient address")
	amountRaw := fs.String("amount", "", "amount in SRCE")
	feeRaw := fs.String("fee", "0.00001000", "fee in SRCE")
	peers := fs.String("peers", "", "comma-separated peer status API URLs")
	_ = fs.Parse(args)

	if *to == "" {
		return fmt.Errorf("--to is required")
	}
	if *amountRaw == "" {
		return fmt.Errorf("--amount is required")
	}

	pass, err := readPassphrase(*passphrase, false)
	if err != nil {
		return err
	}
	amount, err := consensus.ParseAmount(*amountRaw)
	if err != nil {
		return err
	}
	fee, err := consensus.ParseAmount(*feeRaw)
	if err != nil {
		return err
	}

	keyPair, _, err := wallet.LoadEncryptedKeyPair(*walletPath, pass)
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

func readPassphrase(flagValue string, confirm bool) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if env := os.Getenv("SARACIE_WALLET_PASSPHRASE"); env != "" {
		return env, nil
	}
	fmt.Fprint(os.Stderr, "Wallet passphrase: ")
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	pass := string(raw)
	if pass == "" {
		return "", fmt.Errorf("empty passphrase")
	}
	if !confirm {
		return pass, nil
	}

	fmt.Fprint(os.Stderr, "Confirm passphrase: ")
	againRaw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	if pass != string(againRaw) {
		return "", fmt.Errorf("passphrases do not match")
	}
	return pass, nil
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
