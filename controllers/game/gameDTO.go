package controllers

import (
    "gambl/core/game"
    "time"
)

// Request DTOs
type CreateGameRequest struct {
	CreatorID	  string	`json:"creator_id" binding:"required"`
    GamblType     string    `json:"gambl_type" binding:"required,oneof=public_event custom_event esports"`
    Title         string    `json:"title" binding:"required"`
    Description   string    `json:"description" binding:"required"`
    Stakes        []StakeRequest `json:"stakes" binding:"required,min=1"`
    Deadline      time.Time `json:"deadline" binding:"required,gtfield=time.Now"`
    TeamSize      int       `json:"team_size" binding:"omitempty,min=1"`
    VerificationRequirements VerificationConfigRequest `json:"verification_requirements" binding:"required"`
}

type StakeRequest struct {
    Currency      string  `json:"currency" binding:"required"`
    PayoutChannel string  `json:"payout_channel" binding:"required,oneof=wallet bank_account"`
    Amount        float64 `json:"amount" binding:"required,gt=0"`
    WinPercent    float64 `json:"win_percent" binding:"required"`
    LosePercent   float64 `json:"lose_percent" binding:"required"`
}

type VerificationConfigRequest struct {
    RequiredProofs    int      `json:"required_proofs" binding:"required,min=1"`
    AllowedProofTypes []string `json:"allowed_proof_types" binding:"required,min=1,dive,oneof=image video link"`
    MinimumVerifiers  int      `json:"minimum_verifiers" binding:"required,min=1"`
}

// Response DTOs
type GameResponse struct {
    ID            string              `json:"id"`
    CreatorID     string             `json:"creator_id"`
    GamblType     string             `json:"gambl_type"`
    Title         string             `json:"title"`
    Description   string             `json:"description"`
    Stakes        []StakeResponse    `json:"stakes"`
    Status        game.GameStatus    `json:"status"`
    Deadline      time.Time          `json:"deadline"`
    TeamSize      int                `json:"team_size,omitempty"`
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`
    VerificationRequirements VerificationConfigResponse `json:"verification_requirements"`
}

type StakeResponse struct {
    ID            string    `json:"id"`
    GameID        string    `json:"game_id"`
    StakerID      string    `json:"staker_id"`
    Currency      string    `json:"currency"`
    PayoutChannel string    `json:"payout_channel"`
    Amount        float64   `json:"amount"`
    WinPercent    float64   `json:"win_percent"`
    LosePercent   float64   `json:"lose_percent"`
    CreatedAt     time.Time `json:"created_at"`
    Status        string    `json:"status"`
}

type VerificationConfigResponse struct {
    RequiredProofs    int      `json:"required_proofs"`
    AllowedProofTypes []string `json:"allowed_proof_types"`
    MinimumVerifiers  int      `json:"minimum_verifiers"`
}

// Conversion methods
func (req *CreateGameRequest) ToGameModel(creatorID string) *game.Game {
    return &game.Game{
        CreatorID:     creatorID,
        GamblType:     req.GamblType,
        Title:         req.Title,
        Description:   req.Description,
        Stakes:        convertStakeRequests(req.Stakes, ""),  // GameID will be set after creation
        Status:        game.StatusCreated,
        Deadline:      req.Deadline,
        TeamSize:      req.TeamSize,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
        VerificationRequirements: game.VerificationConfig{
            RequiredProofs:    req.VerificationRequirements.RequiredProofs,
            AllowedProofTypes: req.VerificationRequirements.AllowedProofTypes,
            MinimumVerifiers:  req.VerificationRequirements.MinimumVerifiers,
        },
    }
}

func NewGameResponse(g *game.Game) *GameResponse {
    return &GameResponse{
        ID:            g.ID,
        CreatorID:     g.CreatorID,
        GamblType:     g.GamblType,
        Title:         g.Title,
        Description:   g.Description,
        Stakes:        convertToStakeResponses(g.Stakes),
        Status:        g.Status,
        Deadline:      g.Deadline,
        TeamSize:      g.TeamSize,
        CreatedAt:     g.CreatedAt,
        UpdatedAt:     g.UpdatedAt,
        VerificationRequirements: VerificationConfigResponse{
            RequiredProofs:    g.VerificationRequirements.RequiredProofs,
            AllowedProofTypes: g.VerificationRequirements.AllowedProofTypes,
            MinimumVerifiers:  g.VerificationRequirements.MinimumVerifiers,
        },
    }
}

// Helper functions
func convertStakeRequests(stakes []StakeRequest, gameID string) []game.GameStake {
    result := make([]game.GameStake, len(stakes))
    for i, stake := range stakes {
        result[i] = game.GameStake{
            GameID:        gameID,
            Currency:      stake.Currency,
            PayoutChannel: stake.PayoutChannel,
            Amount:        stake.Amount,
            WinPercent:    stake.WinPercent,
            LosePercent:   stake.LosePercent,
            CreatedAt:     time.Now(),
            Status:        "active",
        }
    }
    return result
}

func convertToStakeResponses(stakes []game.GameStake) []StakeResponse {
    result := make([]StakeResponse, len(stakes))
    for i, stake := range stakes {
        result[i] = StakeResponse{
            ID:            stake.ID,
            GameID:        stake.GameID,
            StakerID:      stake.StakerID,
            Currency:      stake.Currency,
            PayoutChannel: stake.PayoutChannel,
            Amount:        stake.Amount,
            WinPercent:    stake.WinPercent,
            LosePercent:   stake.LosePercent,
            CreatedAt:     stake.CreatedAt,
            Status:        stake.Status,
        }
    }
    return result
}