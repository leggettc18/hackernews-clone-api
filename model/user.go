package model

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID             uint   `gorm:"primary_key" json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Links          []Link `json:"links"`
	Votes          []Vote `json:"votes"`
	HashedPassword []byte `json:"-"`
}

// ComparePasswordHash takes a password hash and a plaintext password and returns true
// if the plaintext password hashes into the password hash.
func ComparePasswordHash(hashedPassword, givenPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, givenPassword)
	return err == nil
}
