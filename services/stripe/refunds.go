package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/refund"
)

// RefundService handles Stripe refund operations
type RefundService struct {
	validator *validator.Validate
}

// NewRefundService creates a new refund service
func NewRefundService() *RefundService {
	return &RefundService{
		validator: validator.New(),
	}
}

// RefundRequest represents a request to create a refund
type RefundRequest struct {
	ChargeID string            `json:"charge_id" validate:"required"`
	Amount   int64             `json:"amount,omitempty"` // Optional, if not provided, refunds entire charge
	Reason   string            `json:"reason,omitempty"` // requested_by_customer, duplicate, fraudulent
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Refund represents a Stripe refund
type Refund struct {
	ID        string            `json:"id"`
	ChargeID  string            `json:"charge_id"`
	Amount    int64             `json:"amount"`
	Currency  string            `json:"currency"`
	Status    string            `json:"status"`
	Reason    string            `json:"reason,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateRefund creates a new refund using Stripe
func (s *RefundService) CreateRefund(ctx context.Context, request *RefundRequest) (*Refund, error) {
	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create Stripe refund params
	params := &stripe.RefundParams{
		Charge: stripe.String(request.ChargeID),
		Reason: stripe.String(request.Reason),
	}

	// Set amount if provided (partial refund)
	if request.Amount > 0 {
		params.Amount = stripe.Int64(request.Amount)
	}

	// Set metadata if provided
	if len(request.Metadata) > 0 {
		params.Metadata = request.Metadata
	}

	// Create the refund
	stripeRefund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	// Convert to our Refund type
	refund := &Refund{
		ID:        stripeRefund.ID,
		ChargeID:  stripeRefund.Charge.ID,
		Amount:    stripeRefund.Amount,
		Currency:  string(stripeRefund.Currency),
		Status:    string(stripeRefund.Status),
		Reason:    string(stripeRefund.Reason),
		Metadata:  stripeRefund.Metadata,
		CreatedAt: time.Unix(stripeRefund.Created, 0),
		UpdatedAt: time.Unix(stripeRefund.Created, 0), // Stripe doesn't provide updated_at for refunds
	}

	return refund, nil
}

// GetRefund retrieves a refund by ID
func (s *RefundService) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	if refundID == "" {
		return nil, fmt.Errorf("refund ID is required")
	}

	// Retrieve the refund from Stripe
	stripeRefund, err := refund.Get(refundID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Stripe refund: %w", err)
	}

	// Convert to our Refund type
	refund := &Refund{
		ID:        stripeRefund.ID,
		ChargeID:  stripeRefund.Charge.ID,
		Amount:    stripeRefund.Amount,
		Currency:  string(stripeRefund.Currency),
		Status:    string(stripeRefund.Status),
		Reason:    string(stripeRefund.Reason),
		Metadata:  stripeRefund.Metadata,
		CreatedAt: time.Unix(stripeRefund.Created, 0),
		UpdatedAt: time.Unix(stripeRefund.Created, 0),
	}

	return refund, nil
}

// ListRefunds lists refunds for a specific charge
func (s *RefundService) ListRefunds(ctx context.Context, chargeID string, limit int) ([]*Refund, error) {
	if chargeID == "" {
		return nil, fmt.Errorf("charge ID is required")
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}

	// Create Stripe list params
	params := &stripe.RefundListParams{
		Charge: stripe.String(chargeID),
	}

	// List refunds from Stripe
	iter := refund.List(params)
	var refunds []*Refund

	for iter.Next() {
		stripeRefund := iter.Refund()

		// Convert to our Refund type
		refund := &Refund{
			ID:        stripeRefund.ID,
			ChargeID:  stripeRefund.Charge.ID,
			Amount:    stripeRefund.Amount,
			Currency:  string(stripeRefund.Currency),
			Status:    string(stripeRefund.Status),
			Reason:    string(stripeRefund.Reason),
			Metadata:  stripeRefund.Metadata,
			CreatedAt: time.Unix(stripeRefund.Created, 0),
			UpdatedAt: time.Unix(stripeRefund.Created, 0),
		}

		refunds = append(refunds, refund)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe refunds: %w", err)
	}

	return refunds, nil
}

// ValidateRefundRequest validates a refund request
func (s *RefundService) ValidateRefundRequest(request *RefundRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional business logic validation
	if request.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	// Validate reason if provided
	if request.Reason != "" {
		validReasons := map[string]bool{
			"requested_by_customer": true,
			"duplicate":             true,
			"fraudulent":            true,
		}
		if !validReasons[request.Reason] {
			return fmt.Errorf("invalid refund reason: %s", request.Reason)
		}
	}

	return nil
}
