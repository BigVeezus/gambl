package controllers

import (
    "gambl/core/game"
    "time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request DTOs
type CreateGameRequest struct {
	CreatorID	  string	`json:"creatorID" binding:"required"`
    GamblType     string    `json:"gamblType" binding:"required,oneof=public_event custom_event esports"`
    Title         string    `json:"title" binding:"required"`
    Description   string    `json:"description" binding:"required"`
    // Stakes        []StakeRequest `json:"stakes" binding:"required,min=1"`
    Deadline      time.Time `json:"deadline" binding:"required,gt=time.Now"`
    TeamSize      int       `json:"team_size" binding:"omitempty,min=1"`
    // VerificationRequirements VerificationConfigRequest `json:"verification_requirements" binding:"required"`
}

type ListGamesRequest struct {
    Status    []string  `form:"status"`
    Type      string    `form:"type"`
    CreatorID string    `form:"creator_id"`
    FromDate  time.Time `form:"from_date" time_format:"2006-01-02T15:04:05Z07:00"`
    ToDate    time.Time `form:"to_date" time_format:"2006-01-02T15:04:05Z07:00"`
    Limit     int       `form:"limit,default=10"`
    Offset    int       `form:"offset,default=0"`
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
    ID            primitive.ObjectID `bson:"_id"`
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
        Creator_ID:     creatorID,
        Gambl_Type:     req.GamblType,
        Title:         req.Title,
        Description:   req.Description,
        // Stakes:        convertStakeRequests(req.Stakes, ""),  // GameID will be set after creation
        Status:        game.StatusCreated,
        Deadline:      req.Deadline,
        Team_Size:      req.TeamSize,
        Created_At:     time.Now(),
        Updated_At:     time.Now(),
        // Verification_Requirements: game.Verification_Config{
        //     Required_Proofs:    req.Verification_Requirements.Required_Proofs,
        //     Allowed_Proof_Types: req.Verification_Requirements.Allowed_Proof_Types,
        //     Minimum_Verifiers:  req.Verification_Requirements.Minimum_Verifiers,
        // },
    }
}

func NewGameResponse(g *game.Game) *GameResponse {
    return &GameResponse{
        ID:            g.ID,
        CreatorID:     g.Creator_ID,
        GamblType:     g.Gambl_Type,
        Title:         g.Title,
        Description:   g.Description,
        Stakes:        convertToStakeResponses(g.Stakes),
        Status:        g.Status,
        Deadline:      g.Deadline,
        TeamSize:      g.Team_Size,
        CreatedAt:     g.Created_At,
        UpdatedAt:     g.Updated_At,
        VerificationRequirements: VerificationConfigResponse{
            RequiredProofs:    g.Verification_Requirements.Required_Proofs,
            AllowedProofTypes: g.Verification_Requirements.Allowed_Proof_Types,
            MinimumVerifiers:  g.Verification_Requirements.Minimum_Verifiers,
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
