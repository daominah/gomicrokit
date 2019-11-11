package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"reflect"
	"strconv"
	"time"
	"encoding/json"

	"github.com/daominah/gomicrokit/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

var (
	ErrPemDecode        = errors.New("cannot decode key pem")
	ErrNonPointerOutput = errors.New("non pointer output param")
)

type JWTer struct {
	privateKey          *rsa.PrivateKey
	publicKey           *rsa.PublicKey
	tokenExpireDuration time.Duration
}

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

// Input RSA keys is in Public-Key Cryptography Standards (PKCS) ASN.1 DER form,
// Example can be generated by run `auth/genrsa/genrsa_test.go`.
func NewJWTer(privKeyPkcs1 string, pubKeyPkcs1 string, expireHours string) (
	jwter *JWTer, err error) {
	// pem.Decode can panic
	defer func() {
		if r := recover(); r != nil {
			err = ErrPemDecode
		}
	}()

	jwter = &JWTer{}
	temp, err := strconv.ParseInt(expireHours, 10, 64)
	if err != nil {
		return nil, err
	}
	jwter.tokenExpireDuration = time.Duration(temp) * time.Hour

	tmp, _ := pem.Decode([]byte(privKeyPkcs1))
	jwter.privateKey, err = x509.ParsePKCS1PrivateKey(tmp.Bytes)
	if err != nil {
		return nil, ErrPemDecode
	}

	tmp, _ = pem.Decode([]byte(pubKeyPkcs1))
	jwter.publicKey, err = x509.ParsePKCS1PublicKey(tmp.Bytes)
	if err != nil {
		return nil, ErrPemDecode
	}
	return jwter, nil
}

type JwtClaim struct {
	jwt.StandardClaims
	AuthInfo interface{}
}

func (j JWTer) CreateAuthToken(authInfo interface{}) string {
	claim := JwtClaim{AuthInfo: authInfo}
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

func (j JWTer) CheckAuthToken(jwtToken string, outPointer interface{}) error {
	rv := reflect.ValueOf(outPointer)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrNonPointerOutput
	}

	var claim JwtClaim
	fPubKey := func(*jwt.Token) (interface{}, error) { return j.publicKey, nil }
	_, err := jwt.ParseWithClaims(jwtToken, &claim, fPubKey)
	if err != nil {
		return err
	}

	marshalled, err := json.Marshal(claim.AuthInfo)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshalled, outPointer)
	return err
}
