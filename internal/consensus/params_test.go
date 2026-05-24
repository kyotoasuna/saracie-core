package consensus

import "testing"

func TestSubsidyHalving(t *testing.T) {
	if got := Mainnet.Subsidy(0); got != 0 {
		t.Fatalf("genesis subsidy = %d, want 0", got)
	}
	if got := Mainnet.Subsidy(1); got != Mainnet.InitialSubsidy {
		t.Fatalf("height 1 subsidy = %d, want %d", got, Mainnet.InitialSubsidy)
	}
	if got := Mainnet.Subsidy(Mainnet.HalvingInterval); got != Mainnet.InitialSubsidy {
		t.Fatalf("last first-era subsidy = %d, want %d", got, Mainnet.InitialSubsidy)
	}
	if got := Mainnet.Subsidy(Mainnet.HalvingInterval + 1); got != Mainnet.InitialSubsidy/2 {
		t.Fatalf("first second-era subsidy = %d, want %d", got, Mainnet.InitialSubsidy/2)
	}
}

func TestSupplyDoesNotExceedCap(t *testing.T) {
	mined := Mainnet.MinedThroughHeight(Mainnet.HalvingInterval * 80)
	if mined > Mainnet.MaxSupply {
		t.Fatalf("mined supply %d exceeds cap %d", mined, Mainnet.MaxSupply)
	}
}
