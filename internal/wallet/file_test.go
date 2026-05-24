package wallet

import (
	"path/filepath"
	"testing"
)

func TestEncryptedWalletFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saracie.wallet")
	info, err := CreateEncryptedWalletFile(path, "correct horse battery staple", 0)
	if err != nil {
		t.Fatal(err)
	}
	if info.Address == "" {
		t.Fatal("missing wallet address")
	}

	keyPair, file, err := LoadEncryptedKeyPair(path, "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if keyPair.Account.Address != info.Address {
		t.Fatalf("address = %s, want %s", keyPair.Account.Address, info.Address)
	}
	if file.Address != info.Address {
		t.Fatalf("file address = %s, want %s", file.Address, info.Address)
	}
}

func TestEncryptedWalletRejectsWrongPassphrase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saracie.wallet")
	if _, err := CreateEncryptedWalletFile(path, "right", 0); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadEncryptedKeyPair(path, "wrong"); err == nil {
		t.Fatal("expected decrypt failure")
	}
}
