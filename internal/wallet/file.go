package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/scrypt"
)

const (
	walletVersion = 1
	scryptN       = 32768
	scryptR       = 8
	scryptP       = 1
	keyLen        = 32
)

type EncryptedWallet struct {
	Version    int    `json:"version"`
	KDF        string `json:"kdf"`
	KDFN       int    `json:"kdf_n"`
	KDFR       int    `json:"kdf_r"`
	KDFP       int    `json:"kdf_p"`
	Cipher     string `json:"cipher"`
	Salt       string `json:"salt"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	Address    string `json:"address"`
	Path       string `json:"path"`
	CreatedAt  int64  `json:"created_at"`
}

type walletPayload struct {
	Mnemonic  string `json:"mnemonic"`
	Index     uint32 `json:"index"`
	CreatedAt int64  `json:"created_at"`
}

type WalletFileInfo struct {
	Wallet    string `json:"wallet"`
	Address   string `json:"address"`
	Path      string `json:"path"`
	Created   int64  `json:"created_at"`
	Encrypted bool   `json:"encrypted"`
}

func CreateEncryptedWalletFile(path, passphrase string, index uint32) (*WalletFileInfo, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase is required")
	}

	account, err := NewAccount(index)
	if err != nil {
		return nil, err
	}
	if err := SaveEncryptedWalletFile(path, passphrase, account.Mnemonic, index); err != nil {
		return nil, err
	}

	return &WalletFileInfo{
		Wallet:    path,
		Address:   account.Address,
		Path:      account.Path,
		Created:   time.Now().Unix(),
		Encrypted: true,
	}, nil
}

func SaveEncryptedWalletFile(path, passphrase, mnemonic string, index uint32) error {
	if path == "" {
		return fmt.Errorf("wallet path is required")
	}
	if passphrase == "" {
		return fmt.Errorf("passphrase is required")
	}

	account, err := Derive(mnemonic, index)
	if err != nil {
		return err
	}
	created := time.Now().Unix()
	payload, err := json.Marshal(walletPayload{
		Mnemonic:  mnemonic,
		Index:     index,
		CreatedAt: created,
	})
	if err != nil {
		return err
	}

	salt := make([]byte, 16)
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	key, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	ciphertext := aead.Seal(nil, nonce, payload, nil)

	file := EncryptedWallet{
		Version:    walletVersion,
		KDF:        "scrypt",
		KDFN:       scryptN,
		KDFR:       scryptR,
		KDFP:       scryptP,
		Cipher:     "aes-256-gcm",
		Salt:       hex.EncodeToString(salt),
		Nonce:      hex.EncodeToString(nonce),
		Ciphertext: hex.EncodeToString(ciphertext),
		Address:    account.Address,
		Path:       account.Path,
		CreatedAt:  created,
	}

	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func LoadEncryptedWalletFile(path, passphrase string) (*walletPayload, *EncryptedWallet, error) {
	if path == "" {
		return nil, nil, fmt.Errorf("wallet path is required")
	}
	if passphrase == "" {
		return nil, nil, fmt.Errorf("passphrase is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var file EncryptedWallet
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, nil, err
	}
	if file.Version != walletVersion {
		return nil, nil, fmt.Errorf("unsupported wallet version %d", file.Version)
	}
	if file.KDF != "scrypt" || file.Cipher != "aes-256-gcm" {
		return nil, nil, fmt.Errorf("unsupported wallet crypto")
	}

	salt, err := hex.DecodeString(file.Salt)
	if err != nil {
		return nil, nil, err
	}
	nonce, err := hex.DecodeString(file.Nonce)
	if err != nil {
		return nil, nil, err
	}
	ciphertext, err := hex.DecodeString(file.Ciphertext)
	if err != nil {
		return nil, nil, err
	}

	key, err := scrypt.Key([]byte(passphrase), salt, file.KDFN, file.KDFR, file.KDFP, keyLen)
	if err != nil {
		return nil, nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("wallet decrypt failed")
	}

	var payload walletPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, nil, err
	}
	return &payload, &file, nil
}

func LoadEncryptedKeyPair(path, passphrase string) (*KeyPair, *EncryptedWallet, error) {
	payload, file, err := LoadEncryptedWalletFile(path, passphrase)
	if err != nil {
		return nil, nil, err
	}
	keyPair, err := DeriveKeyPair(payload.Mnemonic, payload.Index)
	if err != nil {
		return nil, nil, err
	}
	return keyPair, file, nil
}
