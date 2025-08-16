package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO"
	passwords := []string{"secret", "courier123", "level1", "basic123", "123456", "password"}
	
	fmt.Println("Testing password hash:", hash)
	for _, password := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err == nil {
			fmt.Printf("✅ Password '%s' matches!\n", password)
		} else {
			fmt.Printf("❌ Password '%s' does not match\n", password)
		}
	}
}