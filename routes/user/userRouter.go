package userRoutes

import (
	controllers "gambl/controllers/user"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func SetupUserRoutes(router *gin.RouterGroup, userController *controllers.UserController) {

	users := router.Group("/users")
	protected := users.Use(middleware.Authentication())

	protected.GET("/:user_id", userController.GetUserId())
	protected.POST("/users/:user_id/edit", userController.EditUser())

	admin := users.Use(middleware.AdminMiddleware())
	admin.GET("/", userController.GetUsers())
	admin.POST("/:user_id/updateUserType", userController.UpdateUserType())
}
