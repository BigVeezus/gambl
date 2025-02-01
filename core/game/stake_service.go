// core/game/stake_service.go
package game

import (
    "context"
    "errors"
    "time"
    "gambl/core/payout"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
    ErrBelowMinimumStake = errors.New("stake amount below minimum")
    ErrGameNotOpen = errors.New("game not open for stakes")
    ErrInvalidTeam = errors.New("invalid team selection")
)

type StakeService interface {
    PlaceStake(ctx context.Context, stake *GameStake) error
    GetStake(ctx context.Context, stakeID primitive.ObjectID) (*GameStake, error)
    ListStakes(ctx context.Context, gameID primitive.ObjectID) ([]GameStake, error)
    UpdateStakeStatus(ctx context.Context, stakeID primitive.ObjectID, status string) error
}

type stakeService struct {
    gameCollection *mongo.Collection
    stakeCollection *mongo.Collection
    payoutChannelCollection *mongo.Collection
}

func NewStakeService(gameCol, stakeCol, payoutCol *mongo.Collection) StakeService {
    return &stakeService{
        gameCollection: gameCol,
        stakeCollection: stakeCol,
        payoutChannelCollection: payoutCol,
    }
}

func (s *stakeService) PlaceStake(ctx context.Context, stake *GameStake) error {
    // 1. Validate the stake
    if err := stake.Validate(); err != nil {
        return err
    }

    // 2. Get and validate game
    var game Game
    err := s.gameCollection.FindOne(ctx, bson.M{"_id": stake.GameID}).Decode(&game)
    if err != nil {
        return err
    }

    if game.Status != StatusOpen {
        return ErrGameNotOpen
    }

    // 3. Validate team if specified
    if stake.TeamID.String() != "" {
        validTeam := false
        for _, team := range game.Teams {
            if team.ID == stake.TeamID {
                validTeam = true
                break
            }
        }
        if !validTeam {
            return ErrInvalidTeam
        }
    }

    // 4. Verify payout channel exists and is active
    var channel payout.PayoutChannel
    err = s.payoutChannelCollection.FindOne(ctx, bson.M{
        "_id": stake.PayoutChannelID,
        "user_id": stake.StakerID,
        "is_active": true,
    }).Decode(&channel)
    if err != nil {
        return ErrInvalidPayoutChannel
    }

    // 5. Convert stake amount to base currency if needed
    // TODO: Implement currency conversion
    stake.EquivalentAmount = stake.Amount // Temporary direct assignment
    
    // 6. Check minimum stake
    if stake.EquivalentAmount < game.MinimumStake {
        return ErrBelowMinimumStake
    }

    // 7. Set initial stake status
    stake.Status = StakePending
    stake.CreatedAt = time.Now()
    stake.UpdatedAt = time.Now()

    // 8. Insert stake
    _, err = s.stakeCollection.InsertOne(ctx, stake)
    return err
}

func (s *stakeService) GetStake(ctx context.Context, stakeID primitive.ObjectID) (*GameStake, error) {
    var stake GameStake
    err := s.stakeCollection.FindOne(ctx, bson.M{"_id": stakeID}).Decode(&stake)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrStakeNotFound
        }
        return nil, err
    }
    return &stake, nil
}

func (s *stakeService) ListStakes(ctx context.Context, gameID primitive.ObjectID) ([]GameStake, error) {
    cursor, err := s.stakeCollection.Find(ctx, bson.M{"game_id": gameID})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var stakes []GameStake
    if err = cursor.All(ctx, &stakes); err != nil {
        return nil, err
    }
    return stakes, nil
}

func (s *stakeService) UpdateStakeStatus(ctx context.Context, stakeID primitive.ObjectID, status string) error {
    update := bson.M{
        "$set": bson.M{
            "status": status,
            "updated_at": time.Now(),
        },
    }
    
    result, err := s.stakeCollection.UpdateOne(ctx, bson.M{"_id": stakeID}, update)
    if err != nil {
        return err
    }
    
    if result.ModifiedCount == 0 {
        return ErrStakeNotFound
    }
    
    return nil
}