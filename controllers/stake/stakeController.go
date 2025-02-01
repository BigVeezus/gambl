package controllers

import (
    "log"
    "net/http"
    gameCore "gambl/core/game"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type StakeController struct {
    stakeService gameCore.StakeService
    logger       *log.Logger
}

func NewStakeController(ss gameCore.StakeService, l *log.Logger) *StakeController {
    return &StakeController{
        stakeService: ss,
        logger:      l,
    }
}

// PlaceStake handles the creation of a new stake
func (sc *StakeController) PlaceStake() gin.HandlerFunc {
    return func(c *gin.Context) {
        var req PlaceStakeRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            if validationErrors, ok := err.(validator.ValidationErrors); ok {
                var errorMessages []string
                for _, e := range validationErrors {
                    errorMessages = append(errorMessages, e.Error())
                }
                c.JSON(http.StatusBadRequest, gin.H{
                    "error":   "Validation failed",
                    "details": errorMessages,
                })
                return
            }
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        // Get user ID from context (set by auth middleware)
        userID := c.GetString("uid")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        // Convert request to model
        stake, err := req.ToStakeModel(userID)
        if err != nil {
            sc.logger.Printf("Failed to convert request to model: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game or payout channel ID"})
            return
        }

        // Place stake
        if err := sc.stakeService.PlaceStake(c.Request.Context(), stake); err != nil {
            sc.logger.Printf("Failed to place stake: %v", err)
            switch err {
            case gameCore.ErrGameNotOpen:
                c.JSON(http.StatusBadRequest, gin.H{"error": "Game is not open for stakes"})
            case gameCore.ErrInvalidPayoutChannel:
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payout channel"})
            case gameCore.ErrBelowMinimumStake:
                c.JSON(http.StatusBadRequest, gin.H{"error": "Stake amount below minimum"})
            case gameCore.ErrInvalidTeam:
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team selection"})
            default:
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place stake"})
            }
            return
        }

        c.JSON(http.StatusCreated, NewStakeResponse(stake))
    }
}

// GetStake handles retrieving a single stake by ID
func (sc *StakeController) GetStake() gin.HandlerFunc {
    return func(c *gin.Context) {
        stakeID, err := primitive.ObjectIDFromHex(c.Param("stake_id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stake ID"})
            return
        }

        stake, err := sc.stakeService.GetStake(c.Request.Context(), stakeID)
        if err != nil {
            sc.logger.Printf("Failed to get stake: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "Stake not found"})
            return
        }

        // Verify user has permission to view this stake
        userID := c.GetString("uid")
        if stake.StakerID != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this stake"})
            return
        }

        c.JSON(http.StatusOK, NewStakeResponse(stake))
    }
}

// ListStakes handles retrieving stakes for a game
func (sc *StakeController) ListStakes() gin.HandlerFunc {
    return func(c *gin.Context) {
        var req ListStakesRequest
        if err := c.ShouldBindQuery(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
            return
        }

        gameID, err := primitive.ObjectIDFromHex(req.GameID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
            return
        }

        stakes, err := sc.stakeService.ListStakes(c.Request.Context(), gameID)
        if err != nil {
            sc.logger.Printf("Failed to list stakes: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stakes"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "stakes": NewStakeListResponse(stakes),
            "meta": gin.H{
                "limit":  req.Limit,
                "offset": req.Offset,
                "count":  len(stakes),
            },
        })
    }
}

// UpdateStakeStatus handles updating a stake's status
func (sc *StakeController) UpdateStakeStatus() gin.HandlerFunc {
    return func(c *gin.Context) {
        stakeID, err := primitive.ObjectIDFromHex(c.Param("stake_id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stake ID"})
            return
        }

        var req UpdateStakeStatusRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        if err := sc.stakeService.UpdateStakeStatus(c.Request.Context(), stakeID, req.Status); err != nil {
            sc.logger.Printf("Failed to update stake status: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stake status"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Stake status updated"})
    }
}