package gameRoutes

import (
	controllers "gambl/controllers/stake"
	"gambl/core/payout"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func SetupStakeRoutes(router *gin.RouterGroup, stakeController *controllers.StakeController, payoutService payout.PayoutChannelService) {

	stakes := router.Group("/stakes")
	protected := stakes.Use(middleware.Authentication())
	// Apply the payout channel middleware to routes that require staking
	createStakeRoute := stakes.Group("/")
	createStakeRoute.Use(middleware.RequirePayoutChannel(payoutService))
	{
		createStakeRoute.POST("", stakeController.PlaceStake())
	}
	{
		protected.GET("/", stakeController.ListStakes())
		protected.GET("/:stake_id", stakeController.GetStake())
	}
	// v1.GET("/stakes", gc.GetStakes())

	// admin := protected.Use(middleware.AdminMiddleware())
	// admin.GET("/stakes", stakeController.GetStakes())
}
