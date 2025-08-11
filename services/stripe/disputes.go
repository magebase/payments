package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// DisputeService handles Stripe dispute operations
type DisputeService struct {
	validator *validator.Validate
}

// NewDisputeService creates a new dispute service
func NewDisputeService() *DisputeService {
	return &DisputeService{
		validator: validator.New(),
	}
}

// DisputeRequest represents a request to create a dispute
type DisputeRequest struct {
	ChargeID string            `json:"charge_id" validate:"required"`
	Amount   int64             `json:"amount" validate:"required,min=1"`
	Reason   string            `json:"reason" validate:"required"`
	Evidence map[string]string `json:"evidence,omitempty"`
}

// Dispute represents a Stripe dispute
type Dispute struct {
	ID        string            `json:"id"`
	ChargeID  string            `json:"charge_id"`
	Amount    int64             `json:"amount"`
	Currency  string            `json:"currency"`
	Status    string            `json:"status"`
	Reason    string            `json:"reason"`
	Evidence  map[string]string `json:"evidence,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateDispute creates a new dispute using Stripe
func (s *DisputeService) CreateDispute(ctx context.Context, request *DisputeRequest) (*Dispute, error) {
	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate dispute reason
	if err := s.validateDisputeReason(request.Reason); err != nil {
		return nil, err
	}

	// Note: Stripe doesn't have a direct API to create disputes
	// Disputes are typically created by customers through their bank
	// This method would be used for internal dispute tracking
	// For now, we'll create a mock dispute for testing purposes

	// In a real implementation, you would:
	// 1. Verify the charge exists
	// 2. Check if a dispute already exists
	// 3. Create a dispute record in your database
	// 4. Set up webhook handling for dispute events

	mockDispute := &Dispute{
		ID:        fmt.Sprintf("dp_%d", time.Now().Unix()),
		ChargeID:  request.ChargeID,
		Amount:    request.Amount,
		Currency:  "usd", // Default currency
		Status:    "needs_response",
		Reason:    request.Reason,
		Evidence:  request.Evidence,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return mockDispute, nil
}

// GetDispute retrieves a dispute by ID
func (s *DisputeService) GetDispute(ctx context.Context, disputeID string) (*Dispute, error) {
	if disputeID == "" {
		return nil, fmt.Errorf("dispute ID is required")
	}

	// In a real implementation, you would:
	// 1. Query your database for the dispute
	// 2. Optionally sync with Stripe if the dispute exists there

	// For now, return a mock dispute for testing
	mockDispute := &Dispute{
		ID:        disputeID,
		ChargeID:  "ch_mock",
		Amount:    1000,
		Currency:  "usd",
		Status:    "needs_response",
		Reason:    "fraudulent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return mockDispute, nil
}

// ListDisputes lists disputes for a specific charge
func (s *DisputeService) ListDisputes(ctx context.Context, chargeID string, limit int) ([]*Dispute, error) {
	if chargeID == "" {
		return nil, fmt.Errorf("charge ID is required")
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}

	// In a real implementation, you would:
	// 1. Query your database for disputes related to the charge
	// 2. Optionally sync with Stripe for the latest status

	// For now, return mock disputes for testing
	mockDisputes := []*Dispute{
		{
			ID:        "dp_1234567890",
			ChargeID:  chargeID,
			Amount:    1000,
			Currency:  "usd",
			Status:    "needs_response",
			Reason:    "fraudulent",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "dp_1234567891",
			ChargeID:  chargeID,
			Amount:    500,
			Currency:  "usd",
			Status:    "under_review",
			Reason:    "duplicate",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return mockDisputes, nil
}

// UpdateDisputeStatus updates the status of a dispute
func (s *DisputeService) UpdateDisputeStatus(ctx context.Context, disputeID string, status string) (*Dispute, error) {
	if disputeID == "" {
		return nil, fmt.Errorf("dispute ID is required")
	}

	if status == "" {
		return nil, fmt.Errorf("status is required")
	}

	// Validate dispute status
	if err := s.validateDisputeStatus(status); err != nil {
		return nil, err
	}

	// In a real implementation, you would:
	// 1. Update the dispute status in your database
	// 2. Optionally sync with Stripe if the dispute exists there
	// 3. Trigger any necessary workflows based on status change

	// For now, return a mock dispute with updated status for testing
	mockDispute := &Dispute{
		ID:        disputeID,
		ChargeID:  "ch_mock",
		Amount:    1000,
		Currency:  "usd",
		Status:    status,
		Reason:    "fraudulent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return mockDispute, nil
}

// ValidateDisputeRequest validates a dispute request
func (s *DisputeService) ValidateDisputeRequest(request *DisputeRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Additional business logic validation
	if request.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// Validate reason
	return s.validateDisputeReason(request.Reason)
}

// validateDisputeReason validates the dispute reason
func (s *DisputeService) validateDisputeReason(reason string) error {
	validReasons := map[string]bool{
		"fraudulent":           true,
		"duplicate":            true,
		"product_not_received": true,
		"product_unacceptable": true,
		"incorrect_amount":     true,
		"credit_not_processed": true,
		"general":              true,
	}

	if !validReasons[reason] {
		return fmt.Errorf("invalid dispute reason: %s", reason)
	}

	return nil
}

// validateDisputeStatus validates the dispute status
func (s *DisputeService) validateDisputeStatus(status string) error {
	validStatuses := map[string]bool{
		"needs_response":         true,
		"under_review":           true,
		"won":                    true,
		"lost":                   true,
		"warning_needs_response": true,
		"warning_under_review":   true,
		"warning_closed":         true,
		"responded":              true,
		"closed":                 true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid dispute status: %s", status)
	}

	return nil
}
