package userRoutes

import (
	controller "gambl/controllers"

	"github.com/gin-gonic/gin"
)

// Auth Routes function
func AuthRoutes(v1 *gin.RouterGroup) {
	v1.POST("/signup", controller.SignUp())
	v1.POST("/admin/signup", controller.SignUpAdmin())
	v1.POST("/login", controller.Login())
}
