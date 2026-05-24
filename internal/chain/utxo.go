package chain

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
)

type UTXO struct {
	TxID       string `json:"txid"`
	Vout       int    `json:"vout"`
	Value      int64  `json:"value"`
	PubKeyHash string `json:"pub_key_hash"`
	Height     uint64 `json:"height"`
	Coinbase   bool   `json:"coinbase"`
}

type Balance struct {
	Address          string `json:"address"`
	Confirmed        int64  `json:"confirmed_base_units"`
	Spendable        int64  `json:"spendable_base_units"`
	ConfirmedDisplay string `json:"confirmed"`
	SpendableDisplay string `json:"spendable"`
	UTXOCount        int    `json:"utxo_count"`
}

func OutPointKey(txid string, vout int) string {
	return fmt.Sprintf("%s:%d", txid, vout)
}

func BuildUTXOSet(blocks []Block) (map[string]UTXO, error) {
	utxos := make(map[string]UTXO)

	for _, block := range blocks {
		for txIndex, tx := range block.Transactions {
			if tx.ID == "" || tx.ID != tx.Hash() {
				return nil, fmt.Errorf("invalid transaction id at height %d", block.Header.Height)
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					delete(utxos, OutPointKey(in.TxID, in.Vout))
				}
			}

			for vout, out := range tx.Outputs {
				if out.Value <= 0 {
					return nil, fmt.Errorf("invalid output value in tx %s", tx.ID)
				}
				if _, err := hex.DecodeString(out.PubKeyHash); err != nil {
					return nil, fmt.Errorf("invalid pubkey hash in tx %s: %w", tx.ID, err)
				}
				utxos[OutPointKey(tx.ID, vout)] = UTXO{
					TxID:       tx.ID,
					Vout:       vout,
					Value:      out.Value,
					PubKeyHash: out.PubKeyHash,
					Height:     block.Header.Height,
					Coinbase:   txIndex == 0 && tx.IsCoinbase(),
				}
			}
		}
	}

	return utxos, nil
}

func FilterUTXOsByPubKeyHash(utxos map[string]UTXO, pubKeyHash []byte) []UTXO {
	want := hex.EncodeToString(pubKeyHash)
	result := make([]UTXO, 0)
	for _, utxo := range utxos {
		if utxo.PubKeyHash == want {
			result = append(result, utxo)
		}
	}
	return result
}

func ValidateTransaction(tx Transaction, utxos map[string]UTXO, spent map[string]bool) (int64, error) {
	if tx.IsCoinbase() {
		return 0, errors.New("coinbase cannot be validated as regular transaction")
	}
	if tx.ID == "" || tx.ID != tx.Hash() {
		return 0, errors.New("invalid transaction id")
	}
	if len(tx.Inputs) == 0 {
		return 0, errors.New("transaction has no inputs")
	}
	if len(tx.Outputs) == 0 {
		return 0, errors.New("transaction has no outputs")
	}

	var inputTotal int64
	var outputTotal int64

	for i, out := range tx.Outputs {
		if out.Value <= 0 {
			return 0, fmt.Errorf("output %d value must be positive", i)
		}
		pubKeyHash, err := hex.DecodeString(out.PubKeyHash)
		if err != nil || len(pubKeyHash) != 20 {
			return 0, fmt.Errorf("output %d has invalid pubkey hash", i)
		}
		outputTotal += out.Value
		if outputTotal <= 0 {
			return 0, errors.New("output overflow")
		}
	}

	touched := make([]string, 0, len(tx.Inputs))
	for i, in := range tx.Inputs {
		key := OutPointKey(in.TxID, in.Vout)
		if spent[key] {
			return 0, fmt.Errorf("double spend in transaction input %d", i)
		}

		utxo, ok := utxos[key]
		if !ok {
			return 0, fmt.Errorf("missing utxo %s", key)
		}

		pubKeyBytes, err := hex.DecodeString(in.PubKey)
		if err != nil {
			return 0, fmt.Errorf("invalid public key in input %d", i)
		}
		pubKey, err := btcec.ParsePubKey(pubKeyBytes)
		if err != nil {
			return 0, fmt.Errorf("parse public key in input %d: %w", i, err)
		}
		if hex.EncodeToString(scrypto.Hash160(pubKeyBytes)) != utxo.PubKeyHash {
			return 0, fmt.Errorf("input %d public key does not match utxo", i)
		}

		sigBytes, err := hex.DecodeString(in.Signature)
		if err != nil {
			return 0, fmt.Errorf("invalid signature in input %d", i)
		}
		sig, err := ecdsa.ParseDERSignature(sigBytes)
		if err != nil {
			return 0, fmt.Errorf("parse signature in input %d: %w", i, err)
		}
		if !sig.Verify(tx.SigningHash(i), pubKey) {
			return 0, fmt.Errorf("invalid signature in input %d", i)
		}

		inputTotal += utxo.Value
		if inputTotal <= 0 {
			return 0, errors.New("input overflow")
		}
		touched = append(touched, key)
	}

	if outputTotal > inputTotal {
		return 0, errors.New("outputs exceed inputs")
	}

	for _, key := range touched {
		spent[key] = true
	}

	return inputTotal - outputTotal, nil
}
