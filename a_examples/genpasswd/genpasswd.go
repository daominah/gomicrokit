package main

import (
	"fmt"

	"github.com/daominah/gomicrokit/auth/password"
)

func main() {
	for i := 0; i < 10; i++ {
		passwd := password.GenRandomPassword(12)
		_, _ = fmt.Println, passwd
		fmt.Println(passwd)
	}
}
