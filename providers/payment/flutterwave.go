package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FlutterwaveProvider struct {
	SecretKey string
	PublicKey string
}

// NewFlutterwaveProvider creates a new instance of FlutterwaveProvider
func NewFlutterwaveProvider(secretKey, publicKey string) *FlutterwaveProvider {
	return &FlutterwaveProvider{
		SecretKey: secretKey,
		PublicKey: publicKey,
	}
}

// OnSuccessful handles successful payment callback from Flutterwave
func (f *FlutterwaveProvider) OnSuccessful(reference string) (*PaymentResponse, error) {
	// Implement Flutterwave-specific verification logic here
	// This could include:
	// 1. Verifying the transaction with Flutterwave API
	// 2. Processing the payment data
	// 3. Updating necessary records

	return &PaymentResponse{
		Reference: reference,
		Amount:    0.0, // Replace with actual amount from verification
		Status:    "success",
	}, nil
}

func (f *FlutterwaveProvider) CreatePaymentLink(amount string, email, name, phoneNumber string, currency string, gameId string, userId string, paymentId string) (*FlutterwavePaymentResponse, error) {
	// Create unique transaction reference
	txRef := fmt.Sprintf("tx-%d", time.Now().Unix())

	paymentRequest := FlutterwavePaymentRequest{
		TxRef:       txRef,
		Amount:      amount,
		Currency:    currency,
		RedirectURL: "https://example_company.com/success",
		Customer: Customer{
			Email:       email,
			Name:        name,
			PhoneNumber: phoneNumber,
		},
		Customization: Customization{
			Title: "Flutterwave Standard Payment",
		},
		Meta: map[string]interface{}{
			"gameID":    gameId,
			"userID":    userId,
			"paymentID": paymentId,
		},
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	const FLUTTERWAVE_API_URL = "https://api.flutterwave.com/v3/payments"

	// Create HTTP request
	req, err := http.NewRequest("POST", FLUTTERWAVE_API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.SecretKey))
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, fmt.Errorf("error decoding error response: %v", err)
		}
		return nil, fmt.Errorf("flutterwave API error: %v", errorResponse)
	}

	// Parse response
	var paymentResponse FlutterwavePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &paymentResponse, nil
}
