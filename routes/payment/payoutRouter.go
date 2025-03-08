package gameRoutes

import (
	"gambl/controllers/payment"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPayoutChannelRoutes(router *gin.RouterGroup, controller *controllers.PayoutChannelController) {
	payoutChannelRouter := router.Group("/payout-channels")
	payoutChannelRouter.Use(middleware.Authentication())
	
	{
		// CRUD operations
		payoutChannelRouter.POST("", controller.CreatePayoutChannel())
		payoutChannelRouter.GET("", controller.GetPayoutChannels())
		payoutChannelRouter.GET("/:id", controller.GetPayoutChannel())
		payoutChannelRouter.PUT("/:id", controller.UpdatePayoutChannel())
		payoutChannelRouter.DELETE("/:id", controller.DeletePayoutChannel())
		
		// Utility operations
		payoutChannelRouter.POST("/:id/set-default", controller.SetDefaultPayoutChannel())
		payoutChannelRouter.GET("/has-channel", controller.HasPayoutChannel())
		
		// Bank verification
		payoutChannelRouter.GET("/banks", controller.GetBanks())
		payoutChannelRouter.POST("/verify-bank", controller.VerifyBankAccount())
	}
}