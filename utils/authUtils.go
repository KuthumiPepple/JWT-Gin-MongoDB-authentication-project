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

func VerifyPassword(userPasswd, existingPasswd string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(existingPasswd), []byte(userPasswd))
	check := true
	var msg string
	if err != nil {
		msg = "email or password is incorrect"
		check = false
	}
	return check, msg
}
