package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "host=localhost port=5432 user=rocalight dbname=openpenpal sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test users
	users := []string{"admin", "alice"}
	passwords := map[string]string{
		"admin": "admin123",
		"alice": "secret",
	}

	for _, username := range users {
		fmt.Printf("\nTesting user: %s\n", username)
		fmt.Println(strings.Repeat("=", 50))

		// Get user from database
		var passwordHash string
		var isActive bool
		err := db.QueryRow("SELECT password_hash, is_active FROM users WHERE username = $1", username).Scan(&passwordHash, &isActive)
		if err != nil {
			fmt.Printf("❌ Error getting user: %v\n", err)
			continue
		}

		fmt.Printf("User found, active: %v\n", isActive)
		fmt.Printf("Hash length: %d\n", len(passwordHash))

		// Test password
		password := passwords[username]
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
		if err == nil {
			fmt.Printf("✅ Password '%s' is VALID\n", password)
		} else {
			fmt.Printf("❌ Password '%s' is INVALID: %v\n", password, err)
		}
	}
}