package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := []string{"secret", "courier123"}
	
	for _, password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", password, err)
			continue
		}
		fmt.Printf("Password: %s\nHash: %s\n\n", password, string(hash))
	}
}