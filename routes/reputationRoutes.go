package userRoutes

import (
	controller "gambl/controllers"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func RepRoutes(v1 *gin.RouterGroup) {
	// v1.Use(middleware.CORSMiddleware())
	v1.Use(middleware.Authentication())
	v1.Use(middleware.AdminMiddleware())
	v1.POST("/reputations", controller.CreateReputation())
	v1.GET("/reputations", controller.GetAllReputations())
}
