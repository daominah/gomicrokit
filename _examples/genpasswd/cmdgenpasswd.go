package main

import (
	"fmt"

	"github.com/daominah/gomicrokit/auth/genpasswd"
)

func main() {
	for i := 0; i < 10; i++ {
		passwd := genpasswd.GenPassword(12)
		fmt.Println(passwd)
	}
}
