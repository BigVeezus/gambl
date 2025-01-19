package gameRoutes

import (
	controllers "gambl/controllers/game"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func SetupGameRoutes(router *gin.RouterGroup, gameController *controllers.GameController) {

	games := router.Group("/games")
	protected := games.Use(middleware.Authentication())
	{
		protected.POST("/", gameController.CreateGame())
		protected.GET("/", gameController.ListGames())
		protected.GET("/:game_id", gameController.GetGame())
	}
	// v1.GET("/games", gc.GetGames())

	// admin := protected.Use(middleware.AdminMiddleware())
	// admin.GET("/games", gameController.GetGames())
}
