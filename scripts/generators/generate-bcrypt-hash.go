package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := []string{"secret", "admin123"}
	
	for _, password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Printf("Password: %s\n", password)
		fmt.Printf("Hash: %s\n", string(hash))
		
		// Verify the hash
		err = bcrypt.CompareHashAndPassword(hash, []byte(password))
		if err == nil {
			fmt.Println("✅ Verification: Valid")
		} else {
			fmt.Println("❌ Verification: Invalid")
		}
		fmt.Println()
	}
}