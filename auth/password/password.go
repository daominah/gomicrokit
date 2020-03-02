// Package password provides password related funcs
package password

import (
	"math/rand"
	"strings"
	"time"

	"github.com/daominah/gomicrokit/gofast"
	"golang.org/x/crypto/bcrypt"
)

var (
	lowers   = strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	uppers   = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")
	digits   = strings.Split("0123456789", "")
	symbols  = strings.Split("_", "")
	allChars []string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	for _, chars := range [][]string{lowers, uppers, digits, symbols} {
		for _, char := range chars {
			allChars = append(allChars, char)
		}
	}
}

// GenRandomPassword returns a mixture of uppercase, lowercase, number and
// special characters.
func GenRandomPassword(lenPasswd int) string {
	if lenPasswd < 4 {
		lenPasswd = 4
	}
	indices := rand.Perm(lenPasswd)
	forceIndices := indices[:4]
	password := make([]string, lenPasswd)
	for i, charType := range [][]string{lowers, uppers, digits, symbols} {
		forceIndex := forceIndices[i]
		password[forceIndex] = charType[rand.Intn(len(charType))]
	}
	for i := range password {
		if gofast.IndexInts(forceIndices, i) == -1 {
			password[i] = allChars[rand.Intn(len(allChars))]
		}
	}
	return strings.Join(password, "")
}

// HashPassword is a slow hash func with salt automatically included (
// results in different call can be varied).
func HashPassword(plain string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// CheckHashPassword checks if the hashed was hashed from the plain
func CheckHashPassword(hashed string, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err != nil {
		return false
	}
	return true
}
