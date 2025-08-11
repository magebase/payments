package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// CustomerService handles Stripe customer operations
type CustomerService struct {
	validator *validator.Validate
	tracer    trace.Tracer
}

// NewCustomerService creates a new customer service
func NewCustomerService() *CustomerService {
	return &CustomerService{
		validator: validator.New(),
		tracer:    otel.Tracer("payments.customer"),
	}
}

// CustomerRequest represents a request to create a customer
type CustomerRequest struct {
	Email       string            `json:"email" validate:"required,email"`
	Name        string            `json:"name" validate:"required,min=1"`
	Phone       string            `json:"phone,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Customer represents a Stripe customer
type Customer struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Phone       string            `json:"phone,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Created     int64             `json:"created"`
	Updated     int64             `json:"updated"`
}

// PaymentMethodRequest represents a request to add a payment method
type PaymentMethodRequest struct {
	Type     string            `json:"type" validate:"required,oneof=card sepa_debit ideal sofort"`
	Card     *CardRequest      `json:"card,omitempty"`
	Customer string            `json:"customer" validate:"required"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CardRequest represents card-specific payment method details
type CardRequest struct {
	Token string `json:"token" validate:"required"`
}

// PaymentMethod represents a Stripe payment method
type PaymentMethod struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Card     *Card             `json:"card,omitempty"`
	Customer string            `json:"customer"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Created  int64             `json:"created"`
}

// Card represents card details
type Card struct {
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint"`
}

// CreateCustomer creates a new customer using Stripe
func (s *CustomerService) CreateCustomer(ctx context.Context, request *CustomerRequest) (*Customer, error) {
	ctx, span := s.tracer.Start(ctx, "CreateCustomer")
	defer span.End()

	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe customer params
	params := &stripe.CustomerParams{
		Email:       stripe.String(request.Email),
		Name:        stripe.String(request.Name),
		Phone:       stripe.String(request.Phone),
		Description: stripe.String(request.Description),
		Metadata:    request.Metadata,
	}

	// Create the customer
	stripeCustomer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Convert to our Customer type
	customer := &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     stripeCustomer.Created, // Stripe doesn't provide updated timestamp
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *CustomerService) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	ctx, span := s.tracer.Start(ctx, "GetCustomer")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID cannot be empty")
	}

	params := &stripe.CustomerParams{}
	stripeCustomer, err := customer.Get(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	customer := &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     stripeCustomer.Created,
	}

	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerService) UpdateCustomer(ctx context.Context, customerID string, request *CustomerRequest) (*Customer, error) {
	ctx, span := s.tracer.Start(ctx, "UpdateCustomer")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID cannot be empty")
	}

	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe customer params
	params := &stripe.CustomerParams{
		Email:       stripe.String(request.Email),
		Name:        stripe.String(request.Name),
		Phone:       stripe.String(request.Phone),
		Description: stripe.String(request.Description),
		Metadata:    request.Metadata,
	}

	// Update the customer
	stripeCustomer, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update Stripe customer: %w", err)
	}

	// Convert to our Customer type
	customer := &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     time.Now().Unix(),
	}

	return customer, nil
}

// DeleteCustomer deletes a customer
func (s *CustomerService) DeleteCustomer(ctx context.Context, customerID string) error {
	ctx, span := s.tracer.Start(ctx, "DeleteCustomer")
	defer span.End()

	if customerID == "" {
		return fmt.Errorf("customer ID cannot be empty")
	}

	params := &stripe.CustomerParams{}
	_, err := customer.Del(customerID, params)
	if err != nil {
		return fmt.Errorf("failed to delete Stripe customer: %w", err)
	}

	return nil
}

// AddPaymentMethod adds a payment method to a customer
func (s *CustomerService) AddPaymentMethod(ctx context.Context, request *PaymentMethodRequest) (*PaymentMethod, error) {
	ctx, span := s.tracer.Start(ctx, "AddPaymentMethod")
	defer span.End()

	// Validate the request
	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe payment method params
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(request.Type),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(request.Card.Token),
		},
		Metadata: request.Metadata,
	}

	// Create the payment method
	stripePaymentMethod, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment method: %w", err)
	}

	// Attach to customer
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(request.Customer),
	}
	_, err = paymentmethod.Attach(stripePaymentMethod.ID, attachParams)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method to customer: %w", err)
	}

	// Convert to our PaymentMethod type
	paymentMethod := &PaymentMethod{
		ID:       stripePaymentMethod.ID,
		Type:     string(stripePaymentMethod.Type),
		Customer: request.Customer,
		Metadata: stripePaymentMethod.Metadata,
		Created:  stripePaymentMethod.Created,
	}

	// Add card details if available
	if stripePaymentMethod.Card != nil {
		paymentMethod.Card = &Card{
			Last4:       stripePaymentMethod.Card.Last4,
			Brand:       string(stripePaymentMethod.Card.Brand),
			ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
			ExpYear:     int(stripePaymentMethod.Card.ExpYear),
			Fingerprint: stripePaymentMethod.Card.Fingerprint,
		}
	}

	return paymentMethod, nil
}

// GetPaymentMethod retrieves a payment method by ID
func (s *CustomerService) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error) {
	ctx, span := s.tracer.Start(ctx, "GetPaymentMethod")
	defer span.End()

	if paymentMethodID == "" {
		return nil, fmt.Errorf("payment method ID cannot be empty")
	}

	params := &stripe.PaymentMethodParams{}
	stripePaymentMethod, err := paymentmethod.Get(paymentMethodID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment method: %w", err)
	}

	// Convert to our PaymentMethod type
	paymentMethod := &PaymentMethod{
		ID:       stripePaymentMethod.ID,
		Type:     string(stripePaymentMethod.Type),
		Customer: stripePaymentMethod.Customer.ID,
		Metadata: stripePaymentMethod.Metadata,
		Created:  stripePaymentMethod.Created,
	}

	// Add card details if available
	if stripePaymentMethod.Card != nil {
		paymentMethod.Card = &Card{
			Last4:       stripePaymentMethod.Card.Last4,
			Brand:       string(stripePaymentMethod.Card.Brand),
			ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
			ExpYear:     int(stripePaymentMethod.Card.ExpYear),
			Fingerprint: stripePaymentMethod.Card.Fingerprint,
		}
	}

	return paymentMethod, nil
}

// ListPaymentMethods retrieves payment methods for a customer
func (s *CustomerService) ListPaymentMethods(ctx context.Context, customerID string, limit int64) ([]*PaymentMethod, error) {
	ctx, span := s.tracer.Start(ctx, "ListPaymentMethods")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID cannot be empty")
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	if limit > 0 {
		params.Limit = stripe.Int64(limit)
	}

	iter := paymentmethod.List(params)
	var paymentMethods []*PaymentMethod

	for iter.Next() {
		stripePaymentMethod := iter.PaymentMethod()
		paymentMethod := &PaymentMethod{
			ID:       stripePaymentMethod.ID,
			Type:     string(stripePaymentMethod.Type),
			Customer: stripePaymentMethod.Customer.ID,
			Metadata: stripePaymentMethod.Metadata,
			Created:  stripePaymentMethod.Created,
		}

		// Add card details if available
		if stripePaymentMethod.Card != nil {
			paymentMethod.Card = &Card{
				Last4:       stripePaymentMethod.Card.Last4,
				Brand:       string(stripePaymentMethod.Card.Brand),
				ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
				ExpYear:     int(stripePaymentMethod.Card.ExpYear),
				Fingerprint: stripePaymentMethod.Card.Fingerprint,
			}
		}

		paymentMethods = append(paymentMethods, paymentMethod)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	return paymentMethods, nil
}

// DetachPaymentMethod removes a payment method from a customer
func (s *CustomerService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	ctx, span := s.tracer.Start(ctx, "DetachPaymentMethod")
	defer span.End()

	if paymentMethodID == "" {
		return fmt.Errorf("payment method ID cannot be empty")
	}

	_, err := paymentmethod.Detach(paymentMethodID, &stripe.PaymentMethodDetachParams{})
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}

	return nil
}

// ValidateCustomerRequest validates a customer request
func (s *CustomerService) ValidateCustomerRequest(request *CustomerRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if request.Email == "" {
		return fmt.Errorf("email is required")
	}

	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}

// ValidatePaymentMethodRequest validates a payment method request
func (s *CustomerService) ValidatePaymentMethodRequest(request *PaymentMethodRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if request.Type == "" {
		return fmt.Errorf("type is required")
	}

	if request.Customer == "" {
		return fmt.Errorf("customer is required")
	}

	if request.Type == "card" && request.Card == nil {
		return fmt.Errorf("card details are required for card payment methods")
	}

	if request.Card != nil && request.Card.Token == "" {
		return fmt.Errorf("card token is required")
	}

	return nil
}

// GenerateCustomerID generates a unique customer ID for internal use
func (s *CustomerService) GenerateCustomerID() string {
	return fmt.Sprintf("cus_%s", uuid.New().String()[:8])
}
