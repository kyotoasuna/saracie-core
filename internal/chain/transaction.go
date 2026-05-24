package chain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
)

type TxInput struct {
	TxID      string `json:"txid"`
	Vout      int    `json:"vout"`
	Signature string `json:"signature,omitempty"`
	PubKey    string `json:"pub_key,omitempty"`
}

type TxOutput struct {
	Value      int64  `json:"value"`
	PubKeyHash string `json:"pub_key_hash"`
}

type Transaction struct {
	ID      string     `json:"id"`
	Version int32      `json:"version"`
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
	Lock    uint32     `json:"lock_time"`
}

type signingPayload struct {
	InputIndex  int         `json:"input_index"`
	Transaction Transaction `json:"transaction"`
}

func NewCoinbase(toPubKeyHash []byte, height uint64, value int64) Transaction {
	tx := Transaction{
		Version: 1,
		Inputs: []TxInput{{
			TxID:      "",
			Vout:      -1,
			Signature: fmt.Sprintf("Saracie coinbase height %d", height),
		}},
		Outputs: []TxOutput{{
			Value:      value,
			PubKeyHash: hex.EncodeToString(toPubKeyHash),
		}},
	}
	tx.ID = tx.Hash()
	return tx
}

func NewSignedTransaction(utxos []UTXO, priv *btcec.PrivateKey, pubKey []byte, toPubKeyHash, changePubKeyHash []byte, amount, fee int64) (Transaction, error) {
	if amount <= 0 {
		return Transaction{}, fmt.Errorf("amount must be positive")
	}
	if fee < 0 {
		return Transaction{}, fmt.Errorf("fee cannot be negative")
	}

	target := amount + fee
	var selected []UTXO
	var inputTotal int64
	for _, utxo := range utxos {
		selected = append(selected, utxo)
		inputTotal += utxo.Value
		if inputTotal >= target {
			break
		}
	}
	if inputTotal < target {
		return Transaction{}, fmt.Errorf("insufficient funds")
	}

	tx := Transaction{
		Version: 1,
		Inputs:  make([]TxInput, len(selected)),
		Outputs: []TxOutput{{
			Value:      amount,
			PubKeyHash: hex.EncodeToString(toPubKeyHash),
		}},
	}

	change := inputTotal - target
	if change > 0 {
		tx.Outputs = append(tx.Outputs, TxOutput{
			Value:      change,
			PubKeyHash: hex.EncodeToString(changePubKeyHash),
		})
	}

	pubHex := hex.EncodeToString(pubKey)
	for i, utxo := range selected {
		tx.Inputs[i] = TxInput{
			TxID:   utxo.TxID,
			Vout:   utxo.Vout,
			PubKey: pubHex,
		}
	}

	for i := range tx.Inputs {
		sig := ecdsa.Sign(priv, tx.SigningHash(i))
		tx.Inputs[i].Signature = hex.EncodeToString(sig.Serialize())
	}

	tx.ID = tx.Hash()
	return tx, nil
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].TxID == "" && tx.Inputs[0].Vout == -1
}

func (tx Transaction) SigningHash(inputIndex int) []byte {
	copyTx := tx
	copyTx.ID = ""
	copyTx.Inputs = append([]TxInput(nil), tx.Inputs...)
	for i := range copyTx.Inputs {
		copyTx.Inputs[i].Signature = ""
		copyTx.Inputs[i].PubKey = ""
	}

	encoded, _ := json.Marshal(signingPayload{
		InputIndex:  inputIndex,
		Transaction: copyTx,
	})
	return scrypto.Hash256(encoded)
}

func (tx Transaction) Hash() string {
	copyTx := tx
	copyTx.ID = ""
	encoded, _ := json.Marshal(copyTx)
	return scrypto.Hash256Hex(encoded)
}
