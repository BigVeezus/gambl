package controllers

import (
	aiCore "gambl/core/ai"
	gameCore "gambl/core/game"
	"time"

	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"

	"net/http"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	gameService gameCore.GameService
	logger      *log.Logger
	aiService   aiCore.AIService
}

func NewGameController(gs gameCore.GameService, l *log.Logger, a aiCore.AIService) *GameController {
	log.Printf("Init: game controller constructor")
	return &GameController{
		gameService: gs,
		logger:      l,
		aiService:   a,
	}
}

func (gc *GameController) CreateGame() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Init: create game controller")

		var req CreateGameRequest
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
				gc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return
			}
			gc.logger.Printf("invalid request, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		creatorID := c.GetString("uid")

		gameModel := req.ToGameModel(creatorID)

		// Validate wager using AI before creating it
		validationResult, err := gc.aiService.ValidateWager(c.Request.Context(), &gameCore.Game{
			Statement:   gameModel.Statement,
			Creator_ID:  gameModel.Creator_ID,
			Deadline:    gameModel.Deadline,
			Description: gameModel.Description,
			Created_At:  time.Now(),
		})

		if err != nil {
			gc.logger.Printf("AI validation failed, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to validate game",
				"message": "Our system couldn't determine if this wager can be resolved. Please try again."})
			return
		}

		// Handle invalid wagers based on AI validation
		if !validationResult.IsValid {
			gc.logger.Printf("Invalid wager detected: %s", validationResult.Reasoning)

			// If AI suggested clarification questions, include them in the response
			var suggestions []string
			if validationResult.ClarificationNeeded && len(validationResult.ClarificationQuestions) > 0 {
				suggestions = validationResult.ClarificationQuestions
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "Invalid wager",
				"reasoning":   validationResult.Reasoning,
				"suggestions": suggestions,
			})
			return
		}

		// If AI suggested a better end date, you might want to use it
		if validationResult.SuggestedEndDate != "" {
			suggestedDate, err := time.Parse("2006-01-02", validationResult.SuggestedEndDate)
			if err == nil && suggestedDate.After(gameModel.Deadline) {
				// Store the AI suggestion to present to the user
				// You could add this to the response or update the model
				gameModel.Deadline = suggestedDate
			}
		}

		// Include validation info in the game model
		gameModel.Is_Valid = validationResult.IsValid
		gameModel.ValidationReasoning = validationResult.Reasoning

		// Store suggested resolution sources if provided
		if len(validationResult.SuggestedResolutionSources) > 0 {
			gameModel.ResolutionSources = strings.Join(validationResult.SuggestedResolutionSources, ", ")
		}

		err = gc.gameService.CreateGame(c.Request.Context(), gameModel)
		if err != nil {
			gc.logger.Printf("failed to create game, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create game", "message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Game Created"})
	}
}

// GetGame handles retrieving a single game by ID
func (gc *GameController) GetGame() gin.HandlerFunc {
	return func(c *gin.Context) {
		gameID := c.Param("game_id")
		if gameID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "game ID is required"})
			return
		}

		game, err := gc.gameService.GetGame(c.Request.Context(), gameID)
		if err != nil {
			if err == gameCore.ErrGameNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
				return
			}
			gc.logger.Printf("failed to get game, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve game"})
			return
		}

		c.JSON(http.StatusOK, game)
	}
}

// ToGameFilters converts the request to GameFilters
func (r *ListGamesRequest) ToGameFilters() gameCore.GameFilters {
	var statuses []gameCore.GameStatus
	for _, s := range r.Status {
		statuses = append(statuses, gameCore.GameStatus(s))
	}

	return gameCore.GameFilters{
		Status:    statuses,
		Type:      r.Type,
		CreatorID: r.CreatorID,
		FromDate:  r.FromDate,
		ToDate:    r.ToDate,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
}

// ListGames handles retrieving multiple games based on filters
func (gc *GameController) ListGames() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListGamesRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			gc.logger.Printf("invalid request parameters, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
			return
		}

		// Convert request to filters
		filters := req.ToGameFilters()

		games, err := gc.gameService.ListGames(c.Request.Context(), filters)
		if err != nil {
			gc.logger.Printf("failed to list games, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve games"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"games": games,
			"meta": gin.H{
				"limit":  filters.Limit,
				"offset": filters.Offset,
				"count":  len(games),
			},
		})
	}
}
