package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Test the common passwords for the different hashes in database.go
	hashes := map[string]string{
		"secret hash": "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO",
		"courier001 hash": "$2a$10$Cm0hFv7kUKfUc5Q6booKiehnQsHSFF7.4LYuqWVkgFqCYda3qqGCS",
		"courier002 hash": "$2a$10$b75vhT53SdpdtJRcf4WzrOOpLAaBRgZ9Ix.AEfrH/UngIxoxscQNm",
		"courier003 hash": "$2a$10$ClnxSMuPM6YdlWXuswYE1OjWm06yR48cdGEqp0/YP/h9OI/u2gwvm",
		"courier004 hash": "$2a$10$9V.Mbl5QqL0.tZWaJ0nTrulHIXPgeyWaex.lKrvG.r5HqDaldbd6S",
		"admin123 hash": "$2a$10$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW",
	}
	
	passwords := []string{"secret", "courier123", "courier001", "courier002", "courier003", "courier004", "admin123", "basic123", "city123", "school123", "zone123"}
	
	for hashName, hash := range hashes {
		fmt.Printf("\n=== Testing %s ===\n", hashName)
		for _, password := range passwords {
			err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			if err == nil {
				fmt.Printf("✅ Password '%s' matches %s!\n", password, hashName)
			}
		}
	}
}