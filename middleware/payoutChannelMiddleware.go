// middleware/payoutChannelMiddleware.go
package middleware

import (
	"gambl/core/payout"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePayoutChannel is middleware that checks if a user has at least one valid payout channel
func RequirePayoutChannel(payoutService payout.PayoutChannelService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		has, err := payoutService.HasValidPayoutChannel(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check payout channels"})
			c.Abort()
			return
		}

		if !has {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "payout channel required",
				"message": "You need to set up a payout channel before staking",
				"code": "PAYOUT_CHANNEL_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}