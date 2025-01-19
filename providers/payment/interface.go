// providers/payment/interface.go
package providers

type PaymentResponse struct {
	Reference string
	Amount    float64
	Status    string
}

type PaymentProvider interface {
	OnSuccessful(reference string) (*PaymentResponse, error)
}

type Customer struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
}

type Customization struct {
	Title string `json:"title"`
}

type FlutterwavePaymentRequest struct {
	TxRef         string        `json:"tx_ref"`
	Amount        string        `json:"amount"`
	Currency      string        `json:"currency"`
	RedirectURL   string        `json:"redirect_url"`
	Customer      Customer      `json:"customer"`
	Customization Customization `json:"customizations"`
	Meta          interface{}   `json:"meta"`
}

type FlutterwavePaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Link string `json:"link"`
	} `json:"data"`
}

type FlutterwaveWebhookResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID                int     `json:"id"`
		TxRef             string  `json:"tx_ref"`
		FlwRef            string  `json:"flw_ref"`
		DeviceFingerprint string  `json:"device_fingerprint"`
		Amount            float64 `json:"amount"`
		Currency          string  `json:"currency"`
		ChargedAmount     float64 `json:"charged_amount"`
		AppFee            float64 `json:"app_fee"`
		MerchantFee       float64 `json:"merchant_fee"`
		ProcessorResponse string  `json:"processor_response"`
		AuthModel         string  `json:"auth_model"`
		IP                string  `json:"ip"`
		Narration         string  `json:"narration"`
		Status            string  `json:"status"`
		PaymentType       string  `json:"payment_type"`
		CreatedAt         string  `json:"created_at"`
		AccountID         int     `json:"account_id"`
		Card              struct {
			First6Digits string `json:"first_6digits"`
			Last4Digits  string `json:"last_4digits"`
			Issuer       string `json:"issuer"`
			Country      string `json:"country"`
			Type         string `json:"type"`
			Token        string `json:"token"`
			Expiry       string `json:"expiry"`
		} `json:"card"`
		Meta struct {
			CheckoutInitAddress string `json:"__CheckoutInitAddress"`
			GameID              string `json:"gameId"`
			UserID              string `json:"userId"`
			PaymentID           string `json:"paymentId"`
		} `json:"meta"`
		AmountSettled float64 `json:"amount_settled"`
		Customer      struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			PhoneNumber string `json:"phone_number"`
			Email       string `json:"email"`
			CreatedAt   string `json:"created_at"`
		} `json:"customer"`
	} `json:"data"`
}
