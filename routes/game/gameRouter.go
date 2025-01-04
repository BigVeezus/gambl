package gameRoutes

import (
	"gambl/controllers/game"
	"gambl/middleware"
	"github.com/gin-gonic/gin"
)
// UserRoutes function
func SetupGameRoutes(router *gin.RouterGroup, gameController *controllers.GameController) {

	games := router.Group("/games")
	protected := games.Use(middleware.Authentication())
	// v1.Use(middleware.CORSMiddleware())
	{
		protected.POST("/", gameController.CreateGame())
		protected.GET("/:game_id", gameController.GetGame())
		protected.GET("/", gameController.ListGames())
	}
	// v1.GET("/games", gc.GetGames())

	// admin := protected.Use(middleware.AdminMiddleware())
	// admin.GET("/games", gameController.GetGames())
}
