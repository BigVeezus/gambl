package userRoutes

import (
	controller "gambl/controllers"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func UserRoutes(v1 *gin.RouterGroup) {
	// v1.Use(middleware.CORSMiddleware())
	v1.Use(middleware.Authentication())
	v1.GET("/users/:user_id", controller.GetUser())
	v1.POST("/users/:user_id/edit", controller.EditUser())

	v1.Use(middleware.AdminMiddleware())
	v1.GET("/users", controller.GetUsers())
	v1.POST("/users/:user_id/updateUserType", controller.UpdateUserType())
}

// func SetupUserRoutes(router *gin.RouterGroup) {

// 	users := router.Group("/users")
// 	protected := users.Use(middleware.Authentication())
// 	// v1.Use(middleware.CORSMiddleware())
// 	{
// 		protected.GET("/:user_id", controller.GetUser())
// 		protected.PUT("/:user_id/edit", controller.EditUser())
// 	}
// 	// v1.GET("/games", gc.GetGames())

// 	admin := protected.Use(middleware.AdminMiddleware())
// 	{
// 		admin.GET("/users", controller.GetUsers())
// 		admin.POST("/:user_id/updateUserType", controller.UpdateUserType())
// 	}
// 	// admin.GET("/games", gameController.GetGames())
// }
