// core/game/model.go
package game

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GameStatus string
type StakeStatus string

const (
	StatusCreated         GameStatus  = "CREATED"
	StatusOpen            GameStatus  = "OPEN"
	StatusInPlay          GameStatus  = "IN_PLAY"
	StatusVerifying       GameStatus  = "VERIFYING"
	StatusComplete        GameStatus  = "COMPLETE"
	StatusDisputed        GameStatus  = "DISPUTED"
	StakePending          StakeStatus = "PENDING"
	StakeActive           StakeStatus = "ACTIVE"
	StakeProcessingPayout StakeStatus = "PROCESSING_PAYOUT"
	StakePaid             StakeStatus = "PAID"
	StakeFailed           StakeStatus = "FAILED"
	StakeRefunded         StakeStatus = "REFUNDED"
)

// At the top of model.go with other error definitions

var (
	// Existing errors
	ErrInvalidAmount        = errors.New("stake amount must be greater than 0")
	ErrInvalidPercentages   = errors.New("win and lose percentages must sum to 100")
	ErrInvalidDeadline      = errors.New("deadline must be in the future")
	ErrInvalidTeamSize      = errors.New("team size must be greater than 0")
	ErrMissingCreator       = errors.New("creator ID is required")
	ErrInvalidProofs        = errors.New("verification requirements are invalid")
	ErrInvalidCurrency      = errors.New("unsupported currency")
	ErrInvalidPayoutChannel = errors.New("invalid payout channel")
	ErrNoGameId             = errors.New("game ID is required")
	ErrNoStakerId           = errors.New("staker ID is required")

	// Add service-level errors here
	ErrGameNotFound     = errors.New("game not found")
	ErrStakeNotFound    = errors.New("stake not found")
	ErrInvalidGameState = errors.New("invalid game state")
	ErrUnauthorized     = errors.New("unauthorized action")
	ErrStakeNotAllowed  = errors.New("staking not allowed")
	ErrDuplicateStake   = errors.New("user has already staked")
)

type Game struct {
	ID                        primitive.ObjectID `bson:"_id"`
	Creator_ID                primitive.ObjectID `json:"creator_id" validate:"required"`
	Gambl_Type                string             `json:"gambl_type" validate:"required"` // public_event, custom_event, esports
	Statement                 string             `json:"statement" validate:"required"`
	Description               string             `json:"description"`
	Tags                      []string           `json:"tags"`
	Stakes                    []GameStake        `json:"stakes"`
	Status                    GameStatus         `json:"status" validate:"required"`
	Deadline                  time.Time          `json:"deadline" validate:"required"`
	Team_Size                 int                `json:"team_size,omitempty"` // Optional, for team games
	Is_Valid                  bool               `json:"is_valid,omitempty"`
	ValidationReasoning       string             `json:"validation_reasoning,omitempty"`
	ResolutionSources         string             `json:"resolution_sources,omitempty"`
	Created_At                time.Time          `json:"created_at"`
	Updated_At                time.Time          `json:"updated_at"`
	Verification_Requirements VerificationConfig `json:"verification_requirements"`
	MinimumStake              float64            `json:"minimum_stake" bson:"minimum_stake"`
	CreatorPercentage         float64            `json:"creator_percentage" bson:"creator_percentage"`
	BaseCurrency              string             `json:"base_currency" bson:"base_currency"`
	Teams                     []Team             `json:"teams,omitempty" bson:"teams,omitempty"`
}

type Team struct {
	ID      primitive.ObjectID `json:"id" bson:"id"`
	Name    string             `json:"name" bson:"name"`
	Players []string           `json:"players" bson:"players"`
}

// ValidationResult represents the AI validation response
type ValidationResult struct {
	IsValid                    bool     `json:"is_valid"`
	Confidence                 float64  `json:"confidence"`
	Reasoning                  string   `json:"reasoning"`
	SuggestedResolutionSources []string `json:"suggested_resolution_sources"`
	SuggestedEndDate           string   `json:"suggested_end_date"`
	ClarificationNeeded        bool     `json:"clarification_needed"`
	ClarificationQuestions     []string `json:"clarification_questions"`
}

type ResolutionResult struct {
	IsResolved        bool     `json:"is_resolved"`
	WinningPosition   string   `json:"winning_position"` // support, oppose, inconclusive
	Confidence        float64  `json:"confidence"`
	Reasoning         string   `json:"reasoning"`
	Evidence          []string `json:"evidence"`
	NeedsHumanReview  bool     `json:"needs_human_review"`
	HumanReviewReason string   `json:"human_review_reason"`
}

type GameStake struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GameID           primitive.ObjectID `json:"game_id"`
	StakerID         primitive.ObjectID `json:"staker_id"`
	Currency         string             `json:"currency"`
	PayoutChannel    string             `json:"payout_channel"` // wallet/bank_account
	WinPercent       float64            `json:"win_percent"`
	LosePercent      float64            `json:"lose_percent"`
	CreatedAt        time.Time          `json:"created_at"`
	TeamID           primitive.ObjectID `json:"team_id,omitempty" bson:"team_id,omitempty"` // For team games
	Amount           float64            `json:"amount" bson:"amount"`
	EquivalentAmount float64            `json:"equivalent_amount" bson:"equivalent_amount"` // Amount in game's base currency
	PayoutChannelID  primitive.ObjectID `json:"payout_channel_id" bson:"payout_channel_id"`
	Status           StakeStatus        `json:"status" bson:"status"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

type GameResult struct {
	GameID       string              `json:"game_id"`
	Winners      []Winner            `json:"winners"`
	Verification []VerificationProof `json:"verification"`
	ResultStatus string              `json:"result_status"` // pending, verified, disputed
	SubmittedAt  time.Time           `json:"submitted_at"`
	VerifiedAt   time.Time           `json:"verified_at,omitempty"`
}

type Winner struct {
	PlayerID      string  `json:"player_id"`
	TeamID        string  `json:"team_id,omitempty"`
	StakeID       string  `json:"stake_id"`
	WinningAmount float64 `json:"winning_amount"`
	PayoutStatus  string  `json:"payout_status"`
}

type VerificationConfig struct {
	Required_Proofs     int      `json:"required_proofs"`
	Allowed_Proof_Types []string `json:"allowed_proof_types"` // image, video, link
	Minimum_Verifiers   int      `json:"minimum_verifiers"`
}

type VerificationProof struct {
	Verifier_ID  string    `json:"verifier_id"`
	Proof_Type   string    `json:"proof_type"`
	Proof_URL    string    `json:"proof_url"`
	Submitted_At time.Time `json:"submitted_at"`
}

// type GameWinner struct {
// 	gameId   string
// 	playerId string
// 	teamId   string
// }

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
	if g.Creator_ID.String() == "" {
		return ErrMissingCreator
	}

	if g.Deadline.Before(time.Now()) {
		return ErrInvalidDeadline
	}

	if g.Team_Size < 0 {
		return ErrInvalidTeamSize
	}

	if g.MinimumStake < 0 {
		return errors.New("minimum stake cannot be negative")
	}

	if g.CreatorPercentage < 1 || g.CreatorPercentage > 15 {
		return errors.New("creator percentage must be between 1 and 15")
	}

	if g.BaseCurrency == "" {
		return errors.New("base currency is required")
	}

	// Validate all stakes
	for _, stake := range g.Stakes {
		if err := stake.Validate(); err != nil {
			return err
		}
	}

	if len(g.Teams) > 0 {
		seen := make(map[string]bool)
		for _, team := range g.Teams {
			if team.ID.String() == "" || team.Name == "" {
				return errors.New("team ID and name are required")
			}
			if seen[team.ID.String()] {
				return errors.New("duplicate team ID")
			}
			seen[team.ID.String()] = true
		}
	}

	return g.Verification_Requirements.Validate()
}

func (s *GameStake) Validate() error {
	if s.Amount <= 0 {
		return ErrInvalidAmount
	}

	// if s.WinPercent+s.LosePercent != 100.0 {
	// 	return ErrInvalidPercentages
	// }

	if !ValidCurrencies[s.Currency] {
		return ErrInvalidCurrency
	}

	if !ValidPayoutChannels[s.PayoutChannel] {
		return ErrInvalidPayoutChannel
	}

	if s.PayoutChannelID.String() == "" {
		return ErrInvalidPayoutChannel
	}

	if s.GameID.String() == "" {
		return ErrNoGameId
	}

	if s.StakerID.String() == "" {
		return ErrNoStakerId
	}

	return nil
}

func (v *VerificationConfig) Validate() error {
	// if v.RequiredProofs <= 0 || v.MinimumVerifiers <= 0 {
	//     return ErrInvalidProofs
	// }

	// // Validate proof types
	// for _, proofType := range v.AllowedProofTypes {
	//     if !ValidProofTypes[proofType] {
	//         return errors.New("invalid proof type: " + proofType)
	//     }
	// }

	// if len(v.AllowedProofTypes) == 0 {
	//     return errors.New("at least one proof type must be allowed")
	// }

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
	if v.Verifier_ID == "" {
		return errors.New("verifier ID is required")
	}

	if !ValidProofTypes[v.Proof_Type] {
		return errors.New("invalid proof type")
	}

	if v.Proof_URL == "" {
		return errors.New("proof URL is required")
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
