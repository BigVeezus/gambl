package providers


import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gambl/core/payout"
	"io"
	"net/http"
)

type PaystackProvider struct {
	SecretKey string
	PublicKey string
	baseURL   string
	client    *http.Client
}

// PaystackBank represents bank information from Paystack API
type PaystackBank struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Slug      string `json:"slug"`
	Active    bool   `json:"active"`
	Country   string `json:"country"`
	Currency  string `json:"currency"`
}

// PaystackVerifyResponse represents account verification response
type PaystackVerifyResponse struct {
	Status  bool `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		BankID        int    `json:"bank_id"`
	} `json:"data"`
}

// PaystackBanksResponse represents the list of banks response
type PaystackBanksResponse struct {
	Status  bool          `json:"status"`
	Message string        `json:"message"`
	Data    []PaystackBank `json:"data"`
}


// NewPaystackProvider creates a new instance of PaystackProvider
func NewPaystackProvider(secretKey, publicKey string) *PaystackProvider {
	return &PaystackProvider{
		SecretKey: secretKey,
		PublicKey: publicKey,
	}
}

// OnSuccessful handles successful payment callback from Paystack
func (p *PaystackProvider) OnSuccessful(reference string) (*PaymentResponse, error) {
	// Implement Paystack-specific verification logic here
	// This could include:
	// 1. Verifying the transaction with Paystack API
	// 2. Processing the payment data
	// 3. Updating necessary records

	return &PaymentResponse{
		Reference: reference,
		Amount:    0.0, // Replace with actual amount from verification
		Status:    "success",
	}, nil
}


// VerifyBankAccount verifies a bank account using Paystack API
func (p *PaystackProvider) VerifyBankAccount(ctx context.Context, bankCode, accountNumber string) (string, error) {
	url := fmt.Sprintf("%s/bank/resolve?account_number=%s&bank_code=%s", p.baseURL, accountNumber, bankCode)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paystack API error: %s", string(body))
	}
	
	var response PaystackVerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	
	if !response.Status {
		return "", errors.New(response.Message)
	}
	
	return response.Data.AccountName, nil
}

// GetBanks retrieves the list of supported banks from Paystack API
func (p *PaystackProvider) GetBanks(ctx context.Context) ([]payout.Bank, error) {
	url := fmt.Sprintf("%s/bank", p.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paystack API error: %s", string(body))
	}
	
	var response PaystackBanksResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	
	if !response.Status {
		return nil, errors.New(response.Message)
	}
	
	// Convert to payout.Bank format
	banks := make([]payout.Bank, len(response.Data))
	for i, bank := range response.Data {
		banks[i] = payout.Bank{
			ID:   fmt.Sprintf("%d", bank.ID),
			Name: bank.Name,
			Code: bank.Code,
		}
	}
	
	return banks, nil
}
