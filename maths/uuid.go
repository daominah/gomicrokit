package maths

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// This implement does not follow rfc4122 standard,
// but this follows 8-4-4-4-12 layout
func GenUUIDWithHyphen() string {
	id := GenUUID()
	if len(id) != 32 { // just to be safe
		return ""
	}
	timeLow := id[:8]
	timeMid := id[8:12]
	timeHiAndVersion := id[12:16]
	clockSeq := id[16:20]
	node := id[20:]
	return fmt.Sprintf("%v-%v-%v-%v-%v",
		timeLow, timeMid, timeHiAndVersion, clockSeq, node)
}
