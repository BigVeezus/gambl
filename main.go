package main

import (
	"log"
	"os"

	gameController "gambl/controllers/game"
	userController "gambl/controllers/user"
	"gambl/core/game"
	"gambl/core/user"
	"gambl/database"
	gameRoutes "gambl/routes/game"
	userRoutes "gambl/routes/user"

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

	// Initialize services with the respective repositories
	userService := user.NewUserService(database.OpenCollection(mongoClient, "users"))
	gameService := game.NewGameService()

	// Initialize controllers with the respective services and logger
	userController := userController.NewUserController(*userService, logger)
	gameController := gameController.NewGameController(gameService, logger)

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

	// API-2

	router.Run(":" + port)
}
