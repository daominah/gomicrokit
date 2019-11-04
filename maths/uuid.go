package maths

import (
	"crypto/rand"
	"encoding/hex"
)

// GenUUID generates unique 128-bit ID, represent by 32-char hex encoding.
// Probability that a ID will be duplicated is close enough to zero (10**-38)
// This implement does not follow rfc4122 standard.
func GenUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	r := hex.EncodeToString(b)
	return r
}
