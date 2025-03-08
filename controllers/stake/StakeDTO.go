package controllers

import (
    "time"
    "gambl/core/game"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Request DTOs
type PlaceStakeRequest struct {
    GameID          string  `json:"game_id" binding:"required"`
    TeamID          string  `json:"team_id,omitempty"`
    Currency        string  `json:"currency" binding:"required"`
    Amount          float64 `json:"amount" binding:"required,gt=0"`
    PayoutChannelID string  `json:"payout_channel_id" binding:"required"`
}

type ListStakesRequest struct {
    GameID    string    `form:"game_id" binding:"required"`
    Status    []string  `form:"status"`
    FromDate  time.Time `form:"from_date" time_format:"2006-01-02T15:04:05Z07:00"`
    ToDate    time.Time `form:"to_date" time_format:"2006-01-02T15:04:05Z07:00"`
    Limit     int       `form:"limit,default=10"`
    Offset    int       `form:"offset,default=0"`
}

type UpdateStakeStatusRequest struct {
    Status string `json:"status" binding:"required,oneof=PENDING ACTIVE PROCESSING_PAYOUT PAID FAILED REFUNDED"`
}

// Response DTOs
type StakeResponse struct {
    ID               primitive.ObjectID `json:"id"`
    GameID           primitive.ObjectID `json:"game_id"`
    StakerID         string            `json:"staker_id"`
    TeamID           string            `json:"team_id,omitempty"`
    Currency         string            `json:"currency"`
    Amount           float64           `json:"amount"`
    EquivalentAmount float64           `json:"equivalent_amount"`
    PayoutChannelID  primitive.ObjectID `json:"payout_channel_id"`
    Status           game.StakeStatus            `json:"status"`
    CreatedAt        time.Time         `json:"created_at"`
    UpdatedAt        time.Time         `json:"updated_at"`
}

// Conversion methods
func (req *PlaceStakeRequest) ToStakeModel(stakerID string) (*game.GameStake, error) {
    gameID, err := primitive.ObjectIDFromHex(req.GameID)
    if err != nil {
        return nil, err
    }

    payoutChannelID, err := primitive.ObjectIDFromHex(req.PayoutChannelID)
    if err != nil {
        return nil, err
    }
    _stakerID, err := primitive.ObjectIDFromHex(stakerID)
    if err != nil {
        return nil, err
    }

    _teamID, err := primitive.ObjectIDFromHex(req.TeamID)
    if err != nil {
        return nil, err
    }


    return &game.GameStake{
        ID:              primitive.NewObjectID(),
        GameID:          gameID,
        StakerID:        _stakerID,
        TeamID:          _teamID,
        Currency:        req.Currency,
        Amount:          req.Amount,
        PayoutChannelID: payoutChannelID,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }, nil
}

func NewStakeResponse(s *game.GameStake) *StakeResponse {
    
    return &StakeResponse{
        ID:               s.ID,
        GameID:          s.GameID,
        StakerID:        s.StakerID.Hex(),
        TeamID:          s.TeamID.Hex(),
        Currency:        s.Currency,
        Amount:          s.Amount,
        EquivalentAmount: s.EquivalentAmount,
        PayoutChannelID: s.PayoutChannelID,
        Status:          s.Status,
        CreatedAt:       s.CreatedAt,
        UpdatedAt:       s.UpdatedAt,
    }
}

func NewStakeListResponse(stakes []game.GameStake) []*StakeResponse {
    response := make([]*StakeResponse, len(stakes))
    for i, stake := range stakes {
        response[i] = NewStakeResponse(&stake)
    }
    return response
}