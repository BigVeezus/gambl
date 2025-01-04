package controllers

import (
	gameCore "gambl/core/game"

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
}

func NewGameController(gs gameCore.GameService, l *log.Logger) *GameController {
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
        err := gc.gameService.CreateGame(c.Request.Context(), req.ToGameModel(creatorID))
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
