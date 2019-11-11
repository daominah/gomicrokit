package genpasswd

import (
	"fmt"
	"testing"
)

func TestGenPassword(t *testing.T) {
	for i := 0; i < 10; i++ {
		passwd := GenPassword(12)
		_, _ = fmt.Println, passwd
		//fmt.Println(passwd)
	}
}
