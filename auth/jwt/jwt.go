// Package jwt provides simple usages of JSON Web Token
package jwt

import (
	"crypto/rsa"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/daominah/gomicrokit/auth/genrsa"
	"github.com/daominah/gomicrokit/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// JWTer can create a authorization token or read data from a valid token
type JWTer struct {
	privateKey          *rsa.PrivateKey
	publicKey           *rsa.PublicKey
	tokenExpireDuration time.Duration
}

// NewJWTer input is an RSA keypair in PKCS1 form, encoded in PEM.
// The keypair can be generated by run `auth/rsa/genrsa_test.go`.
func NewJWTer(idRsa string, idRsaPub string, expireHours string) (
	jwter *JWTer, err error) {
	jwter = &JWTer{}
	temp, err := strconv.ParseInt(expireHours, 10, 64)
	if err != nil {
		return nil, err
	}
	jwter.tokenExpireDuration = time.Duration(temp) * time.Hour

	jwter.privateKey, jwter.publicKey, err = genrsa.ParseRsaPair(idRsa, idRsaPub)
	if err != nil {
		return nil, err
	}
	return jwter, nil
}

// CreateAuthToken create a JWT from input data.
// Example: authInfo = struct {userId: int64, username: string}
func (j JWTer) CreateAuthToken(authInfo interface{}) string {
	claim := jwtClaim{AuthInfo: authInfo}
	claim.IssuedAt = time.Now().Unix()
	claim.ExpiresAt = claim.IssuedAt + int64(j.tokenExpireDuration.Seconds())
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodRS512, claim)
	token, err := tokenObj.SignedString(j.privateKey)
	if err != nil {
		log.Infof("error in jwt sign string: %v", err)
		return ""
	}
	return token
}

// CheckAuthToken read data from jwtToken to outPointer.
// Expired token or invalid token or invalid outPointer data type returns error
func (j JWTer) CheckAuthToken(jwtToken string, outPointer interface{}) error {
	rv := reflect.ValueOf(outPointer)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errNonPointerOutput
	}

	var claim jwtClaim
	fPubKey := func(*jwt.Token) (interface{}, error) { return j.publicKey, nil }
	_, err := jwt.ParseWithClaims(jwtToken, &claim, fPubKey)
	if err != nil {
		return err
	}
	expire := time.Unix(claim.ExpiresAt, 0)
	if expire.Before(time.Now()) {
		return errExpiredToken
	}

	marshalled, err := json.Marshal(claim.AuthInfo)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshalled, outPointer)
	return err
}

var (
	errNonPointerOutput = errors.New("output must be a pointer")
	errExpiredToken     = errors.New("token expired")
)

type jwtClaim struct {
	jwt.StandardClaims
	AuthInfo interface{}
}

// ReadFromEnv reads envs JWT_PRIVATE_PEM, JWT_PRIVATE_PEM, JWT_EXPIRE_HOURS,
// the output can be used as NewJWTer input
func ReadFromEnv() (privKeyPkcs1 string, pubKeyPkcs1 string, expireHours string) {
	privKeyPkcs1 = os.Getenv("JWT_PRIVATE_PEM")
	pubKeyPkcs1 = os.Getenv("JWT_PUBLIC_PEM")
	expireHours = os.Getenv("JWT_EXPIRE_HOURS")

	if privKeyPkcs1 == "" {
		log.Fatalf("need to set env JWT_PRIVATE_PEM")
	}
	if pubKeyPkcs1 == "" {
		log.Fatalf("need to set env JWT_PUBLIC_PEM")
	}
	if expireHours == "" {
		log.Fatalf("need to set env JWT_EXPIRE_HOURS")
	}
	return privKeyPkcs1, pubKeyPkcs1, expireHours
}
