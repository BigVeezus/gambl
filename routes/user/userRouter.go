package userRoutes

import (
	controller "gambl/controllers"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.Use(middleware.CORSMiddleware())
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.POST("/users/validate-otp", controller.ValidateOTP())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
	incomingRoutes.POST("/users/:user_id/edit", controller.EditUser())
	incomingRoutes.POST("/user/change-password", controller.ChangePassword())
}
