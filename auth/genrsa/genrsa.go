package genrsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func GenRsaPair() (privStr string, pubStr string) {
	//	func main() {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	//fmt.Printf("priv %#v\n", priv)
	//fmt.Printf("Pub %#v\n", priv.PublicKey)
	//fmt.Println("N", priv.N.BitLen())

	privASN1 := x509.MarshalPKCS1PrivateKey(priv)
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privASN1,
		},
	)
	ioutil.WriteFile("key.priv", privBytes, 0644)
	privStr = string(privBytes)

	pubASN1 := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	ioutil.WriteFile("key.pub", pubBytes, 0644)
	pubStr = string(pubBytes)

	return privStr, pubStr
}
