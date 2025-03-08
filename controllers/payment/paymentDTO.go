package controllers

import (
	payment "gambl/core/payment"
	"gambl/core/payout"
	"time"
)

// Request DTOs

type CreatePaymentLinkDTO struct {
	payment.CreatePaymentLink
}

type CreatePayoutChannelRequest struct {
	ChannelType   string  `json:"channel_type" binding:"required,oneof=bank_account wallet"`
	Currency      string  `json:"currency" binding:"required"`
	Label         string  `json:"label,omitempty"`
	
	// Bank specific fields
	BankID        string  `json:"bank_id,omitempty"`
	AccountNumber string  `json:"account_number,omitempty"`
	
	// Wallet specific fields
	WalletAddress string  `json:"wallet_address,omitempty"`
	ChainID       string  `json:"chain_id,omitempty"`
}

type UpdatePayoutChannelRequest struct {
	Label         string  `json:"label,omitempty"`
	
	// Bank specific fields - can update bank details if needed
	BankID        string  `json:"bank_id,omitempty"`
	AccountNumber string  `json:"account_number,omitempty"`
	
	// Wallet specific fields - can update wallet address if needed
	WalletAddress string  `json:"wallet_address,omitempty"`
	ChainID       string  `json:"chain_id,omitempty"`
}

type VerifyBankAccountRequest struct {
	BankID        string  `json:"bank_id" binding:"required"`
	AccountNumber string  `json:"account_number" binding:"required"`
}

// Response DTOs

type PayoutChannelResponse struct {
	ID           string    `json:"id"`
	ChannelType  string    `json:"channel_type"`
	Currency     string    `json:"currency"`
	Label        string    `json:"label,omitempty"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	
	// Bank specific fields
	BankID       string    `json:"bank_id,omitempty"`
	AccountNo    string    `json:"account_no,omitempty"`
	AccountName  string    `json:"account_name,omitempty"`
	
	// Wallet specific fields
	WalletAddress string   `json:"wallet_address,omitempty"`
	ChainID       string   `json:"chain_id,omitempty"`
}

// Conversion methods

func (req *CreatePayoutChannelRequest) ToPayoutChannelModel(userID string) *payout.PayoutChannel {
	channelType := payout.ChannelBank
	if req.ChannelType == "wallet" {
		channelType = payout.ChannelWallet
	}
	
	return &payout.PayoutChannel{
		UserID:        userID,
		ChannelType:   channelType,
		Currency:      req.Currency,
		Label:         req.Label,
		
		// Bank specific fields
		BankID:        req.BankID,
		AccountNo:     req.AccountNumber,
		
		// Wallet specific fields
		WalletAddress: req.WalletAddress,
		ChainID:       req.ChainID,
		
		// Default fields
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (req *UpdatePayoutChannelRequest) ToUpdateMap() map[string]interface{} {
	updates := make(map[string]interface{})
	
	if req.Label != "" {
		updates["label"] = req.Label
	}
	
	if req.BankID != "" {
		updates["bank_id"] = req.BankID
	}
	
	if req.AccountNumber != "" {
		updates["account_no"] = req.AccountNumber
	}
	
	if req.WalletAddress != "" {
		updates["wallet_address"] = req.WalletAddress
	}
	
	if req.ChainID != "" {
		updates["chain_id"] = req.ChainID
	}
	
	// Always update the UpdatedAt timestamp
	updates["updated_at"] = time.Now()
	
	return updates
}

// Helper functions

func PayoutChannelToResponse(channel *payout.PayoutChannel) PayoutChannelResponse {
	response := PayoutChannelResponse{
		ID:           channel.ID.Hex(),
		ChannelType:  string(channel.ChannelType),
		Currency:     channel.Currency,
		Label:        channel.Label,
		IsDefault:    channel.IsDefault,
		CreatedAt:    channel.CreatedAt,
	}
	
	// Add channel-type specific fields
	if channel.ChannelType == payout.ChannelBank {
		response.BankID = channel.BankID
		response.AccountNo = channel.AccountNo
		response.AccountName = channel.AccountName
	} else if channel.ChannelType == payout.ChannelWallet {
		response.WalletAddress = channel.WalletAddress
		response.ChainID = channel.ChainID
	}
	
	return response
}
