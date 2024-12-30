// core/game/model.go
package game

import (
    "errors"
    "time"
)

type GameStatus string

const (
    StatusCreated   GameStatus = "CREATED"
    StatusOpen      GameStatus = "OPEN"
    StatusInPlay    GameStatus = "IN_PLAY"
    StatusVerifying GameStatus = "VERIFYING"
    StatusComplete  GameStatus = "COMPLETE"
    StatusDisputed  GameStatus = "DISPUTED"
)

type Game struct {
    ID            string      `json:"id" bson:"_id,omitempty"`
    CreatorID     string      `json:"creator_id"`
    GamblType     string      `json:"gambl_type"` // public_event, custom_event, esports
    Title         string      `json:"title"`
    Description   string      `json:"description"`
    Stakes        []GameStake `json:"stakes"`
    Status        GameStatus  `json:"status"`
    Deadline      time.Time   `json:"deadline"`
    TeamSize      int         `json:"team_size,omitempty"` // Optional, for team games
    CreatedAt     time.Time   `json:"created_at"`
    UpdatedAt     time.Time   `json:"updated_at"`
    VerificationRequirements VerificationConfig `json:"verification_requirements"`
}

type GameStake struct {
    ID            string    `json:"id" bson:"_id,omitempty"`
    GameID        string    `json:"game_id"`
    StakerID      string    `json:"staker_id"`
    Currency      string    `json:"currency"`
    PayoutChannel string    `json:"payout_channel"` // wallet/bank_account
    Amount        float64   `json:"amount"`
    WinPercent    float64   `json:"win_percent"`
    LosePercent   float64   `json:"lose_percent"`
    CreatedAt     time.Time `json:"created_at"`
    Status        string    `json:"status"` // active, paid, refunded
}

type GameResult struct {
    GameID        string    `json:"game_id"`
    Winners       []Winner  `json:"winners"`
    Verification  []VerificationProof `json:"verification"`
    ResultStatus  string    `json:"result_status"` // pending, verified, disputed
    SubmittedAt   time.Time `json:"submitted_at"`
    VerifiedAt    time.Time `json:"verified_at,omitempty"`
}

type Winner struct {
    PlayerID      string  `json:"player_id"`
    TeamID        string  `json:"team_id,omitempty"`
    StakeID       string  `json:"stake_id"`
    WinningAmount float64 `json:"winning_amount"`
    PayoutStatus  string  `json:"payout_status"`
}

type VerificationConfig struct {
    RequiredProofs    int      `json:"required_proofs"`
    AllowedProofTypes []string `json:"allowed_proof_types"` // image, video, link
    MinimumVerifiers  int      `json:"minimum_verifiers"`
}

type VerificationProof struct {
    VerifierID   string    `json:"verifier_id"`
    ProofType    string    `json:"proof_type"`
    ProofURL     string    `json:"proof_url"`
    SubmittedAt  time.Time `json:"submitted_at"`
}

type GameWinner struct {
	gameId string
	playerId string
	teamId string
}

var (
    ErrInvalidAmount      = errors.New("stake amount must be greater than 0")
    ErrInvalidPercentages = errors.New("win and lose percentages must sum to 100")
    ErrInvalidDeadline    = errors.New("deadline must be in the future")
    ErrInvalidTeamSize    = errors.New("team size must be greater than 0")
    ErrMissingCreator     = errors.New("creator ID is required")
    ErrInvalidProofs      = errors.New("verification requirements are invalid")
    ErrInvalidCurrency    = errors.New("unsupported currency")
    ErrInvalidPayoutChannel = errors.New("invalid payout channel")
)

// Supported values
var (
    ValidCurrencies = map[string]bool{
        "USD": true,
        "EUR": true,
        "BTC": true,
        "ETH": true,
    }
    
    ValidPayoutChannels = map[string]bool{
        "wallet":       true,
        "bank_account": true,
    }
    
    ValidProofTypes = map[string]bool{
        "image": true,
        "video": true,
        "link":  true,
    }
)

func (g *Game) Validate() error {
    if g.CreatorID == "" {
        return ErrMissingCreator
    }

    if g.Deadline.Before(time.Now()) {
        return ErrInvalidDeadline
    }

    if g.TeamSize < 0 {
        return ErrInvalidTeamSize
    }

    // Validate all stakes
    for _, stake := range g.Stakes {
        if err := stake.Validate(); err != nil {
            return err
        }
    }

    return g.VerificationRequirements.Validate()
}

func (s *GameStake) Validate() error {
    if s.Amount <= 0 {
        return ErrInvalidAmount
    }

    if s.WinPercent+s.LosePercent != 100.0 {
        return ErrInvalidPercentages
    }

    if !ValidCurrencies[s.Currency] {
        return ErrInvalidCurrency
    }

    if !ValidPayoutChannels[s.PayoutChannel] {
        return ErrInvalidPayoutChannel
    }

    return nil
}

func (v *VerificationConfig) Validate() error {
    if v.RequiredProofs <= 0 || v.MinimumVerifiers <= 0 {
        return ErrInvalidProofs
    }

    // Validate proof types
    for _, proofType := range v.AllowedProofTypes {
        if !ValidProofTypes[proofType] {
            return errors.New("invalid proof type: " + proofType)
        }
    }

    if len(v.AllowedProofTypes) == 0 {
        return errors.New("at least one proof type must be allowed")
    }

    return nil
}

func (r *GameResult) Validate() error {
    if len(r.Winners) == 0 {
        return errors.New("at least one winner must be specified")
    }

    // Validate each winner
    for _, winner := range r.Winners {
        if err := winner.Validate(); err != nil {
            return err
        }
    }

    // Validate verification proofs
    if len(r.Verification) == 0 {
        return errors.New("at least one verification proof is required")
    }

    for _, proof := range r.Verification {
        if err := proof.Validate(); err != nil {
            return err
        }
    }

    return nil
}

func (w *Winner) Validate() error {
    if w.PlayerID == "" {
        return errors.New("player ID is required")
    }

    if w.WinningAmount <= 0 {
        return errors.New("winning amount must be greater than 0")
    }

    return nil
}

func (v *VerificationProof) Validate() error {
    if v.VerifierID == "" {
        return errors.New("verifier ID is required")
    }

    if !ValidProofTypes[v.ProofType] {
        return errors.New("invalid proof type")
    }

    if v.ProofURL == "" {
        return errors.New("proof URL is required")
    }

    return nil
}

// Helper method to validate game creation parameters
func ValidateGameCreation(g *Game) error {
    if err := g.Validate(); err != nil {
        return err
    }

    // Additional game creation specific validations
    if len(g.Stakes) == 0 {
        return errors.New("game must have at least one stake")
    }

    // Validate game type
    switch g.GamblType {
    case "public_event", "custom_event", "esports":
        // valid
    default:
        return errors.New("invalid game type")
    }

    return nil
}

// Helper method to check if a game can accept new stakes
func (g *Game) CanAcceptStakes() bool {
    return g.Status == StatusOpen && time.Now().Before(g.Deadline)
}

// Helper method to check if a game is ready for result submission
func (g *Game) CanSubmitResult() bool {
    return g.Status == StatusInPlay && time.Now().After(g.Deadline)
}