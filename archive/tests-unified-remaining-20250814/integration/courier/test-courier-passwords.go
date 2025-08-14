package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Hash from database.go for "secret" password
	secretHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO"
	
	// Test common passwords
	passwords := []string{
		"secret", "courier123", "courier001", "courier002", "courier003", "courier004",
		"admin123", "basic123", "city123", "school123", "zone123", "level1", "level2",
		"level3", "level4", "password", "123456", "test123", "courier", "password123",
	}
	
	fmt.Println("Testing secret hash (used by level1-4 couriers):", secretHash)
	fmt.Println("====================================================")
	
	for _, password := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(secretHash), []byte(password))
		if err == nil {
			fmt.Printf("✅ FOUND: Password '%s' matches the secret hash!\n", password)
		}
	}
	
	// Test generating a hash for "secret" to verify
	fmt.Println("\n======================================================")
	fmt.Println("Generating hash for 'secret' to verify:")
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err == nil {
		fmt.Printf("New hash: %s\n", string(hash))
		// Test it
		err = bcrypt.CompareHashAndPassword([]byte(secretHash), []byte("secret"))
		if err == nil {
			fmt.Println("✅ 'secret' password CONFIRMED!")
		} else {
			fmt.Println("❌ 'secret' password does not match")
		}
	}
}