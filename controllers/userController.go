package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kuthumipepple/jwt-project/database"
	"github.com/kuthumipepple/jwt-project/models"
	"github.com/kuthumipepple/jwt-project/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		user := models.User{}
		foundUser := models.User{}

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		switch err {
		case nil:
			break
		case mongo.ErrNoDocuments:
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while checking for the email and password"})
			log.Panic(err)
		}

		isPasswordValid, msg := utils.VerifyPassword(*user.Password, *foundUser.Password)
		if !isPasswordValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, err := utils.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while generating tokens"})
			log.Panic(err)
		}

		utils.UpdateAllTokens(token, refreshToken, foundUser.UserID)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.UserID}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		queriedUserId := c.Param("user_id")
		if err := utils.MatchUserTypeToUid(c, queriedUserId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		user := models.User{}
		err := userCollection.FindOne(ctx, bson.M{"user_id": queriedUserId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !utils.CheckUserType(c, "ADMIN") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized to access this resource"})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordsPerPage, err := strconv.Atoi(c.Query("recordsPerPage"))
		if err != nil || recordsPerPage < 1 {
			recordsPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIndex, err := strconv.Atoi(c.Query("startIndex"))
		if err != nil {
			startIndex = (page - 1) * recordsPerPage
		}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "null"},
				{Key: "total_count", Value: bson.D{
					{Key: "$sum", Value: 1},
				}},
				{Key: "data", Value: bson.D{
					{Key: "$push", Value: "$$ROOT"},
				}},
			}},
		}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []any{"$data", startIndex, recordsPerPage}},
				}},
			}},
		}

		cursor, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			groupStage, projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing user items"})
			return
		}

		var allUsers []bson.M
		if err = cursor.All(ctx, &allUsers); err != nil {
			log.Panic(err)
		}

		c.JSON(http.StatusOK, allUsers[0])
	}
}
