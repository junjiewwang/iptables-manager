package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate hash for admin123
	hash1, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Generate hash for user123
	hash2, err := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	fmt.Printf("admin123 hash: %s\n", hash1)
	fmt.Printf("user123 hash: %s\n", hash2)
}
