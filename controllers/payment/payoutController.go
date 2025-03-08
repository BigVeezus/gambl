package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/go-playground/validator/v10"

	"go.mongodb.org/mongo-driver/bson/primitive"

	paymentCore "gambl/core/payment"
	"gambl/core/payout"

	providers "gambl/providers/payment"
)

type PayoutController struct {
	paymentService paymentCore.PaymentService
	logger         *log.Logger
}

func NewPaymentController(p paymentCore.PaymentService, l *log.Logger) *PayoutController {
	log.Printf("Init: payment controller constructor")
	return &PayoutController{
		paymentService: p,
		logger:         l,
	}
}

func (pc *PayoutController) CreatePaymentLink() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req paymentCore.CreatePaymentLink

		if err := c.BindJSON(&req); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				var errorMessages []string
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field: %s, Error: %s, Value: %v",
						e.Field(),
						e.Tag(),
						e.Value(),
					))
				}
				pc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return
			}
			pc.logger.Printf("invalid request, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		userId := c.GetString("uid")
		req.UserID = userId

		payment := &paymentCore.Payment{
			UserID: func() primitive.ObjectID {
				id, err := primitive.ObjectIDFromHex(req.UserID)
				if err != nil {
					pc.logger.Printf("invalid user ID: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
					return primitive.NilObjectID
				}
				return id
			}(),
			GameID: func() primitive.ObjectID {
				id, err := primitive.ObjectIDFromHex(req.GameID)
				if err != nil {
					pc.logger.Printf("invalid game ID: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
					return primitive.NilObjectID
				}
				return id
			}(),
			Currency: req.Currency,
		}

		paymentModel, err := payment.ToPaymentModel()
		if err != nil {
			pc.logger.Printf("failed to convert payment to model: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
			return
		}

		createdPayment, err := pc.paymentService.CreatePayment(c.Request.Context(), paymentModel)
		if err != nil {
			pc.logger.Printf("failed to create payment: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		req.PaymentID = createdPayment.ID.Hex()

		provider, err := pc.paymentService.CreatePaymentLink(&req, "flutterwave")
		if provider == nil {
			pc.logger.Printf(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create payment link", "message": "provider is empty"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success":  true,
			"provider": provider,
		})
	}
}

func (pc *PayoutController) FlutterwaveWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {

		var reqBody paymentCore.FlutterwaveWebhookData

		// bodyBytes, err := io.ReadAll(c.Request.Body)
		// if err != nil {
		// 	pc.logger.Printf("failed to read request body: %v", err)
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		// 	return
		// }

		// Print the raw JSON request body
		// fmt.Printf("Raw Request Body: %s\n", string(bodyBytes))

		// Restore the request body to its original state
		// c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := c.BindJSON(&reqBody); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				var errorMessages []string
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field: %s, Error: %s",
						e.Field(),
						e.Tag(),
					))
				}
				pc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return
			}
			pc.logger.Printf("invalid request, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		url := fmt.Sprintf("https://api.flutterwave.com/v3/transactions/%d/verify", reqBody.ID)

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", "Bearer FLWSECK_TEST-61625cd73f0ff864a353c3449fadef32-X")
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			pc.logger.Printf("HTTP request failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to make HTTP request"})
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			pc.logger.Printf("Failed to read response body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response body"})
			return
		}

		// fmt.Println(string(body))

		var webhookResponse providers.FlutterwaveWebhookResponse
		if err := json.Unmarshal(body, &webhookResponse); err != nil {
			pc.logger.Printf("Failed to unmarshal response body: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
			return
		}

		// Access individual fields as needed
		if webhookResponse.Status == "success" && webhookResponse.Data.Status == "successful" {
			// Process the payment success
			pc.logger.Printf("Transaction ID: %d, Amount: %.2f", webhookResponse.Data.ID, webhookResponse.Data.Amount)
			// Example: Check the game ID
			gameID := webhookResponse.Data.Meta.GameID
			userID := webhookResponse.Data.Meta.UserID
			paymentID := webhookResponse.Data.Meta.UserID

			fmt.Printf("Game ID: %s\n", gameID)
			fmt.Printf("User ID: %s\n", userID)
			fmt.Printf("Payment ID: %s\n", paymentID)

			// gameObjectID, err := primitive.ObjectIDFromHex(gameID)
			// if err != nil {
			// 	pc.logger.Printf("Invalid gameID: %v", err)
			// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gameID format"})
			// 	return
			// }

			// userObjectID, err := primitive.ObjectIDFromHex(webhookResponse.Data.Meta.GameID)
			// if err != nil {
			// 	pc.logger.Printf("Invalid userID: %v", err)
			// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
			// 	return
			// }

			paymentObjId, err := primitive.ObjectIDFromHex(webhookResponse.Data.Meta.PaymentID)
			if err != nil {
				pc.logger.Printf("Invalid userID: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
				return
			}

			// change payment status
			filter := bson.M{
				"_id": paymentObjId,
			}

			updates := bson.M{
				"status":     paymentCore.StatusSuccessful,
				"updated_at": time.Now(),
			}

			if err := pc.paymentService.UpdatePayment(c.Request.Context(), filter, updates); err != nil {
				pc.logger.Printf("failed to UPDATE payment: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			println("SUCCESSSS")

		} else {
			pc.logger.Printf("Transaction failed: %s", webhookResponse.Data.ProcessorResponse)
			c.JSON(http.StatusBadRequest, gin.H{"error": "transaction failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"body":    bson.M{},
		})
	}
}

type PayoutChannelController struct {
	payoutService payout.PayoutChannelService
	logger        *log.Logger
}

func NewPayoutChannelController(ps payout.PayoutChannelService, l *log.Logger) *PayoutChannelController {
	return &PayoutChannelController{
		payoutService: ps,
		logger:        l,
	}
}

// CreatePayoutChannel handles the creation of a new payout channel
func (pc *PayoutChannelController) CreatePayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req CreatePayoutChannelRequest
		if err := c.BindJSON(&req); err != nil {
			pc.logger.Printf("invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Convert request to model
		payoutChannel := req.ToPayoutChannelModel(userID)

		// If it's a bank account, we need to verify it
		if payoutChannel.ChannelType == payout.ChannelBank {
			accountName, err := pc.payoutService.VerifyBankAccount(c.Request.Context(), payoutChannel.BankID, payoutChannel.AccountNo)
			if err != nil {
				pc.logger.Printf("bank verification failed: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "bank account verification failed", "details": err.Error()})
				return
			}
			payoutChannel.AccountName = accountName
		}

		// Create the payout channel
		err := pc.payoutService.CreatePayoutChannel(c.Request.Context(), payoutChannel)
		if err != nil {
			if err == payout.ErrPayoutChannelExists {
				c.JSON(http.StatusConflict, gin.H{"error": "a payout channel already exists with this currency and type"})
				return
			}
			pc.logger.Printf("failed to create payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payout channel"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "payout channel created successfully",
			"data": PayoutChannelResponse{
				ID:          payoutChannel.ID.Hex(),
				ChannelType: string(payoutChannel.ChannelType),
				Currency:    payoutChannel.Currency,
				IsDefault:   payoutChannel.IsDefault,
				CreatedAt:   payoutChannel.CreatedAt,
			},
		})
	}
}

// GetPayoutChannels handles retrieving all payout channels for a user
func (pc *PayoutChannelController) GetPayoutChannels() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		channels, err := pc.payoutService.GetUserPayoutChannels(c.Request.Context(), userID)
		if err != nil {
			pc.logger.Printf("failed to get payout channels: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payout channels"})
			return
		}

		// Convert to response format
		var response []PayoutChannelResponse
		for _, channel := range channels {
			response = append(response, PayoutChannelToResponse(&channel))
		}

		c.JSON(http.StatusOK, gin.H{
			"data": response,
		})
	}
}

// GetPayoutChannel handles retrieving a single payout channel
func (pc *PayoutChannelController) GetPayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		channelID := c.Param("id")
		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
			return
		}

		channel, err := pc.payoutService.GetPayoutChannel(c.Request.Context(), channelID)
		if err != nil {
			if err == payout.ErrPayoutChannelNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "payout channel not found"})
				return
			}
			pc.logger.Printf("failed to get payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payout channel"})
			return
		}

		// Ensure the channel belongs to the authenticated user
		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": PayoutChannelToResponse(channel),
		})
	}
}

// UpdatePayoutChannel handles updating a payout channel
func (pc *PayoutChannelController) UpdatePayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		channelID := c.Param("id")
		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
			return
		}

		// Get the existing channel first
		channel, err := pc.payoutService.GetPayoutChannel(c.Request.Context(), channelID)
		if err != nil {
			if err == payout.ErrPayoutChannelNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "payout channel not found"})
				return
			}
			pc.logger.Printf("failed to get payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payout channel"})
			return
		}

		// Ensure the channel belongs to the authenticated user
		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		var req UpdatePayoutChannelRequest
		if err := c.BindJSON(&req); err != nil {
			pc.logger.Printf("invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Build updates map
		updates := req.ToUpdateMap()

		// Apply updates
		err = pc.payoutService.UpdatePayoutChannel(c.Request.Context(), channelID, updates)
		if err != nil {
			pc.logger.Printf("failed to update payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payout channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "payout channel updated successfully",
		})
	}
}

// DeletePayoutChannel handles deleting a payout channel
func (pc *PayoutChannelController) DeletePayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		channelID := c.Param("id")
		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
			return
		}

		// Get the existing channel first
		channel, err := pc.payoutService.GetPayoutChannel(c.Request.Context(), channelID)
		if err != nil {
			if err == payout.ErrPayoutChannelNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "payout channel not found"})
				return
			}
			pc.logger.Printf("failed to get payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payout channel"})
			return
		}

		// Ensure the channel belongs to the authenticated user
		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Delete the channel
		err = pc.payoutService.DeletePayoutChannel(c.Request.Context(), channelID)
		if err != nil {
			pc.logger.Printf("failed to delete payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete payout channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "payout channel deleted successfully",
		})
	}
}

// SetDefaultPayoutChannel handles setting a payout channel as default
func (pc *PayoutChannelController) SetDefaultPayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		channelID := c.Param("id")
		if channelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "channel ID is required"})
			return
		}

		// Get the existing channel first
		channel, err := pc.payoutService.GetPayoutChannel(c.Request.Context(), channelID)
		if err != nil {
			if err == payout.ErrPayoutChannelNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "payout channel not found"})
				return
			}
			pc.logger.Printf("failed to get payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payout channel"})
			return
		}

		// Ensure the channel belongs to the authenticated user
		if channel.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Set as default
		err = pc.payoutService.SetDefaultPayoutChannel(c.Request.Context(), userID, channelID)
		if err != nil {
			pc.logger.Printf("failed to set default payout channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set default payout channel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "payout channel set as default successfully",
		})
	}
}

// GetBanks handles retrieving the list of supported banks
func (pc *PayoutChannelController) GetBanks() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder for now - implement based on your bank provider
		c.JSON(http.StatusOK, gin.H{
			"message": "This endpoint will return the list of supported banks",
		})
	}
}

// VerifyBankAccount handles verifying a bank account
func (pc *PayoutChannelController) VerifyBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyBankAccountRequest
		if err := c.BindJSON(&req); err != nil {
			pc.logger.Printf("invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		accountName, err := pc.payoutService.VerifyBankAccount(c.Request.Context(), req.BankID, req.AccountNumber)
		if err != nil {
			pc.logger.Printf("bank verification failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "bank account verification failed", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"account_name": accountName,
				"bank_id":      req.BankID,
				"account_no":   req.AccountNumber,
			},
		})
	}
}

// HasPayoutChannel checks if a user has at least one payout channel
func (pc *PayoutChannelController) HasPayoutChannel() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("uid")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		hasChannel, err := pc.payoutService.HasValidPayoutChannel(c.Request.Context(), userID)
		if err != nil {
			pc.logger.Printf("failed to check payout channels: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check payout channels"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"has_payout_channel": hasChannel,
		})
	}
}
