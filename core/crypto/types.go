package crypto

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidAddress         = errors.New("invalid wallet address")
	ErrUnsupportedChain       = errors.New("unsupported blockchain")
	ErrInsufficientFunds      = errors.New("insufficient funds for transaction")
	ErrAddressGenerationFailed = errors.New("failed to generate address")
	ErrInvalidPrivateKey      = errors.New("invalid private key")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrDecryptionFailed       = errors.New("failed to decrypt private key")
)

// ChainType represents supported blockchain types
type ChainType string

const (
	ChainEVM    ChainType = "evm"    // Ethereum, BSC, Polygon, etc.
	ChainTron   ChainType = "tron"   // Tron blockchain
	ChainSolana ChainType = "solana" // Solana blockchain
)

// TransactionStatus represents the status of a crypto transaction
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
)

// WalletAddress represents a blockchain wallet address with its associated data
type PayinDetails struct {
	ID            string      `json:"id" bson:"_id,omitempty"`
	UserID        string      `json:"user_id" bson:"user_id"`
	Address       string      `json:"address" bson:"address"`
	ChainType     ChainType   `json:"chain_type" bson:"chain_type"`
	EncryptedKey  string      `json:"encrypted_key" bson:"encrypted_key"` // Encrypted private key
	CreatedAt     time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" bson:"updated_at"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID              string            `json:"id" bson:"_id,omitempty"`
	UserID          string            `json:"user_id" bson:"user_id"`
	FromAddress     string            `json:"from_address" bson:"from_address"`
	ToAddress       string            `json:"to_address" bson:"to_address"`
	Amount          string            `json:"amount" bson:"amount"` // String to handle precise decimal values
	Currency        string            `json:"currency" bson:"currency"`
	ChainType       ChainType         `json:"chain_type" bson:"chain_type"`
	TxHash          string            `json:"tx_hash,omitempty" bson:"tx_hash,omitempty"`
	Status          TransactionStatus `json:"status" bson:"status"`
	Fee             string            `json:"fee,omitempty" bson:"fee,omitempty"`
	Confirmations   int               `json:"confirmations,omitempty" bson:"confirmations,omitempty"`
	CreatedAt       time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" bson:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	MetaData        map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// TransferParams contains parameters for a crypto transfer
type TransferParams struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`
	Currency    string    `json:"currency"`
	ChainType   ChainType `json:"chain_type"`
	GasLimit    uint64    `json:"gas_limit,omitempty"` // For EVM chains
	GasPrice    string    `json:"gas_price,omitempty"` // For EVM chains
	Memo        string    `json:"memo,omitempty"`      // For Tron and Solana
}

// FeeEstimate represents estimated transaction fees
type FeeEstimate struct {
	EstimatedFee    string `json:"estimated_fee"`
	Currency        string `json:"currency"`
	BaseFee         string `json:"base_fee,omitempty"`         // For EVM
	PriorityFee     string `json:"priority_fee,omitempty"`     // For EVM
	MaxFee          string `json:"max_fee,omitempty"`          // For EVM
	EstimatedTimeMin int    `json:"estimated_time_min,omitempty"` // Minimum time in minutes
	EstimatedTimeMax int    `json:"estimated_time_max,omitempty"` // Maximum time in minutes
}

// BalanceInfo represents a wallet balance
type BalanceInfo struct {
	Address   string            `json:"address"`
	Currency  string            `json:"currency"`
	Balance   string            `json:"balance"`
	Available string            `json:"available,omitempty"` // May differ from balance for staked assets
	ChainType ChainType         `json:"chain_type"`
	Timestamp time.Time         `json:"timestamp"`
}

// CryptoService defines the interface for cryptocurrency operations
type CryptoService interface {
	// Address generation and management
	GenerateAddress(ctx context.Context, userID string, chainType ChainType) (*PayinDetails, error)
	// GetUserAddress(ctx context.Context, userID string, chainType ChainType) (*PayinDetails, error)
	ValidateAddress(ctx context.Context, address string, chainType ChainType) (bool, error)
	
	// // Transaction operations
	// EstimateFee(ctx context.Context, params TransferParams) (*FeeEstimate, error)
	// Transfer(ctx context.Context, userID string, params TransferParams) (*Transaction, error)
	// GetTransactionStatus(ctx context.Context, txID string) (*Transaction, error)
	
	// Balance operations
	GetBalance(ctx context.Context, address string, currency string, chainType ChainType) (*BalanceInfo, error)
}