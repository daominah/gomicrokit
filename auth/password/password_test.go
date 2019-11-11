package password

import (
	"fmt"
	"testing"
)

func TestGenPassword(t *testing.T) {
	for i := 0; i < 10; i++ {
		passwd := GenRandomPassword(12)
		_, _ = fmt.Println, passwd
		//fmt.Println(passwd)
	}
}

func TestHashPassword(t *testing.T) {
	hashed := HashPassword("123qwe")
	hashed2 := HashPassword("123qwe")
	//fmt.Println("hashed:", hashed)
	if len(hashed) == 0 || len(hashed2) == 0 {
		t.Error()
	}
	if hashed == hashed2 {
		t.Error()
	}
	if !CheckHashPassword(hashed, "123qwe") {
		t.Error()
	}
	if !CheckHashPassword("$2a$04$mxei3XmeAWnw68y180YtXOe/kGe5pwjpSRJELOLBS6fIa9Jsp1sgO", "123qwe") {
		t.Error()
	}
}
