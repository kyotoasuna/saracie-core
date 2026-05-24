package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/kyotoasuna/saracie-core/internal/consensus"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
	"github.com/kyotoasuna/saracie-core/internal/wallet"
)

type Store struct {
	DataDir string        `json:"-"`
	Path    string        `json:"-"`
	Blocks  []Block       `json:"blocks"`
	Mempool []Transaction `json:"mempool"`
}

type Status struct {
	Network      string             `json:"network"`
	Ticker       string             `json:"ticker"`
	Height       uint64             `json:"height"`
	TipHash      string             `json:"tip_hash"`
	GenesisHash  string             `json:"genesis_hash"`
	Scarcity     consensus.Scarcity `json:"scarcity"`
	BlockReward  string             `json:"block_reward"`
	Mined        string             `json:"mined"`
	Remaining    string             `json:"remaining"`
	AddressHRP   string             `json:"address_hrp"`
	HalvingEvery uint64             `json:"halving_every_blocks"`
	MempoolCount int                `json:"mempool_count"`
	Params       consensus.Params   `json:"params"`
}

func Open(dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = ".saracie"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	store := &Store{
		DataDir: dataDir,
		Path:    filepath.Join(dataDir, "chain.json"),
	}

	if _, err := os.Stat(store.Path); errors.Is(err, os.ErrNotExist) {
		store.Blocks = []Block{NewGenesis()}
		return store, store.Save()
	}

	raw, err := os.ReadFile(store.Path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, store); err != nil {
		return nil, err
	}
	if len(store.Blocks) == 0 {
		store.Blocks = []Block{NewGenesis()}
		if err := store.Save(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

func (s *Store) Save() error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, raw, 0o644)
}

func (s *Store) Tip() Block {
	return s.Blocks[len(s.Blocks)-1]
}

func (s *Store) Genesis() Block {
	return s.Blocks[0]
}

func (s *Store) Status() Status {
	tip := s.Tip()
	scarcity := consensus.Mainnet.ScarcityAt(tip.Header.Height)
	return Status{
		Network:      consensus.Mainnet.Name,
		Ticker:       consensus.Mainnet.Ticker,
		Height:       tip.Header.Height,
		TipHash:      tip.Hash(),
		GenesisHash:  s.Genesis().Hash(),
		Scarcity:     scarcity,
		BlockReward:  consensus.FormatAmount(scarcity.CurrentReward),
		Mined:        consensus.FormatAmount(scarcity.Mined),
		Remaining:    consensus.FormatAmount(scarcity.Remaining),
		AddressHRP:   consensus.Mainnet.AddressHRP,
		HalvingEvery: consensus.Mainnet.HalvingInterval,
		MempoolCount: len(s.Mempool),
		Params:       consensus.Mainnet,
	}
}

func (s *Store) AddBlock(block Block) error {
	if int(block.Header.Height) < len(s.Blocks) {
		if s.Blocks[int(block.Header.Height)].Hash() == block.Hash() {
			return nil
		}
		return fmt.Errorf("conflicting block at height %d", block.Header.Height)
	}
	if err := ValidateNextBlockWithHistory(s.Blocks, block); err != nil {
		return err
	}
	s.Blocks = append(s.Blocks, block)
	s.removeMinedTransactions(block)
	return s.Save()
}

func (s *Store) KnowsBlock(block Block) bool {
	if int(block.Header.Height) >= len(s.Blocks) {
		return false
	}
	return s.Blocks[int(block.Header.Height)].Hash() == block.Hash()
}

func (s *Store) ReplaceIfValidLonger(blocks []Block) (bool, error) {
	if len(blocks) == 0 {
		return false, nil
	}
	if err := ValidateBlocks(blocks); err != nil {
		return false, err
	}
	if ChainWork(blocks).Cmp(ChainWork(s.Blocks)) <= 0 {
		return false, nil
	}
	s.Blocks = blocks
	s.Mempool = nil
	return true, s.Save()
}

func (s *Store) MineNext(ctx context.Context, payoutAddress string) (Block, error) {
	pubKeyHash, err := wallet.AddressToPubKeyHash(payoutAddress)
	if err != nil {
		return Block{}, err
	}

	prev := s.Tip()
	height := prev.Header.Height + 1
	pending, fees, err := s.ValidMempoolTransactions()
	if err != nil {
		return Block{}, err
	}
	reward := consensus.Mainnet.Subsidy(height) + fees
	coinbase := NewCoinbase(pubKeyHash, height, reward)
	txs := append([]Transaction{coinbase}, pending...)
	block := NewCandidate(prev, txs, ExpectedBits(s.Blocks))

	if err := block.Mine(ctx); err != nil {
		return Block{}, err
	}
	if err := s.AddBlock(block); err != nil {
		return Block{}, err
	}
	return block, nil
}

func (s *Store) AddMempoolTx(tx Transaction) error {
	if tx.IsCoinbase() {
		return errors.New("coinbase transaction cannot enter mempool")
	}
	if s.HasTransaction(tx.ID) {
		return nil
	}
	if s.HasMempoolTx(tx.ID) {
		return nil
	}

	utxos, err := BuildUTXOSet(s.Blocks)
	if err != nil {
		return err
	}
	spent := make(map[string]bool)
	for _, existing := range s.Mempool {
		if _, err := ValidateTransaction(existing, utxos, spent); err != nil {
			return fmt.Errorf("existing mempool transaction invalid: %w", err)
		}
	}
	if _, err := ValidateTransaction(tx, utxos, spent); err != nil {
		return err
	}
	s.Mempool = append(s.Mempool, tx)
	return s.Save()
}

func (s *Store) HasTransaction(txid string) bool {
	if txid == "" {
		return false
	}
	for _, block := range s.Blocks {
		for _, tx := range block.Transactions {
			if tx.ID == txid {
				return true
			}
		}
	}
	return false
}

func (s *Store) HasMempoolTx(txid string) bool {
	if txid == "" {
		return false
	}
	for _, tx := range s.Mempool {
		if tx.ID == txid {
			return true
		}
	}
	return false
}

func (s *Store) ValidMempoolTransactions() ([]Transaction, int64, error) {
	utxos, err := BuildUTXOSet(s.Blocks)
	if err != nil {
		return nil, 0, err
	}

	var valid []Transaction
	var fees int64
	spent := make(map[string]bool)
	for _, tx := range s.Mempool {
		fee, err := ValidateTransaction(tx, utxos, spent)
		if err != nil {
			continue
		}
		valid = append(valid, tx)
		fees += fee
	}
	return valid, fees, nil
}

func (s *Store) Balance(address string) (Balance, error) {
	pubKeyHash, err := wallet.AddressToPubKeyHash(address)
	if err != nil {
		return Balance{}, err
	}
	confirmedUTXOs, err := s.SpendableUTXOs(pubKeyHash)
	if err != nil {
		return Balance{}, err
	}
	allUTXOs, err := BuildUTXOSet(s.Blocks)
	if err != nil {
		return Balance{}, err
	}

	var confirmed int64
	for _, utxo := range FilterUTXOsByPubKeyHash(allUTXOs, pubKeyHash) {
		confirmed += utxo.Value
	}
	var spendable int64
	for _, utxo := range confirmedUTXOs {
		spendable += utxo.Value
	}

	return Balance{
		Address:          address,
		Confirmed:        confirmed,
		Spendable:        spendable,
		ConfirmedDisplay: consensus.FormatAmount(confirmed),
		SpendableDisplay: consensus.FormatAmount(spendable),
		UTXOCount:        len(confirmedUTXOs),
	}, nil
}

func (s *Store) SpendableUTXOs(pubKeyHash []byte) ([]UTXO, error) {
	utxos, err := BuildUTXOSet(s.Blocks)
	if err != nil {
		return nil, err
	}
	spendable := FilterUTXOsByPubKeyHash(utxos, pubKeyHash)

	spentByMempool := make(map[string]bool)
	for _, tx := range s.Mempool {
		for _, in := range tx.Inputs {
			spentByMempool[OutPointKey(in.TxID, in.Vout)] = true
		}
	}

	filtered := make([]UTXO, 0, len(spendable))
	for _, utxo := range spendable {
		if !spentByMempool[OutPointKey(utxo.TxID, utxo.Vout)] {
			filtered = append(filtered, utxo)
		}
	}
	return filtered, nil
}

func (s *Store) removeMinedTransactions(block Block) {
	included := make(map[string]bool)
	for _, tx := range block.Transactions {
		included[tx.ID] = true
	}

	remaining := s.Mempool[:0]
	for _, tx := range s.Mempool {
		if !included[tx.ID] {
			remaining = append(remaining, tx)
		}
	}
	s.Mempool = remaining
}

func ValidateBlocks(blocks []Block) error {
	if len(blocks) == 0 {
		return errors.New("empty chain")
	}

	genesis := NewGenesis()
	if blocks[0].Hash() != genesis.Hash() {
		return errors.New("invalid genesis block")
	}
	if blocks[0].Header.MerkleRoot != MerkleRoot(blocks[0].Transactions) {
		return errors.New("invalid genesis merkle root")
	}

	for i := 1; i < len(blocks); i++ {
		if err := ValidateNextBlockWithHistory(blocks[:i], blocks[i]); err != nil {
			return fmt.Errorf("invalid block at index %d: %w", i, err)
		}
	}
	return nil
}

func ValidateNextBlock(prev Block, block Block) error {
	return validateNextBlockHeader(prev, block)
}

func ValidateNextBlockWithHistory(prevBlocks []Block, block Block) error {
	if len(prevBlocks) == 0 {
		return errors.New("missing previous chain")
	}
	prev := prevBlocks[len(prevBlocks)-1]
	if err := validateNextBlockHeader(prev, block); err != nil {
		return err
	}
	expectedBits := ExpectedBits(prevBlocks)
	if block.Header.Bits != expectedBits {
		return fmt.Errorf("invalid bits: got %08x want %08x", block.Header.Bits, expectedBits)
	}

	utxos, err := BuildUTXOSet(prevBlocks)
	if err != nil {
		return err
	}

	var fees int64
	spent := make(map[string]bool)
	for i, tx := range block.Transactions {
		if i == 0 {
			continue
		}
		if tx.IsCoinbase() {
			return fmt.Errorf("extra coinbase transaction at index %d", i)
		}
		fee, err := ValidateTransaction(tx, utxos, spent)
		if err != nil {
			return fmt.Errorf("invalid transaction %s: %w", tx.ID, err)
		}
		fees += fee
	}

	expectedSubsidy := consensus.Mainnet.Subsidy(block.Header.Height)
	coinbaseOut := int64(0)
	for _, out := range block.Transactions[0].Outputs {
		coinbaseOut += out.Value
	}
	if coinbaseOut > expectedSubsidy+fees {
		return fmt.Errorf("coinbase pays %d, max is %d", coinbaseOut, expectedSubsidy+fees)
	}
	if consensus.Mainnet.MinedThroughHeight(block.Header.Height) > consensus.Mainnet.MaxSupply {
		return errors.New("max supply exceeded")
	}

	return nil
}

func validateNextBlockHeader(prev Block, block Block) error {
	if block.Header.Height != prev.Header.Height+1 {
		return fmt.Errorf("invalid height: got %d want %d", block.Header.Height, prev.Header.Height+1)
	}
	if block.Header.PrevHash != prev.Hash() {
		return errors.New("invalid previous hash")
	}
	if MerkleRoot(block.Transactions) != block.Header.MerkleRoot {
		return errors.New("invalid merkle root")
	}
	if len(block.Transactions) == 0 || !block.Transactions[0].IsCoinbase() {
		return errors.New("first transaction must be coinbase")
	}
	for i, tx := range block.Transactions {
		if tx.ID == "" || tx.ID != tx.Hash() {
			return fmt.Errorf("invalid transaction id at index %d", i)
		}
	}
	if !scrypto.HashMeetsTarget(block.HashBytes(), scrypto.CompactToBig(block.Header.Bits)) {
		return errors.New("block does not satisfy proof-of-work")
	}

	return nil
}

func ExpectedBits(prevBlocks []Block) uint32 {
	if len(prevBlocks) == 0 {
		return consensus.Mainnet.PowLimitBits
	}

	prev := prevBlocks[len(prevBlocks)-1]
	interval := consensus.Mainnet.RetargetInterval
	if interval == 0 || prev.Header.Height == 0 || prev.Header.Height%interval != 0 || uint64(len(prevBlocks)) <= interval {
		return prev.Header.Bits
	}

	first := prevBlocks[len(prevBlocks)-int(interval)]
	actualTimespan := prev.Header.Timestamp - first.Header.Timestamp
	targetTimespan := int64(interval) * consensus.Mainnet.TargetBlockSeconds
	minTimespan := targetTimespan / consensus.Mainnet.MaxRetargetFactor
	maxTimespan := targetTimespan * consensus.Mainnet.MaxRetargetFactor

	if actualTimespan < minTimespan {
		actualTimespan = minTimespan
	}
	if actualTimespan > maxTimespan {
		actualTimespan = maxTimespan
	}

	oldTarget := scrypto.CompactToBig(prev.Header.Bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(actualTimespan))
	newTarget.Div(newTarget, big.NewInt(targetTimespan))

	powLimit := scrypto.CompactToBig(consensus.Mainnet.PowLimitBits)
	if newTarget.Cmp(powLimit) > 0 {
		newTarget = powLimit
	}
	return scrypto.BigToCompact(newTarget)
}

func ChainWork(blocks []Block) *big.Int {
	total := big.NewInt(0)
	for i, block := range blocks {
		if i == 0 {
			continue
		}
		target := scrypto.CompactToBig(block.Header.Bits)
		if target.Sign() <= 0 {
			continue
		}
		work := new(big.Int).Lsh(big.NewInt(1), 256)
		work.Div(work, new(big.Int).Add(target, big.NewInt(1)))
		total.Add(total, work)
	}
	return total
}
