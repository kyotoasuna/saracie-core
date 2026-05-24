package chain

import (
	"context"
	"testing"

	"github.com/kyotoasuna/saracie-core/internal/consensus"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

func TestMineNextBlock(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	account, err := wallet.NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}

	block, err := store.MineNext(context.Background(), account.Address)
	if err != nil {
		t.Fatal(err)
	}

	if block.Header.Height != 1 {
		t.Fatalf("height = %d, want 1", block.Header.Height)
	}
	if len(block.Transactions) != 1 || !block.Transactions[0].IsCoinbase() {
		t.Fatal("mined block must contain coinbase transaction")
	}
	if got := block.Transactions[0].Outputs[0].Value; got != consensus.Mainnet.InitialSubsidy {
		t.Fatalf("coinbase value = %d, want %d", got, consensus.Mainnet.InitialSubsidy)
	}
}

func TestSignedTransactionLifecycle(t *testing.T) {
	store, err := Open(t.TempDir())
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

	if _, err := store.MineNext(context.Background(), sender.Address); err != nil {
		t.Fatal(err)
	}

	utxos, err := store.SpendableUTXOs(senderKeys.PubKeyHash)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := NewSignedTransaction(
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
	if err := store.AddMempoolTx(tx); err != nil {
		t.Fatal(err)
	}
	if len(store.Mempool) != 1 {
		t.Fatalf("mempool len = %d, want 1", len(store.Mempool))
	}

	if _, err := store.MineNext(context.Background(), sender.Address); err != nil {
		t.Fatal(err)
	}
	if len(store.Mempool) != 0 {
		t.Fatalf("mempool len after mining = %d, want 0", len(store.Mempool))
	}

	receiverBalance, err := store.Balance(receiver.Address)
	if err != nil {
		t.Fatal(err)
	}
	if receiverBalance.Confirmed != 10_000_000 {
		t.Fatalf("receiver balance = %d, want %d", receiverBalance.Confirmed, 10_000_000)
	}
}

func TestExpectedBitsRetargetsAfterFastInterval(t *testing.T) {
	pubKeyHash := make([]byte, 20)
	blocks := []Block{NewGenesis()}
	prev := blocks[0]

	for height := uint64(1); height <= consensus.Mainnet.RetargetInterval; height++ {
		coinbase := NewCoinbase(pubKeyHash, height, consensus.Mainnet.Subsidy(height))
		block := NewCandidate(prev, []Transaction{coinbase}, consensus.Mainnet.PowLimitBits)
		block.Header.Timestamp = blocks[0].Header.Timestamp + int64(height)
		blocks = append(blocks, block)
		prev = block
	}

	got := ExpectedBits(blocks)
	if got == consensus.Mainnet.PowLimitBits {
		t.Fatal("expected retargeted bits to differ from pow limit")
	}
	if scrypto.CompactToBig(got).Cmp(scrypto.CompactToBig(consensus.Mainnet.PowLimitBits)) >= 0 {
		t.Fatal("expected retargeted target to be harder than pow limit")
	}
}
