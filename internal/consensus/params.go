package consensus

import "fmt"

const Coin int64 = 100_000_000

type Params struct {
	Name                  string `json:"name"`
	Ticker                string `json:"ticker"`
	AddressHRP            string `json:"address_hrp"`
	MaxSupply             int64  `json:"max_supply_base_units"`
	InitialSubsidy        int64  `json:"initial_subsidy_base_units"`
	HalvingInterval       uint64 `json:"halving_interval_blocks"`
	RetargetInterval      uint64 `json:"retarget_interval_blocks"`
	MaxRetargetFactor     int64  `json:"max_retarget_factor"`
	TargetBlockSeconds    int64  `json:"target_block_seconds"`
	PowLimitBits          uint32 `json:"pow_limit_bits"`
	GenesisTimestamp      int64  `json:"genesis_timestamp"`
	GenesisMessage        string `json:"genesis_message"`
	DerivationCoinType    uint32 `json:"derivation_coin_type"`
	DefaultP2PPort        int    `json:"default_p2p_port"`
	DefaultRPCPort        int    `json:"default_rpc_port"`
	DefaultHTTPStatusPort int    `json:"default_http_status_port"`
}

var Mainnet = Params{
	Name:                  "Saracie",
	Ticker:                "SRCE",
	AddressHRP:            "sar",
	MaxSupply:             210_000 * Coin,
	InitialSubsidy:        39_954_337,
	HalvingInterval:       262_800,
	RetargetInterval:      60,
	MaxRetargetFactor:     4,
	TargetBlockSeconds:    60,
	PowLimitBits:          0x1f00ffff,
	GenesisTimestamp:      1_779_580_800,
	GenesisMessage:        "Saracie SRCE mainnet launch by Kyoto Asuna - fair launch, zero premine - 2026-05-24",
	DerivationCoinType:    7331,
	DefaultP2PPort:        7337,
	DefaultRPCPort:        7338,
	DefaultHTTPStatusPort: 7339,
}

type Scarcity struct {
	Height               uint64 `json:"height"`
	RewardEra            uint64 `json:"reward_era"`
	CurrentReward        int64  `json:"current_reward_base_units"`
	Mined                int64  `json:"mined_base_units"`
	Remaining            int64  `json:"remaining_base_units"`
	PercentIssued        string `json:"percent_issued"`
	BlocksUntilHalving   uint64 `json:"blocks_until_halving"`
	EstimatedSecondsLeft int64  `json:"estimated_seconds_left"`
}

func (p Params) Subsidy(height uint64) int64 {
	if height == 0 {
		return 0
	}

	era := (height - 1) / p.HalvingInterval
	if era >= 63 {
		return 0
	}

	reward := p.InitialSubsidy >> era
	if reward < 0 {
		return 0
	}
	return reward
}

func (p Params) RewardEra(height uint64) uint64 {
	if height == 0 {
		return 0
	}
	return (height - 1) / p.HalvingInterval
}

func (p Params) MinedThroughHeight(height uint64) int64 {
	var total int64
	var start uint64 = 1

	for start <= height {
		reward := p.Subsidy(start)
		if reward == 0 {
			break
		}

		eraEnd := ((start-1)/p.HalvingInterval + 1) * p.HalvingInterval
		if eraEnd > height {
			eraEnd = height
		}

		blocks := int64(eraEnd - start + 1)
		total += reward * blocks
		if total >= p.MaxSupply {
			return p.MaxSupply
		}

		start = eraEnd + 1
	}

	return total
}

func (p Params) ScarcityAt(height uint64) Scarcity {
	mined := p.MinedThroughHeight(height)
	remaining := p.MaxSupply - mined
	if remaining < 0 {
		remaining = 0
	}

	var blocksUntil uint64
	if height == 0 {
		blocksUntil = p.HalvingInterval
	} else {
		nextHalving := (p.RewardEra(height) + 1) * p.HalvingInterval
		if nextHalving >= height {
			blocksUntil = nextHalving - height + 1
		}
	}

	percent := float64(mined) * 100 / float64(p.MaxSupply)

	return Scarcity{
		Height:               height,
		RewardEra:            p.RewardEra(height),
		CurrentReward:        p.Subsidy(height + 1),
		Mined:                mined,
		Remaining:            remaining,
		PercentIssued:        fmt.Sprintf("%.8f", percent),
		BlocksUntilHalving:   blocksUntil,
		EstimatedSecondsLeft: int64(blocksUntil) * p.TargetBlockSeconds,
	}
}

func FormatAmount(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%08d", sign, v/Coin, v%Coin)
}
