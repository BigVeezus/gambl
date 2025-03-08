// core/payout/service.go
package payout

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Common errors
var (
	ErrPayoutChannelNotFound = errors.New("payout channel not found")
	ErrPayoutChannelExists   = errors.New("payout channel already exists with same currency and type")
	ErrNoBankProvider        = errors.New("no bank verification provider available")
)

// PayoutChannelService defines all payout channel operations
type PayoutChannelService interface {
	// CRUD operations
	CreatePayoutChannel(ctx context.Context, channel *PayoutChannel) error
	GetPayoutChannel(ctx context.Context, id string) (*PayoutChannel, error)
	GetUserPayoutChannels(ctx context.Context, userID string) ([]PayoutChannel, error)
	UpdatePayoutChannel(ctx context.Context, id string, updates map[string]interface{}) error
	DeletePayoutChannel(ctx context.Context, id string) error

	// Utility operations
	VerifyBankAccount(ctx context.Context, bankID, accountNo string) (string, error) // Returns account name
	SetDefaultPayoutChannel(ctx context.Context, userID, channelID string) error
	GetDefaultPayoutChannel(ctx context.Context, userID string) (*PayoutChannel, error)
	GetUserPayoutChannelByCurrency(ctx context.Context, userID, currency string) (*PayoutChannel, error)
	HasValidPayoutChannel(ctx context.Context, userID string) (bool, error)
}

// BankVerificationProvider is an interface for services that verify bank accounts
type BankVerificationProvider interface {
	VerifyBankAccount(ctx context.Context, bankCode, accountNumber string) (string, error)
	GetBanks(ctx context.Context) ([]Bank, error)
}

// Bank represents bank information
type Bank struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
	Code string `json:"code" bson:"code"`
}

// payoutChannelService implements PayoutChannelService
type payoutChannelService struct {
	collection      *mongo.Collection
	bankProviders   map[string]BankVerificationProvider // Different providers for different regions/countries
	defaultProvider string                              // Default provider key
}

// NewPayoutChannelService creates a new payout channel service
func NewPayoutChannelService(collection *mongo.Collection, providers map[string]BankVerificationProvider, defaultProvider string) PayoutChannelService {
	return &payoutChannelService{
		collection:      collection,
		bankProviders:   providers,
		defaultProvider: defaultProvider,
	}
}

// CreatePayoutChannel creates a new payout channel
func (s *payoutChannelService) CreatePayoutChannel(ctx context.Context, channel *PayoutChannel) error {
	// Validate the payout channel
	if err := channel.Validate(); err != nil {
		return err
	}

	// Check if a channel with the same type and currency already exists for this user
	existing, err := s.getUserPayoutChannelByTypeAndCurrency(ctx, channel.UserID, channel.ChannelType, channel.Currency)
	if err != nil && err != ErrPayoutChannelNotFound {
		return err
	}
	
	if existing != nil {
		return ErrPayoutChannelExists
	}

	// For bank accounts, verify the account number if provider is available
	if channel.ChannelType == ChannelBank {
		if len(s.bankProviders) > 0 {
			provider := s.bankProviders[s.defaultProvider]
			accountName, err := provider.VerifyBankAccount(ctx, channel.BankID, channel.AccountNo)
			if err != nil {
				return err
			}
			
			// You could store the account name if needed
			// channel.AccountName = accountName
			log.Printf("Verified bank account: %s", accountName)
		}
	}

	// Set metadata
	channel.ID = primitive.NewObjectID()
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	channel.IsActive = true

	// Check if this is the first channel for the user; if so, make it default
	count, err := s.collection.CountDocuments(ctx, bson.M{"user_id": channel.UserID})
	if err != nil {
		return err
	}
	
	isDefault := count == 0
	
	// Insert the channel
	_, err = s.collection.InsertOne(ctx, channel)
	if err != nil {
		return err
	}
	
	// If this is the first channel, make it default
	if isDefault {
		return s.SetDefaultPayoutChannel(ctx, channel.UserID, channel.ID.Hex())
	}
	
	return nil
}

// GetPayoutChannel gets a payout channel by ID
func (s *payoutChannelService) GetPayoutChannel(ctx context.Context, id string) (*PayoutChannel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var channel PayoutChannel
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&channel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPayoutChannelNotFound
		}
		return nil, err
	}

	return &channel, nil
}

// GetUserPayoutChannels gets all payout channels for a user
func (s *payoutChannelService) GetUserPayoutChannels(ctx context.Context, userID string) ([]PayoutChannel, error) {
	cursor, err := s.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"is_active": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var channels []PayoutChannel
	if err = cursor.All(ctx, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

// UpdatePayoutChannel updates a payout channel
func (s *payoutChannelService) UpdatePayoutChannel(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Add updated timestamp
	updates["updated_at"] = time.Now()

	// Execute update
	_, err = s.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return err
	}

	return nil
}

// DeletePayoutChannel soft-deletes a payout channel by setting isActive to false
func (s *payoutChannelService) DeletePayoutChannel(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Perform soft delete
	_, err = s.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}

	return nil
}

// VerifyBankAccount verifies a bank account
func (s *payoutChannelService) VerifyBankAccount(ctx context.Context, bankID, accountNo string) (string, error) {
	if len(s.bankProviders) == 0 {
		return "", ErrNoBankProvider
	}

	provider := s.bankProviders[s.defaultProvider]
	return provider.VerifyBankAccount(ctx, bankID, accountNo)
}

// SetDefaultPayoutChannel sets a payout channel as default for a user
func (s *payoutChannelService) SetDefaultPayoutChannel(ctx context.Context, userID, channelID string) error {
	objectID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		return err
	}

	// First, unset any existing default channels
	_, err = s.collection.UpdateMany(
		ctx,
		bson.M{"user_id": userID, "is_default": true},
		bson.M{"$set": bson.M{"is_default": false, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}

	// Then set the specified channel as default
	_, err = s.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "user_id": userID},
		bson.M{"$set": bson.M{"is_default": true, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}

	return nil
}

// GetDefaultPayoutChannel gets the default payout channel for a user
func (s *payoutChannelService) GetDefaultPayoutChannel(ctx context.Context, userID string) (*PayoutChannel, error) {
	var channel PayoutChannel
	err := s.collection.FindOne(ctx, bson.M{
		"user_id":    userID,
		"is_default": true,
		"is_active":  true,
	}).Decode(&channel)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPayoutChannelNotFound
		}
		return nil, err
	}

	return &channel, nil
}

// GetUserPayoutChannelByCurrency gets a user's payout channel for a specific currency
func (s *payoutChannelService) GetUserPayoutChannelByCurrency(ctx context.Context, userID, currency string) (*PayoutChannel, error) {
	// First try to find a default one with this currency
	var channel PayoutChannel
	err := s.collection.FindOne(ctx, bson.M{
		"user_id":    userID,
		"currency":   currency,
		"is_default": true,
		"is_active":  true,
	}).Decode(&channel)
	
	if err == nil {
		return &channel, nil
	}
	
	// If no default found, just get any channel with this currency
	if err == mongo.ErrNoDocuments {
		err = s.collection.FindOne(ctx, bson.M{
			"user_id":   userID,
			"currency":  currency,
			"is_active": true,
		}).Decode(&channel)
		
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, ErrPayoutChannelNotFound
			}
			return nil, err
		}
		
		return &channel, nil
	}
	
	return nil, err
}

// HasValidPayoutChannel checks if a user has at least one valid payout channel
func (s *payoutChannelService) HasValidPayoutChannel(ctx context.Context, userID string) (bool, error) {
	count, err := s.collection.CountDocuments(ctx, bson.M{
		"user_id":   userID,
		"is_active": true,
	})
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// Helper method to find payout channel by type and currency
func (s *payoutChannelService) getUserPayoutChannelByTypeAndCurrency(ctx context.Context, userID string, channelType ChannelType, currency string) (*PayoutChannel, error) {
	var channel PayoutChannel
	err := s.collection.FindOne(ctx, bson.M{
		"user_id":      userID,
		"channel_type": channelType,
		"currency":     currency,
		"is_active":    true,
	}).Decode(&channel)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPayoutChannelNotFound
		}
		return nil, err
	}
	
	return &channel, nil
}