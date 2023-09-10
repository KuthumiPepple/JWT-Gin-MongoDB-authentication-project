package utils

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
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

func MatchUserTypeToUid(c *gin.Context, queriedUserId string) error {
	userType := c.GetString("user_type")
	userId := c.GetString("user_id")
	// usertype USER cannot get the data of another user
	if userType == "USER" && userId != queriedUserId {
		return errors.New("unauthorized to access this resource")
	}
	return nil
}
