package userRoutes

import (
	controllers "gambl/controllers/user"

	"github.com/gin-gonic/gin"
)

// AuthRoutes function
func SetupAuthRoutes(router *gin.RouterGroup, userController *controllers.UserController) {

	auth := router.Group("/auth")

	auth.POST("/signup", userController.CreateUser())
	auth.POST("/login", userController.Login())

	// v1.GET("/games", gc.GetGames())

	// admin := protected.Use(middleware.AdminMiddleware())
	// admin.GET("/games", gameController.GetGames())
}
