package payment

import (
	"context"
	"fmt"
	providers "gambl/providers/payment"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentService interface {
	// Payment Management
	CreatePayment(ctx context.Context, payment *Payment) (*Payment, error)
	GetOnePayment(ctx context.Context, filter interface{}) (*Payment, error)
	UpdatePayment(ctx context.Context, filter interface{}, updates interface{}) error
	ListPayment(ctx context.Context, filters PaymentFilters) ([]Payment, error)
	HandlePaymentCallback(provider string, reference string) (*providers.PaymentResponse, error)
	CreatePaymentLink(payment *CreatePaymentLink, provider string) (*providers.FlutterwavePaymentResponse, error)
}

type PaymentFilters struct {
	Status   []PaymentStatus
	Type     string
	UserID   string
	GameID   string
	FromDate time.Time
	ToDate   time.Time
	Limit    int
	Offset   int
}

type StakeFilters struct {
	StakeID       []string
	MinAmount     float64
	MaxAmount     float64
	Currency      []string
	PayoutChannel []string
	FromDate      time.Time
	ToDate        time.Time
	Limit         int
	Offset        int
}

// type Payment struct {
// 	Amount      float64
// 	Email       string
// 	Name        string
// 	PhoneNumber string
// }

type paymentService struct {
	Collection          *mongo.Collection
	PaystackProvider    *providers.PaystackProvider
	FlutterwaveProvider *providers.FlutterwaveProvider
}

// CreatePayment implements PaymentService.
func (p *paymentService) CreatePayment(ctx context.Context, payment *Payment) (*Payment, error) {
	// Implement the logic for creating a payment

	filter := bson.M{"user_id": payment.UserID, "game_id": payment.GameID, "status": StatusSuccessful}

	num, _ := p.CountPayments(ctx, filter)

	if num > 0 {
		return nil, fmt.Errorf("user has paid for this game")
	}

	// Set initial payment status
	payment.Status = StatusPending
	payment.ID = primitive.NewObjectID()
	payment.Created_At = time.Now()
	payment.Updated_At = time.Now()

	// gameJSON, _ := json.MarshalIndent(payment, "", "  ")
	// log.Printf("payment details Before Creation:\n%s", string(gameJSON))
	_, err := p.Collection.InsertOne(ctx, payment)

	return payment, err
}

func (p *paymentService) CreatePaymentLink(payment *CreatePaymentLink, provider string) (*providers.FlutterwavePaymentResponse, error) {
	switch provider {
	case "paystack":
		// return s.PaystackProvider.OnSuccessful(reference)
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	case "flutterwave":
		response, err := p.FlutterwaveProvider.CreatePaymentLink(payment.Amount, payment.Email, payment.Name, payment.PhoneNumber, payment.Currency, payment.GameID, payment.UserID, payment.PaymentID)
		return response, err
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}
}

func (p *paymentService) UpdatePayment(ctx context.Context, filter interface{}, updates interface{}) error {
	// Implement the logic for updating a payment
	doc, err := p.Collection.UpdateOne(ctx, filter, bson.M{"$set": updates})

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	if doc.MatchedCount == 0 {
		return fmt.Errorf("no documents matched the filter")
	}
	if doc.ModifiedCount == 0 {
		return fmt.Errorf("document was matched but not updated (possibly no change)")
	}

	return err
}

func (s *paymentService) HandlePaymentCallback(provider string, reference string) (*providers.PaymentResponse, error) {
	switch provider {
	case "paystack":
		return s.PaystackProvider.OnSuccessful(reference)
	case "flutterwave":
		return s.FlutterwaveProvider.OnSuccessful(reference)
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}
}

// GetOnePayment implements PaymentService.
func (p *paymentService) GetOnePayment(ctx context.Context, filter interface{}) (*Payment, error) {
	var payment Payment
	err := p.Collection.FindOne(ctx, filter).Decode(&payment)
	return &payment, err
}

// Count documents implements PaymentService.
func (p *paymentService) CountPayments(ctx context.Context, filter interface{}) (int64, error) {
	num, err := p.Collection.CountDocuments(ctx, filter)
	return num, err
}

// ListPayment implements PaymentService.
func (p *paymentService) ListPayment(ctx context.Context, filters PaymentFilters) ([]Payment, error) {
	panic("unimplemented")
}

func NewPaymentService(collection *mongo.Collection, paystackProvider *providers.PaystackProvider,
	flutterwaveProvider *providers.FlutterwaveProvider) PaymentService {
	log.Printf("Init: create payment service")
	return &paymentService{Collection: collection, PaystackProvider: paystackProvider,
		FlutterwaveProvider: flutterwaveProvider}
}
