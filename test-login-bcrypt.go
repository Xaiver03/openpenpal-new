package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to database
	connStr := "host=localhost port=5432 user=rocalight dbname=openpenpal sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Query for courier_level1
	var passwordHash string
	var username string
	err = db.QueryRow("SELECT username, password_hash FROM users WHERE username = $1", "courier_level1").Scan(&username, &passwordHash)
	if err != nil {
		fmt.Printf("Error querying user: %v\n", err)
		return
	}

	fmt.Printf("User: %s\n", username)
	fmt.Printf("Current hash: %s\n", passwordHash)

	// Test with "secret"
	password := "secret"
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		fmt.Printf("❌ Password verification failed: %v\n", err)
		
		// Generate new hash
		newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		fmt.Printf("\nNew hash to use: %s\n", newHash)
		
		// Update in database
		_, err = db.Exec("UPDATE users SET password_hash = $1 WHERE username = $2", string(newHash), "courier_level1")
		if err != nil {
			fmt.Printf("Error updating password: %v\n", err)
		} else {
			fmt.Println("✅ Password updated successfully")
		}
	} else {
		fmt.Println("✅ Password verification successful")
	}
}