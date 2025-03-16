package main

import (
	gameController "gambl/controllers/game"
	controllers "gambl/controllers/payment"
	stakeController "gambl/controllers/stake"
	userController "gambl/controllers/user"
	"gambl/core/ai"
	"gambl/core/game"
	"gambl/core/payment"
	"gambl/core/payout" // New import
	"gambl/core/user"
	"gambl/database"
	gameRoutes "gambl/routes/game"
	paymentRoutes "gambl/routes/payment"
	stakeRoutes "gambl/routes/stake"
	userRoutes "gambl/routes/user"
	"log"
	"os"

	providers "gambl/providers/payment"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	// initialize logger
	logger := log.New(os.Stdout, "[GAMBL]", log.LstdFlags)

	mongoClient := database.Client

	// Initialize providers with keys from environment variables
	paystackProvider := providers.NewPaystackProvider(
		os.Getenv("PAYSTACK_SECRET_KEY"),
		os.Getenv("PAYSTACK_PUBLIC_KEY"),
	)

	flutterwaveProvider := providers.NewFlutterwaveProvider(
		os.Getenv("FLW_SECRET_KEY"),
		os.Getenv("FLW_PUB_KEY"),
	)

	// Initialize bank verification provider
	bankVerificationProvider := providers.NewPaystackProvider(
		os.Getenv("PAYSTACK_SECRET_KEY"),
		os.Getenv("PAYSTACK_PUBLIC_KEY"),
	)

	// Map bank verification providers
	bankProviders := map[string]payout.BankVerificationProvider{
		"paystack": bankVerificationProvider,
	}

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize AI controller
	aiService := ai.NewAIService(apiKey)

	// Initialize services with the respective repositories
	userService := user.NewUserService(database.OpenCollection(mongoClient, "users"))
	gameService := game.NewGameService(database.OpenCollection(mongoClient, "games"))
	payoutChannelService := payout.NewPayoutChannelService(
		database.OpenCollection(mongoClient, "payout_channels"),
		bankProviders,
		"paystack",
	)
	stakeService := game.NewStakeService(
		database.OpenCollection(mongoClient, "games"),
		database.OpenCollection(mongoClient, "stakes"),
		database.OpenCollection(mongoClient, "payout_channels"),
	)
	paymentService := payment.NewPaymentService(
		database.OpenCollection(mongoClient, "payments"),
		paystackProvider,
		flutterwaveProvider,
	)

	// Initialize controllers with the respective services and logger
	userController := userController.NewUserController(*userService, logger)
	gameController := gameController.NewGameController(gameService, logger, aiService)
	stakeController := stakeController.NewStakeController(stakeService, logger)
	paymentController := controllers.NewPaymentController(paymentService, logger)
	payoutChannelController := controllers.NewPayoutChannelController(payoutChannelService, logger)

	router := gin.New()

	router.Use(gin.Logger())
	ginpprof.Wrap(router)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000/*", "http://localhost:3000", "http://localhost:3000/"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Accept-Language", "Content-Length", "Accept-Language", "Accept-Encoding", "X-CSRF-Token", "accept", "origin", "Cache-Control", "authorizationrequired", "Authorizationrequired", "authorization", "Connection", "Access-Control-Allow-Origin", "Authorization"},
		AllowWildcard:    true,
		AllowCredentials: true,
	}))

	// Initialize version group
	v1 := router.Group("/v1")

	// Unprotected routes under version 1
	userRoutes.SetupAuthRoutes(v1, userController)

	// Protected routes under version 1
	userRoutes.SetupUserRoutes(v1, userController)
	gameRoutes.SetupGameRoutes(v1, gameController)
	stakeRoutes.SetupStakeRoutes(v1, stakeController)
	paymentRoutes.SetupPaymentsRoutes(v1, paymentController)
	paymentRoutes.SetupPayoutChannelRoutes(v1, payoutChannelController) // Add new routes for payout channels

	// API-2

	router.Run(":" + port)
}
