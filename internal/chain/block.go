package chain

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"time"

	"github.com/kyotoasuna/saracie-core/internal/consensus"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
)

type BlockHeader struct {
	Version    int32  `json:"version"`
	Height     uint64 `json:"height"`
	PrevHash   string `json:"prev_hash"`
	MerkleRoot string `json:"merkle_root"`
	Timestamp  int64  `json:"timestamp"`
	Bits       uint32 `json:"bits"`
	Nonce      uint64 `json:"nonce"`
}

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
}

func NewGenesis() Block {
	genesisTx := Transaction{
		Version: 1,
		Inputs: []TxInput{{
			TxID:      "",
			Vout:      -1,
			Signature: consensus.Mainnet.GenesisMessage,
		}},
		Outputs: []TxOutput{},
	}
	genesisTx.ID = genesisTx.Hash()

	block := Block{
		Header: BlockHeader{
			Version:   1,
			Height:    0,
			PrevHash:  "",
			Timestamp: consensus.Mainnet.GenesisTimestamp,
			Bits:      consensus.Mainnet.PowLimitBits,
			Nonce:     0,
		},
		Transactions: []Transaction{genesisTx},
	}
	block.Header.MerkleRoot = MerkleRoot(block.Transactions)
	return block
}

func NewCandidate(prev Block, txs []Transaction, bits uint32) Block {
	block := Block{
		Header: BlockHeader{
			Version:   1,
			Height:    prev.Header.Height + 1,
			PrevHash:  prev.Hash(),
			Timestamp: time.Now().Unix(),
			Bits:      bits,
			Nonce:     0,
		},
		Transactions: txs,
	}
	block.Header.MerkleRoot = MerkleRoot(block.Transactions)
	return block
}

func (b Block) HeaderBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, b.Header.Version)
	_ = binary.Write(&buf, binary.LittleEndian, b.Header.Height)
	writeString(&buf, b.Header.PrevHash)
	writeString(&buf, b.Header.MerkleRoot)
	_ = binary.Write(&buf, binary.LittleEndian, b.Header.Timestamp)
	_ = binary.Write(&buf, binary.LittleEndian, b.Header.Bits)
	_ = binary.Write(&buf, binary.LittleEndian, b.Header.Nonce)
	return buf.Bytes()
}

func writeString(buf *bytes.Buffer, s string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(s)))
	buf.WriteString(s)
}

func (b Block) HashBytes() []byte {
	return scrypto.Hash256(b.HeaderBytes())
}

func (b Block) Hash() string {
	return hex.EncodeToString(b.HashBytes())
}

func (b *Block) Mine(ctx context.Context) error {
	target := scrypto.CompactToBig(b.Header.Bits)
	if target.Sign() <= 0 {
		return errors.New("invalid pow target")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if scrypto.HashMeetsTarget(b.HashBytes(), target) {
				return nil
			}
			b.Header.Nonce++
			if b.Header.Nonce == 0 {
				b.Header.Timestamp = time.Now().Unix()
			}
		}
	}
}

func MerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return scrypto.Hash256Hex(nil)
	}

	level := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		raw, err := hex.DecodeString(tx.ID)
		if err != nil {
			raw = scrypto.Hash256([]byte(tx.ID))
		}
		level = append(level, raw)
	}

	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := left
			if i+1 < len(level) {
				right = level[i+1]
			}
			pair := append(append([]byte{}, left...), right...)
			next = append(next, scrypto.Hash256(pair))
		}
		level = next
	}

	return hex.EncodeToString(level[0])
}
