package main

import (
	"os"

	routes "gambl/routes"

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

	routes.AuthRoutes(v1)
	routes.UserRoutes(v1)
	routes.RepRoutes(v1)

	// API-2

	router.Run(":" + port)
}
