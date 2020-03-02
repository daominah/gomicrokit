package genrsa

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestGenAndParseRsa(t *testing.T) {
	idRsa, idRsaPub := GenRsaPair()
	err := ioutil.WriteFile("key.priv", []byte(idRsa), 0644)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile("key.pub", []byte(idRsaPub), 0644)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(`saved rsa key pairs to "./key.priv" and "./key.pub"`)

	privKeyObj, pubKeyObj, err := ParseRsaPair(idRsa, idRsaPub)
	if err != nil {
		t.Error(err)
	}
	if privKeyObj.Size() != 4096/8 {
		t.Error()
	}
	if privKeyObj.PublicKey.N.Cmp(pubKeyObj.N) != 0 {
		t.Error()
	}
}
