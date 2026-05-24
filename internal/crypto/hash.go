package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

func Hash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

func Hash256Hex(data []byte) string {
	return hex.EncodeToString(Hash256(data))
}

func Hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	r := ripemd160.New()
	_, _ = r.Write(sha[:])
	return r.Sum(nil)
}

func CompactToBig(compact uint32) *big.Int {
	size := compact >> 24
	word := compact & 0x007fffff

	var result *big.Int
	if size <= 3 {
		word >>= 8 * (3 - size)
		result = big.NewInt(int64(word))
	} else {
		result = big.NewInt(int64(word))
		result.Lsh(result, uint(8*(size-3)))
	}

	if compact&0x00800000 != 0 {
		result.Neg(result)
	}
	return result
}

func BigToCompact(n *big.Int) uint32 {
	if n.Sign() == 0 {
		return 0
	}

	bytes := n.Bytes()
	size := uint32(len(bytes))
	var compact uint32
	if size <= 3 {
		compact = uint32(n.Int64() << uint(8*(3-size)))
	} else {
		compact = uint32(bytes[0])<<16 | uint32(bytes[1])<<8 | uint32(bytes[2])
	}

	if compact&0x00800000 != 0 {
		compact >>= 8
		size++
	}

	compact |= size << 24
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

func HashMeetsTarget(hash []byte, target *big.Int) bool {
	n := new(big.Int).SetBytes(hash)
	return n.Cmp(target) <= 0
}
