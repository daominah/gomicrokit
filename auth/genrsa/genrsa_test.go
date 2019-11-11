package genrsa

import (
	"testing"
	"fmt"
)

func _TestGenRsa(t *testing.T) {
	GenRsaPair()
	fmt.Println(`saved rsa key pairs to "./key.priv" and "./key.pub"`)
}
