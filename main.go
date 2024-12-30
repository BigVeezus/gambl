package main

import (
	"os"
	"log"

	userRoutes "gambl/routes/user"
	gameRoutes "gambl/routes/game"
	gameController "gambl/controllers/game"
	gameService "gambl/core/game"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_"github.com/heroku/x/hmetrics/onload"
)


func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	logger := log.New(os.Stdout, "[GAMBL]", log.LstdFlags) 
	gameService := gameService.NewGameService()
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
	userRoutes.AuthRoutes(v1)

	// Protected routes under version 1
	userRoutes.UserRoutes(v1)
	gameRoutes.SetupGameRoutes(v1, gameController)

	// API-2

	router.Run(":" + port)
}
