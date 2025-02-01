// core/payout/model.go
package payout

import (
    "errors"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
    ErrInvalidChannel = errors.New("invalid payout channel type")
    ErrInvalidCurrency = errors.New("invalid currency")
    ErrInvalidBankDetails = errors.New("invalid bank account details")
    ErrInvalidWalletAddress = errors.New("invalid wallet address")
)

type ChannelType string
const (
    ChannelBank ChannelType = "bank_account"
    ChannelWallet ChannelType = "wallet"
)

type PayoutChannel struct {
    ID          primitive.ObjectID `bson:"_id"`
    UserID      string    `json:"user_id" bson:"user_id"`
    ChannelType ChannelType `json:"channel_type" bson:"channel_type"`
    Currency    string    `json:"currency" bson:"currency"`
    
    // Bank specific fields
    BankID      string    `json:"bank_id,omitempty" bson:"bank_id,omitempty"`
    AccountNo   string    `json:"account_no,omitempty" bson:"account_no,omitempty"`
    
    // Wallet specific fields
    WalletAddress string   `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
    ChainID      string    `json:"chain_id,omitempty" bson:"chain_id,omitempty"`
    
    CreatedAt   time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
    IsActive    bool      `json:"is_active" bson:"is_active"`
}

func (pc *PayoutChannel) Validate() error {
    if pc.UserID == "" {
        return errors.New("user ID is required")
    }

    if pc.Currency == "" {
        return errors.New("currency is required")
    }

    switch pc.ChannelType {
    case ChannelBank:
        if pc.BankID == "" || pc.AccountNo == "" {
            return ErrInvalidBankDetails
        }
    case ChannelWallet:
        if pc.WalletAddress == "" {
            return ErrInvalidWalletAddress
        }
        if pc.ChainID == "" {
            return errors.New("chain ID is required for wallet channels")
        }
    default:
        return ErrInvalidChannel
    }

    return nil
}