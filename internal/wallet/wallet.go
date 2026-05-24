package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/kyotoasuna/saracie-core/internal/consensus"
	scrypto "github.com/kyotoasuna/saracie-core/internal/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

type Account struct {
	Mnemonic   string `json:"mnemonic,omitempty"`
	Path       string `json:"path"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PubKeyHash string `json:"pub_key_hash"`
}

type KeyPair struct {
	Account    Account
	PrivateKey *btcec.PrivateKey
	PublicKey  []byte
	PubKeyHash []byte
}

func NewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func Derive(mnemonic string, index uint32) (*Account, error) {
	keyPair, err := DeriveKeyPair(mnemonic, index)
	if err != nil {
		return nil, err
	}
	account := keyPair.Account
	return &account, nil
}

func DeriveKeyPair(mnemonic string, index uint32) (*KeyPair, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	key := master
	for _, child := range []uint32{
		bip32.FirstHardenedChild + 84,
		bip32.FirstHardenedChild + consensus.Mainnet.DerivationCoinType,
		bip32.FirstHardenedChild,
		0,
		index,
	} {
		key, err = key.NewChildKey(child)
		if err != nil {
			return nil, err
		}
	}

	priv, pub := btcec.PrivKeyFromBytes(key.Key)
	pubBytes := pub.SerializeCompressed()
	pubHash := scrypto.Hash160(pubBytes)
	address, err := scrypto.EncodeSegWitAddress(consensus.Mainnet.AddressHRP, 0, pubHash)
	if err != nil {
		return nil, err
	}

	account := Account{
		Path:       fmt.Sprintf("m/84'/%d'/0'/0/%d", consensus.Mainnet.DerivationCoinType, index),
		Address:    address,
		PublicKey:  hex.EncodeToString(pubBytes),
		PubKeyHash: hex.EncodeToString(pubHash),
	}

	return &KeyPair{
		Account:    account,
		PrivateKey: priv,
		PublicKey:  pubBytes,
		PubKeyHash: pubHash,
	}, nil
}

func NewAccount(index uint32) (*Account, error) {
	mnemonic, err := NewMnemonic()
	if err != nil {
		return nil, err
	}
	account, err := Derive(mnemonic, index)
	if err != nil {
		return nil, err
	}
	account.Mnemonic = mnemonic
	return account, nil
}

func AddressToPubKeyHash(address string) ([]byte, error) {
	hrp, version, program, err := scrypto.DecodeSegWitAddress(address)
	if err != nil {
		return nil, err
	}
	if hrp != consensus.Mainnet.AddressHRP {
		return nil, fmt.Errorf("wrong address prefix %q", hrp)
	}
	if version != 0 {
		return nil, fmt.Errorf("unsupported address version %d", version)
	}
	if len(program) != 20 {
		return nil, fmt.Errorf("invalid public key hash length %d", len(program))
	}
	return program, nil
}
