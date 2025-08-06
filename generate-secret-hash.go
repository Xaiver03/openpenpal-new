package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "secret"
	
	// Generate hash for "secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hash))
	
	// Verify it works
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == nil {
		fmt.Println("✅ Hash verification successful!")
	} else {
		fmt.Printf("❌ Hash verification failed: %v\n", err)
	}
	
	// Test against the hash in database.go
	dbHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO"
	fmt.Printf("\nTesting against database hash: %s\n", dbHash)
	err = bcrypt.CompareHashAndPassword([]byte(dbHash), []byte(password))
	if err == nil {
		fmt.Println("✅ Database hash matches 'secret'!")
	} else {
		fmt.Printf("❌ Database hash does not match 'secret': %v\n", err)
	}
}