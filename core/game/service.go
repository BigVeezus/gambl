// core/game/service.go
package game

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

// GameService defines all game-related business operations
type GameService interface {
    // Game Management
    CreateGame(ctx context.Context, game *Game) error
    GetGame(ctx context.Context, id string) (*Game, error)
    ListGames(ctx context.Context, filters GameFilters) ([]Game, error)
    
    // // Stake Operations
    // PlaceStake(ctx context.Context, gameID string, stake *GameStake) error
    // GetStake(ctx context.Context, gameID, stakeID string) (*GameStake, error)
    // GetStakes(ctx context.Context, gameID, filters StakeFilters) ([]GameStake, error)
    
    // // Game Flow
    // StartGame(ctx context.Context, gameID string) error
    // SubmitResult(ctx context.Context, gameID string, result *GameResult) error
    // VerifyResult(ctx context.Context, gameID string, proof *VerificationProof) error
    
    // // Dispute Handling
    // RaiseDispute(ctx context.Context, gameID string, reason string) error
    // ResolveDispute(ctx context.Context, gameID string, resolution string) error
}

// GameFilters for listing games
type GameFilters struct {
    Status    []GameStatus
    Type      string
    CreatorID string
    FromDate  time.Time
    ToDate    time.Time
    Limit     int
    Offset    int
}

type StakeFilters struct {
    StakeID    []string
    MinAmount   float64
    MaxAmount   float64
    Currency    []string
    PayoutChannel   []string
    FromDate  time.Time
    ToDate    time.Time
    Limit     int
    Offset    int
}

// gameService implements GameService
type gameService struct {
    Collection       *mongo.Collection
    // notifier   NotifierService // Interface for notifications
}

// NewGameService creates a new game service
func NewGameService(collection *mongo.Collection) GameService {
    log.Printf("Init: create game service")
    return &gameService{Collection: collection}
}

// Implementation of CreateGame
func (s *gameService) CreateGame(ctx context.Context, game *Game) error {
    log.Printf("Init: create game service")

    // Validate game creation parameters
    err := ValidateGameCreation(game)
    if err != nil {
        return err
    }

    // Set initial game status
    game.Status = StatusCreated
    game.ID = primitive.NewObjectID()
    game.Created_At = time.Now()
    game.Updated_At = time.Now()

//     // Store the game
//     if err := s.repo.Create(ctx, game); err != nil {
//         return err
//     }
gameJSON, _ := json.MarshalIndent(game, "", "  ")
log.Printf("Game Details Before Creation:\n%s", string(gameJSON))
    _, err = s.Collection.InsertOne(ctx, game)

//     // Notify relevant parties
//     s.notifier.NotifyGameCreated(ctx, game)
    
    return err
}

// GetGame retrieves a single game by ID
func (s *gameService) GetGame(ctx context.Context, id string) (*Game, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }

    var game Game
    err = s.Collection.FindOne(ctx, primitive.M{"_id": objectID}).Decode(&game)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrGameNotFound
        }
        return nil, err
    }

    return &game, nil
}

// ListGames retrieves games based on filters
func (s *gameService) ListGames(ctx context.Context, filters GameFilters) ([]Game, error) {
    // Build filter
    filter := primitive.M{}
    
    if len(filters.Status) > 0 {
        filter["status"] = primitive.M{"$in": filters.Status}
    }
    
    if filters.Type != "" {
        filter["gambl_type"] = filters.Type
    }
    
    if filters.CreatorID != "" {
        filter["creator_id"] = filters.CreatorID
    }
    
    if !filters.FromDate.IsZero() {
        filter["created_at"] = primitive.M{"$gte": filters.FromDate}
    }
    
    if !filters.ToDate.IsZero() {
        if _, exists := filter["created_at"]; exists {
            filter["created_at"].(primitive.M)["$lte"] = filters.ToDate
        } else {
            filter["created_at"] = primitive.M{"$lte": filters.ToDate}
        }
    }

    // Set up options for pagination
    opts := options.Find()
    if filters.Limit > 0 {
        opts.SetLimit(int64(filters.Limit))
    }
    if filters.Offset > 0 {
        opts.SetSkip(int64(filters.Offset))
    }

    cursor, err := s.Collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var games []Game
    if err = cursor.All(ctx, &games); err != nil {
        return nil, err
    }

    return games, nil
}


// Implementation of PlaceStake
// func (s *gameService) PlaceStake(ctx context.Context, gameID string, stake *GameStake) error {
//     game, err := s.repo.FindByID(ctx, gameID)
//     if err != nil {
//         return err
//     }

//     // Validate game state
//     if !game.CanAcceptStakes() {
//         return ErrStakeNotAllowed
//     }

//     // Validate stake
//     if err := stake.Validate(); err != nil {
//         return err
//     }


//     // Add stake to game
//     game.Stakes = append(game.Stakes, *stake)
//     if err := s.repo.Update(ctx, game); err != nil {
//         // Unlock funds if stake fails
//         s.walletSvc.UnlockFunds(ctx, stake.StakerID, stake.Amount, stake.Currency)
//         return err
//     }

//     return nil
// }

// Implementation of SubmitResult
// func (s *gameService) SubmitResult(ctx context.Context, gameID string, result *GameResult) error {
//     game, err := s.repo.FindByID(ctx, gameID)
//     if err != nil {
//         return err
//     }

//     if !game.CanSubmitResult() {
//         return ErrInvalidGameState
//     }

//     // Validate result
//     if err := result.Validate(); err != nil {
//         return err
//     }

//     // If verification is required
//     if game.NeedsVerification() {
//         game.Status = StatusVerifying
//         result.ResultStatus = "pending"
//     } else {
//         // Auto-complete for games without verification
//         game.Status = StatusComplete
//         result.ResultStatus = "verified"
        
//         // Process payouts
//         if err := s.processPayout(ctx, game, result); err != nil {
//             return err
//         }
//     }

//     // Store result
//     if err := s.repo.UpdateResult(ctx, gameID, result); err != nil {
//         return err
//     }

//     return nil
// }