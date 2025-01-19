package gameRoutes

import (
	// controllers "gambl/controllers/payment"
	controllers "gambl/controllers/payment"
	"gambl/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func SetupPaymentsRoutes(router *gin.RouterGroup, paymentController *controllers.PayoutController) {

	payments := router.Group("/payment")

	payments.POST("/paystack/webhook", paymentController.FlutterwaveWebhook())

	protected := payments.Use(middleware.Authentication())
	{
		protected.POST("/createLink", paymentController.CreatePaymentLink())
	}

}
