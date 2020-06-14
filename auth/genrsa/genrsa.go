// Package genrsa provides convenient functions to generate and parse RSA key pair files
package genrsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GenRsaPair generates an RSA keypair of 4096 bit size
// (the keys is in PKCS1 form, encoded in PEM)
func GenRsaPair() (idRsa string, idRsaPub string) {
	priv, _ := rsa.GenerateKey(rand.Reader, 4096)

	privASN1 := x509.MarshalPKCS1PrivateKey(priv)
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privASN1,
		},
	)

	idRsa = string(privBytes)

	pubASN1 := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	idRsaPub = string(pubBytes)

	return idRsa, idRsaPub
}

// ParseRsaPair reads RSA key pair from files encoded in PEM
func ParseRsaPair(idRsa string, idRsaPub string) (
	privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, err error) {
	// pem.Decode can panic
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("invalid PEM input")
		}
	}()

	tmp, _ := pem.Decode([]byte(idRsa))
	privateKey, err = x509.ParsePKCS1PrivateKey(tmp.Bytes)
	if err != nil {
		return nil, nil, err
	}

	tmp, _ = pem.Decode([]byte(idRsaPub))
	publicKey, err = x509.ParsePKCS1PublicKey(tmp.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, publicKey, err
}
