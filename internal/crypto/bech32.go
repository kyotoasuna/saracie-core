package crypto

import (
	"errors"
	"fmt"
	"strings"
)

const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func bech32Polymod(values []byte) uint32 {
	chk := uint32(1)
	generator := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

	for _, value := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(value)
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}

	return chk
}

func bech32HRPExpand(hrp string) []byte {
	out := make([]byte, 0, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		out = append(out, hrp[i]>>5)
	}
	out = append(out, 0)
	for i := 0; i < len(hrp); i++ {
		out = append(out, hrp[i]&31)
	}
	return out
}

func bech32CreateChecksum(hrp string, data []byte) []byte {
	values := append(bech32HRPExpand(hrp), data...)
	values = append(values, []byte{0, 0, 0, 0, 0, 0}...)
	polymod := bech32Polymod(values) ^ 1

	checksum := make([]byte, 6)
	for i := 0; i < 6; i++ {
		checksum[i] = byte((polymod >> uint(5*(5-i))) & 31)
	}
	return checksum
}

func bech32VerifyChecksum(hrp string, data []byte) bool {
	values := append(bech32HRPExpand(hrp), data...)
	return bech32Polymod(values) == 1
}

func Bech32Encode(hrp string, data []byte) (string, error) {
	if hrp == "" {
		return "", errors.New("empty bech32 hrp")
	}
	if strings.ToLower(hrp) != hrp {
		return "", errors.New("bech32 hrp must be lowercase")
	}

	combined := append(data, bech32CreateChecksum(hrp, data)...)
	var b strings.Builder
	b.WriteString(hrp)
	b.WriteByte('1')
	for _, value := range combined {
		if value >= 32 {
			return "", fmt.Errorf("invalid bech32 data value %d", value)
		}
		b.WriteByte(charset[value])
	}
	return b.String(), nil
}

func Bech32Decode(input string) (string, []byte, error) {
	if input == "" {
		return "", nil, errors.New("empty bech32 string")
	}

	if strings.ToLower(input) != input && strings.ToUpper(input) != input {
		return "", nil, errors.New("mixed case bech32 string")
	}

	input = strings.ToLower(input)
	separator := strings.LastIndexByte(input, '1')
	if separator < 1 || separator+7 > len(input) {
		return "", nil, errors.New("invalid bech32 separator")
	}

	hrp := input[:separator]
	raw := input[separator+1:]
	data := make([]byte, len(raw))
	for i := range raw {
		idx := strings.IndexByte(charset, raw[i])
		if idx < 0 {
			return "", nil, fmt.Errorf("invalid bech32 character %q", raw[i])
		}
		data[i] = byte(idx)
	}

	if !bech32VerifyChecksum(hrp, data) {
		return "", nil, errors.New("invalid bech32 checksum")
	}
	return hrp, data[:len(data)-6], nil
}

func ConvertBits(data []byte, from, to uint, pad bool) ([]byte, error) {
	var acc uint
	var bits uint
	maxv := uint((1 << to) - 1)
	maxAcc := uint((1 << (from + to - 1)) - 1)
	ret := make([]byte, 0, len(data)*int(from)/int(to))

	for _, value := range data {
		v := uint(value)
		if v>>from != 0 {
			return nil, errors.New("invalid value while converting bits")
		}
		acc = ((acc << from) | v) & maxAcc
		bits += from
		for bits >= to {
			bits -= to
			ret = append(ret, byte((acc>>bits)&maxv))
		}
	}

	if pad {
		if bits > 0 {
			ret = append(ret, byte((acc<<(to-bits))&maxv))
		}
	} else if bits >= from || ((acc<<(to-bits))&maxv) != 0 {
		return nil, errors.New("invalid incomplete group")
	}

	return ret, nil
}

func EncodeSegWitAddress(hrp string, version byte, program []byte) (string, error) {
	if version > 16 {
		return "", errors.New("invalid witness version")
	}
	converted, err := ConvertBits(program, 8, 5, true)
	if err != nil {
		return "", err
	}
	data := append([]byte{version}, converted...)
	return Bech32Encode(hrp, data)
}

func DecodeSegWitAddress(address string) (string, byte, []byte, error) {
	hrp, data, err := Bech32Decode(address)
	if err != nil {
		return "", 0, nil, err
	}
	if len(data) == 0 {
		return "", 0, nil, errors.New("missing witness version")
	}
	version := data[0]
	program, err := ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return "", 0, nil, err
	}
	return hrp, version, program, nil
}
