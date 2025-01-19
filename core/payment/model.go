package payment

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentStatus string

const (
	StatusSuccessful PaymentStatus = "SUCCESSFUL"
	StatusPending    PaymentStatus = "PENDING"
	StatusFailed     PaymentStatus = "FAILED"
)

// At the top of model.go with other error definitions

var (
	// Existing errors
	ErrInvalidAmount        = errors.New("payment amount must be greater than 1")
	ErrInvalidDeadline      = errors.New("deadline must be in the future")
	ErrInvalidCurrency      = errors.New("unsupported currency")
	ErrInvalidPayoutChannel = errors.New("invalid payout channel")

	// Add service-level errors here
	ErrGameNotFound     = errors.New("game not found")
	ErrInvalidGameState = errors.New("invalid game state")
	ErrUnauthorized     = errors.New("unauthorized action")
	ErrStakeNotAllowed  = errors.New("staking not allowed")
	ErrDuplicateStake   = errors.New("user has already staked")
)

type Payment struct {
	ID         primitive.ObjectID `bson:"_id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	GameID     primitive.ObjectID `bson:"game_id" json:"game_id" validate:"required"`
	PaymentRef string             `json:"ref"`
	Currency   string             `json:"currency" validate:"required"`
	Status     PaymentStatus      `json:"status"`
	Created_At time.Time          `json:"created_at"`
	Updated_At time.Time          `json:"updated_at"`
}

func (req *Payment) ToPaymentModel() (*Payment, error) {

	return &Payment{
		UserID:     req.UserID,
		GameID:     req.GameID,
		Currency:   req.Currency,
		Status:     StatusPending,
		Created_At: time.Now(),
		Updated_At: time.Now(),
	}, nil
}

type CreatePaymentLink struct {
	UserID      string `json:"user_id" `
	Amount      string `json:"amount" binding:"required"`
	Email       string `json:"email" binding:"email,required"`
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Currency    string `json:"currency" binding:"required,oneof=USD EUR NGN GBP"`
	GameID      string `json:"game_id" binding:"required"`
	PaymentID   string `json:"payment_id"`
}

type FlutterwaveCustomer struct {
	ID            int       `json:"id"`
	Phone         string    `json:"phone"`
	FullName      string    `json:"fullName"`
	CustomerToken any       `json:"customertoken"` // using 'any' since it's null in the example
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DeletedAt     any       `json:"deletedAt"` // using 'any' since it's null
	AccountID     int       `json:"AccountId"`
}

type FlutterwaveEntity struct {
	Card6          string    `json:"card6"`
	CardLast4      string    `json:"card_last4"`
	CardCountryISO string    `json:"card_country_iso"`
	CreatedAt      time.Time `json:"createdAt"`
}

type FlutterwaveMetaData struct {
	CheckoutInitAddress string `json:"__CheckoutInitAddress"`
}

type FlutterwaveWebhookData struct {
	ID               int                 `json:"id"`
	TxRef            string              `json:"txRef"`
	FlwRef           string              `json:"flwRef"`
	OrderRef         string              `json:"orderRef"`
	PaymentPlan      any                 `json:"paymentPlan"` // using 'any' since it's null
	PaymentPage      any                 `json:"paymentPage"` // using 'any' since it's null
	CreatedAt        time.Time           `json:"createdAt"`
	Amount           float64             `json:"amount"`
	ChargedAmount    float64             `json:"charged_amount"`
	Status           string              `json:"status"`
	IP               string              `json:"IP"`
	Currency         string              `json:"currency"`
	AppFee           float64             `json:"appfee"`
	MerchantFee      float64             `json:"merchantfee"`
	MerchantBearsFee int                 `json:"merchantbearsfee"`
	ChargeType       string              `json:"charge_type"`
	Customer         FlutterwaveCustomer `json:"customer"`
	Entity           FlutterwaveEntity   `json:"entity"`
	MetaData         FlutterwaveMetaData `json:"meta_data"`
}

type FlutterwaveWebhook struct {
	EventType string                 `json:"event.type"`
	Data      FlutterwaveWebhookData `json:"-"` // The data is not nested in this case
}
