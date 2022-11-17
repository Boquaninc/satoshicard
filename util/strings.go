package util

import (
	"math/rand"
	"time"
	"unsafe"
)

const (
	LETTER_BYTES    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	LETTER_IDX_BITS = 6                      // 6 bits to represent a letter index
	LETTER_IDX_MASK = 1<<LETTER_IDX_BITS - 1 // All 1-bits, as many as letterIdxBits
	LETTER_IDX_MAX  = 63 / LETTER_IDX_BITS   // # of letter indices fitting in 63 bits
)

var gSrc = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, gSrc.Int63(), LETTER_IDX_MAX; i >= 0; {
		if remain == 0 {
			cache, remain = gSrc.Int63(), LETTER_IDX_MAX
		}
		if idx := int(cache & LETTER_IDX_MASK); idx < len(LETTER_BYTES) {
			b[i] = LETTER_BYTES[idx]
			i--
		}
		cache >>= LETTER_IDX_MASK
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
