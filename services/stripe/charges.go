package stripe

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/charge"
)

// ChargeService handles Stripe charge operations
type ChargeService struct {
	validator *validator.Validate
}

// NewChargeService creates a new charge service
func NewChargeService() *ChargeService {
	return &ChargeService{
		validator: validator.New(),
	}
}

// CreateCharge creates a new charge using Stripe
func (s *ChargeService) CreateCharge(ctx context.Context, request *ChargeRequest) (*Charge, error) {
	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Additional business logic validation
	if request.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Convert to Stripe charge params
	params := &stripe.ChargeParams{
		Amount:      stripe.Int64(request.Amount),
		Currency:    stripe.String(request.Currency),
		Customer:    stripe.String(request.CustomerID),
		Description: stripe.String(request.Description),
	}

	// Set source using the proper method
	if err := params.SetSource(request.Source); err != nil {
		return nil, fmt.Errorf("failed to set source: %w", err)
	}

	// Create the charge
	stripeCharge, err := charge.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe charge: %w", err)
	}

	// Convert to our Charge type
	charge := &Charge{
		ID:          stripeCharge.ID,
		Amount:      stripeCharge.Amount,
		Currency:    string(stripeCharge.Currency),
		Status:      string(stripeCharge.Status),
		CustomerID:  stripeCharge.Customer.ID,
		Description: stripeCharge.Description,
		Created:     stripeCharge.Created,
	}

	return charge, nil
}

// ChargeRequest represents a request to create a charge
type ChargeRequest struct {
	Amount      int64  `json:"amount" validate:"required,min=1"`
	Currency    string `json:"currency" validate:"required"`
	CustomerID  string `json:"customer_id" validate:"required"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source" validate:"required"`
}

// Charge represents a Stripe charge
type Charge struct {
	ID              string            `json:"id"`
	Amount          int64             `json:"amount"`
	Currency        string            `json:"currency"`
	Status          string            `json:"status"`
	CustomerID      string            `json:"customer_id"`
	PaymentMethodID string            `json:"payment_method_id,omitempty"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Created         int64             `json:"created"`
}

// ValidateChargeRequest validates a charge request
func (s *ChargeService) ValidateChargeRequest(request *ChargeRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional business logic validation
	if request.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if request.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if request.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}

	if request.Source == "" {
		return fmt.Errorf("source is required")
	}

	return nil
}

// GetCharge retrieves a charge by ID
func (s *ChargeService) GetCharge(ctx context.Context, chargeID string) (*Charge, error) {
	if chargeID == "" {
		return nil, fmt.Errorf("charge ID cannot be empty")
	}

	params := &stripe.ChargeParams{}

	stripeCharge, err := charge.Get(chargeID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve charge: %w", err)
	}

	charge := &Charge{
		ID:          stripeCharge.ID,
		Amount:      stripeCharge.Amount,
		Currency:    string(stripeCharge.Currency),
		Status:      string(stripeCharge.Status),
		CustomerID:  stripeCharge.Customer.ID,
		Description: stripeCharge.Description,
		Created:     stripeCharge.Created,
	}

	return charge, nil
}

// ListCharges retrieves a list of charges
func (s *ChargeService) ListCharges(ctx context.Context, customerID string, limit int64) ([]*Charge, error) {
	params := &stripe.ChargeListParams{}

	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}

	iter := charge.List(params)
	var charges []*Charge

	for iter.Next() {
		stripeCharge := iter.Charge()
		charge := &Charge{
			ID:          stripeCharge.ID,
			Amount:      stripeCharge.Amount,
			Currency:    string(stripeCharge.Currency),
			Status:      string(stripeCharge.Status),
			CustomerID:  stripeCharge.Customer.ID,
			Description: stripeCharge.Description,
			Created:     stripeCharge.Created,
		}
		charges = append(charges, charge)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list charges: %w", err)
	}

	return charges, nil
}

// FormatAmount formats an amount in cents to a human-readable string
func (s *ChargeService) FormatAmount(amount int64, currency string) string {
	// Convert cents to dollars
	dollars := float64(amount) / 100.0

	// Format based on currency
	switch currency {
	case "usd":
		return fmt.Sprintf("$%.2f", dollars)
	case "eur":
		return fmt.Sprintf("€%.2f", dollars)
	case "gbp":
		return fmt.Sprintf("£%.2f", dollars)
	default:
		return fmt.Sprintf("%.2f %s", dollars, currency)
	}
}

// ParseAmount parses a human-readable amount string to cents
func (s *ChargeService) ParseAmount(amountStr string) (int64, error) {
	// Remove currency symbols and spaces
	cleanStr := amountStr
	for _, symbol := range []string{"$", "€", "£", " "} {
		cleanStr = strings.ReplaceAll(cleanStr, symbol, "")
	}

	// Parse as float
	amount, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %w", err)
	}

	// Convert to cents
	cents := int64(amount * 100)
	return cents, nil
}
