package main

import (
	"database/sql"
	"fmt"
	"log"
	
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Connect to database
	db, err := sql.Open("sqlite3", "backend/backend/openpenpal.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Hash the password 'secret'
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Update alice's password
	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hashedPassword), "alice")
	if err != nil {
		log.Fatal("Failed to update password:", err)
	}

	fmt.Println("Successfully updated alice's password to 'secret'")
	
	// Verify the hash works
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("secret"))
	if err != nil {
		log.Fatal("Password verification failed:", err)
	}
	
	fmt.Println("Password verification successful!")
}