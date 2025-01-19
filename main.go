package main

import (
	"log"
	"os"

	gameController "gambl/controllers/game"
	paymentController "gambl/controllers/payment"
	userController "gambl/controllers/user"
	"gambl/core/game"
	"gambl/core/payment"
	"gambl/core/user"
	"gambl/database"
	gameRoutes "gambl/routes/game"
	paymentRoutes "gambl/routes/payment"
	userRoutes "gambl/routes/user"

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

	// Initialize services with the respective repositories
	userService := user.NewUserService(database.OpenCollection(mongoClient, "users"))
	gameService := game.NewGameService(database.OpenCollection(mongoClient, "games"))
	paymentService := payment.NewPaymentService(database.OpenCollection(mongoClient, "payments"), paystackProvider,
		flutterwaveProvider)

	// Initialize controllers with the respective services and logger
	userController := userController.NewUserController(*userService, logger)
	gameController := gameController.NewGameController(gameService, logger)
	paymentController := paymentController.NewPaymentController(paymentService, logger)

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
	paymentRoutes.SetupPaymentsRoutes(v1, paymentController)

	// API-2

	router.Run(":" + port)
}
