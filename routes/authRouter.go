package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kuthumipepple/jwt-project/controllers"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/users/signup", controllers.Signup())
	router.POST("/users/login", controllers.Login())
}
