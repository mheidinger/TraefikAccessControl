package utils

import (
	"math/rand"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString returns a random string of the given length. It uses 0-9, a-z and A-Z characters.
// You should not use this for IDs, as random strings are not sortable.
func RandomString(length int) string {
	b := make([]byte, length)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := length-1, seededRand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = seededRand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
