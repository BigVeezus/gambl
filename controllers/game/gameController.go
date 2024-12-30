package controllers

import (
	"gambl/core/game"

	"github.com/go-playground/validator/v10"
	"log"
    "fmt"
    "strings"

	"github.com/gin-gonic/gin"
	"net/http"
)

type GameController struct {
	gameService game.GameService
	logger      *log.Logger
}

func NewGameController(gs game.GameService, l *log.Logger) *GameController {
	log.Printf("Init: game controller constructor")
	return &GameController{
		gameService: gs,
		logger:      l,
	}
}

func (gc *GameController) CreateGame() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Init: create game controller")

		var req CreateGameRequest
		if err := c.BindJSON(&req); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				// Format validation errors
				var errorMessages []string
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field: %s, Error: %s, Value: %v",
						e.Field(),
						e.Tag(),
						e.Value(),
					))
				}

				gc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return

				// gc.logger.Printf("invalid request, error: %e", err)
				// c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid request"})
				// return
			}

			creatorID := c.GetString("uid")

			err := gc.gameService.CreateGame(c.Request.Context(), req.ToGameModel(creatorID))
			if err != nil {
				gc.logger.Printf("failed to create game, error: %e", err)
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to create game"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"Message": "Game Created"})
		}
	}
}
