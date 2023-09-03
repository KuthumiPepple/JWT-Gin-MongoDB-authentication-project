package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kuthumipepple/jwt-project/database"
	"github.com/kuthumipepple/jwt-project/models"
	"github.com/kuthumipepple/jwt-project/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var validate = validator.New()
var userCollection = database.OpenCollection("users")

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		user := models.User{}

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		filter := bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"phone": user.Phone},
			},
		}
		count, err := userCollection.CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while checking for the email and phone number"})
			log.Panic(err)
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "this email or phone number already exists!"})
		}

		hashedPasswd := utils.HashPassword(*user.Password)
		user.Password = &hashedPasswd
		
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken, err := utils.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while generating tokens"})
			log.Panic(err)
		}
		user.Token = &token
		user.RefreshToken = &refreshToken

		insertionResult, insertionErr := userCollection.InsertOne(ctx, user)
		if insertionErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User was not created"})
			return
		}
		c.JSON(http.StatusCreated, insertionResult)
	}
}
