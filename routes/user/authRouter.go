package userRoutes

import (
	controller "gambl/controllers"

	"github.com/gin-gonic/gin"
)

// Auth Routes function
func AuthRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.Use(middleware.CORSMiddleware())
	incomingRoutes.POST("/users/signup", controller.SignUp())
	incomingRoutes.POST("/users/login", controller.Login())
	incomingRoutes.POST("/users/resend-otp", controller.ResendOTP())
	incomingRoutes.POST("/otp", controller.TestOTP())
}
