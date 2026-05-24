package node

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/chain"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

func TestSyncFromPeer(t *testing.T) {
	source, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	account, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.MineNext(context.Background(), account.Address); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(New(source).Handler())
	defer server.Close()

	target, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	replaced, err := SyncFromPeers(target, []string{server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if replaced != 1 {
		t.Fatalf("replaced = %d, want 1", replaced)
	}
	if target.Tip().Header.Height != 1 {
		t.Fatalf("height = %d, want 1", target.Tip().Header.Height)
	}
}

func TestBlockGossip(t *testing.T) {
	source, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	miner, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	account, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	block, err := miner.MineNext(context.Background(), account.Address)
	if err != nil {
		t.Fatal(err)
	}

	targetHTTP := httptest.NewServer(New(target).Handler())
	defer targetHTTP.Close()

	sourceServer := NewWithPeers(source, "", []string{targetHTTP.URL})
	sourceHTTP := httptest.NewServer(sourceServer.Handler())
	defer sourceHTTP.Close()

	if err := SubmitBlock(sourceHTTP.URL, block); err != nil {
		t.Fatal(err)
	}

	waitFor(t, time.Second, func() bool {
		return target.Tip().Header.Height == 1
	})
}

func TestTransactionGossip(t *testing.T) {
	source, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target, err := chain.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	sender, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	senderKeys, err := wallet.DeriveKeyPair(sender.Mnemonic, 0)
	if err != nil {
		t.Fatal(err)
	}
	receiver, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	receiverHash, err := wallet.AddressToPubKeyHash(receiver.Address)
	if err != nil {
		t.Fatal(err)
	}

	block, err := source.MineNext(context.Background(), sender.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err := target.AddBlock(block); err != nil {
		t.Fatal(err)
	}

	utxos, err := source.SpendableUTXOs(senderKeys.PubKeyHash)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := chain.NewSignedTransaction(
		utxos,
		senderKeys.PrivateKey,
		senderKeys.PublicKey,
		receiverHash,
		senderKeys.PubKeyHash,
		10_000_000,
		1_000,
	)
	if err != nil {
		t.Fatal(err)
	}

	targetHTTP := httptest.NewServer(New(target).Handler())
	defer targetHTTP.Close()

	sourceServer := NewWithPeers(source, "", []string{targetHTTP.URL})
	sourceHTTP := httptest.NewServer(sourceServer.Handler())
	defer sourceHTTP.Close()

	if err := SubmitTransaction(sourceHTTP.URL, tx); err != nil {
		t.Fatal(err)
	}

	waitFor(t, time.Second, func() bool {
		return len(target.Mempool) == 1
	})
}

func waitFor(t *testing.T, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
