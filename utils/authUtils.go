package utils

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(passwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), 12)
	if err != nil {
		log.Panic(err)
	}
	return string(hash)
}
