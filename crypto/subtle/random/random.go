package random

import (
	"crypto/rand"
	"encoding/binary"
)

// GetRandomBytes randomly generates n bytes.
func GetRandomBytes(n uint32) []byte {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}
	return buf
}

// GetRandomUint32 randomly generates an unsigned 32-bit integer.
func GetRandomUint32() uint32 {
	b := GetRandomBytes(4)
	return binary.BigEndian.Uint32(b)
}
