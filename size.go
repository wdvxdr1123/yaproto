package yaproto

import (
	"math/bits"
)

// VarintSize returns the encoded size of an encoded varint.
func VarintSize(x uint64) int {
	return (bits.Len64(x|1) + 6) / 7
}

// ZigZagSize returns the encoded size of an encoded zig-zag value.
func ZigZagSize(x uint64) int {
	return VarintSize((x << 1) ^ uint64(int64(x)>>63))
}
