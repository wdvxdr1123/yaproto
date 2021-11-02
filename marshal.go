package yaproto

func PutBool(buf []byte, i *int, b bool) {
	if b {
		buf[*i] = 1
	} else {
		buf[*i] = 0
	}
	*i++
}

// PutVarint serializes a varint-encoded uint64 into buf
func PutVarint(buf []byte, i *int, u uint64) {
	for u >= 0x80 {
		buf[*i] = byte(u) | 0x80
		u >>= 7
		*i++
	}
	buf[*i] = byte(u)
	*i++
}
