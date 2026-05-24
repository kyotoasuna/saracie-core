package wallet

import (
	"strings"
	"testing"
)

func TestNewAccountAndAddressDecode(t *testing.T) {
	account, err := NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(account.Address, "sar1") {
		t.Fatalf("address %q does not use sar1 prefix", account.Address)
	}
	pubKeyHash, err := AddressToPubKeyHash(account.Address)
	if err != nil {
		t.Fatal(err)
	}
	if len(pubKeyHash) != 20 {
		t.Fatalf("pubkey hash len = %d, want 20", len(pubKeyHash))
	}
}

func TestDeriveIsDeterministic(t *testing.T) {
	account, err := NewAccount(0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Derive(account.Mnemonic, 0)
	if err != nil {
		t.Fatal(err)
	}
	if account.Address != again.Address {
		t.Fatalf("derived address mismatch: %s != %s", account.Address, again.Address)
	}
}
