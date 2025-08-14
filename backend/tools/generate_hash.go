package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	password := "secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err \!= nil {
		panic(err)
	}
	fmt.Printf("Hash for password: %s\n", string(hash))
}
