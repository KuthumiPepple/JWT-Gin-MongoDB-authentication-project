package utils

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kuthumipepple/jwt-project/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UserID    string
	UserType  string
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("JWT_SECRET_KEY")
var userCollection = database.OpenCollection("users")

func GenerateAllTokens(email, firstName, lastName, userType, uid string) (signedToken, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    uid,
		UserType:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(24) * time.Hour)),
		},
	}

	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(168) * time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
	}

	return token, refreshToken, err
}

func UpdateAllTokens(signedToken, signedRefreshToken, uid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshtoken", Value: signedRefreshToken})
	updateObj = append(updateObj, bson.E{Key: "updatedat", Value: time.Now()})

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	filter := bson.M{"userid": uid}
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)
	if err != nil {
		log.Panic(err)
	}

}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "the token is invalid!"
		return
	}

	if claims.ExpiresAt.Local().Unix() < time.Now().Unix() {
		msg = "token is expired!"
		return
	}
	return
}
