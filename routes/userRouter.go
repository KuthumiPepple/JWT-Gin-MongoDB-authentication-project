package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kuthumipepple/jwt-project/controllers"
	"github.com/kuthumipepple/jwt-project/middleware"
)

func UserRoutes(router *gin.Engine) {
	router.Use(middleware.Authenticate())
	router.GET("/users/:user_id", controllers.GetUser())
	router.GET("/users", controllers.GetUsers())
}
