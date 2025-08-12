package services

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v78"
)

// PaymentGateway defines the core interface for payment gateway operations
type PaymentGateway interface {
	// Provider information
	GetProvider() string
	GetCapabilities() GatewayCapabilities
	
	// Customer management
	CustomerVault
	// Payment processing
	PaymentProcessor
	// Refund handling
	RefundProcessor
	// Subscription management (if supported)
	SubscriptionManager
}

// GatewayCapabilities defines what features a payment gateway supports
type GatewayCapabilities struct {
	SupportsCustomers     bool
	SupportsCharges       bool
	SupportsRefunds       bool
	SupportsSubscriptions bool
	SupportsDisputes      bool
	SupportsConnect       bool
	SupportsTax           bool
	MaxChargeAmount       int64  // in cents
	MinChargeAmount       int64  // in cents
	SupportedCurrencies   []string
	SupportedCountries    []string
}

// CustomerVault defines customer management operations
type CustomerVault interface {
	// CreateCustomer creates a new customer
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error)
	
	// GetCustomer retrieves a customer by ID
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
	
	// UpdateCustomer updates an existing customer
	UpdateCustomer(ctx context.Context, customerID string, req UpdateCustomerRequest) (*Customer, error)
	
	// DeleteCustomer deletes a customer
	DeleteCustomer(ctx context.Context, customerID string) error
	
	// ListCustomers lists customers with optional filtering
	ListCustomers(ctx context.Context, req ListCustomersRequest) (*CustomerList, error)
	
	// AddPaymentMethod adds a payment method to a customer
	AddPaymentMethod(ctx context.Context, customerID string, req AddPaymentMethodRequest) (*PaymentMethod, error)
	
	// RemovePaymentMethod removes a payment method from a customer
	RemovePaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error
	
	// ListPaymentMethods lists payment methods for a customer
	ListPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethod, error)
}

// PaymentProcessor defines payment processing operations
type PaymentProcessor interface {
	// CreateCharge creates a new charge
	CreateCharge(ctx context.Context, req CreateChargeRequest) (*Charge, error)
	
	// GetCharge retrieves a charge by ID
	GetCharge(ctx context.Context, chargeID string) (*Charge, error)
	
	// UpdateCharge updates an existing charge
	UpdateCharge(ctx context.Context, chargeID string, req UpdateChargeRequest) (*Charge, error)
	
	// CaptureCharge captures a previously authorized charge
	CaptureCharge(ctx context.Context, chargeID string, req CaptureChargeRequest) (*Charge, error)
	
	// ListCharges lists charges with optional filtering
	ListCharges(ctx context.Context, req ListChargesRequest) (*ChargeList, error)
}

// RefundProcessor defines refund processing operations
type RefundProcessor interface {
	// CreateRefund creates a new refund
	CreateRefund(ctx context.Context, req CreateRefundRequest) (*Refund, error)
	
	// GetRefund retrieves a refund by ID
	GetRefund(ctx context.Context, refundID string) (*Refund, error)
	
	// UpdateRefund updates an existing refund
	UpdateRefund(ctx context.Context, refundID string, req UpdateRefundRequest) (*Refund, error)
	
	// ListRefunds lists refunds with optional filtering
	ListRefunds(ctx context.Context, req ListRefundsRequest) (*RefundList, error)
}

// SubscriptionManager defines subscription management operations (optional)
type SubscriptionManager interface {
	// CreateSubscription creates a new subscription
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*Subscription, error)
	
	// GetSubscription retrieves a subscription by ID
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	
	// UpdateSubscription updates an existing subscription
	UpdateSubscription(ctx context.Context, subscriptionID string, req UpdateSubscriptionRequest) (*Subscription, error)
	
	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, subscriptionID string, req CancelSubscriptionRequest) (*Subscription, error)
	
	// ListSubscriptions lists subscriptions with optional filtering
	ListSubscriptions(ctx context.Context, req ListSubscriptionsRequest) (*SubscriptionList, error)
}

// Common data structures

// Customer represents a customer in the payment system
type Customer struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	Phone         string                 `json:"phone,omitempty"`
	Address       *Address               `json:"address,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ProviderID    string                 `json:"provider_id"` // Original provider's customer ID
	Provider      string                 `json:"provider"`    // Which provider this customer belongs to
}

// Address represents a customer's address
type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// PaymentMethod represents a payment method (card, bank account, etc.)
type PaymentMethod struct {
	ID           string                 `json:"id"`
	CustomerID   string                 `json:"customer_id"`
	Type         string                 `json:"type"` // card, bank_account, etc.
	Card         *Card                  `json:"card,omitempty"`
	BankAccount  *BankAccount           `json:"bank_account,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ProviderID   string                 `json:"provider_id"`
	Provider     string                 `json:"provider"`
}

// Card represents a credit/debit card
type Card struct {
	Brand       string `json:"brand"`
	Last4       string `json:"last4"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Country     string `json:"country,omitempty"`
}

// BankAccount represents a bank account
type BankAccount struct {
	BankName    string `json:"bank_name"`
	Last4       string `json:"last4"`
	RoutingNumber string `json:"routing_number,omitempty"`
	AccountType string `json:"account_type,omitempty"`
	Country     string `json:"country,omitempty"`
}

// Charge represents a payment charge
type Charge struct {
	ID              string                 `json:"id"`
	Amount          int64                  `json:"amount"` // in cents
	Currency        string                 `json:"currency"`
	CustomerID      string                 `json:"customer_id"`
	PaymentMethodID string                 `json:"payment_method_id"`
	Status          string                 `json:"status"`
	Description     string                 `json:"description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ProviderID      string                 `json:"provider_id"`
	Provider        string                 `json:"provider"`
}

// Refund represents a refund
type Refund struct {
	ID          string                 `json:"id"`
	ChargeID    string                 `json:"charge_id"`
	Amount      int64                  `json:"amount"` // in cents
	Currency    string                 `json:"currency"`
	Reason      string                 `json:"reason,omitempty"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ProviderID  string                 `json:"provider_id"`
	Provider    string                 `json:"provider"`
}

// Subscription represents a subscription
type Subscription struct {
	ID           string                 `json:"id"`
	CustomerID   string                 `json:"customer_id"`
	PlanID       string                 `json:"plan_id"`
	Status       string                 `json:"status"`
	CurrentPeriodStart time.Time         `json:"current_period_start"`
	CurrentPeriodEnd   time.Time         `json:"current_period_end"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ProviderID   string                 `json:"provider_id"`
	Provider     string                 `json:"provider"`
}

// Request/Response structures

type CreateCustomerRequest struct {
	Email    string                 `json:"email"`
	Name     string                 `json:"name,omitempty"`
	Phone    string                 `json:"phone,omitempty"`
	Address  *Address               `json:"address,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateCustomerRequest struct {
	Email    string                 `json:"email,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Phone    string                 `json:"phone,omitempty"`
	Address  *Address               `json:"address,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ListCustomersRequest struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Email  string `json:"email,omitempty"`
}

type CustomerList struct {
	Customers []*Customer `json:"customers"`
	Total     int         `json:"total"`
	HasMore   bool        `json:"has_more"`
}

type AddPaymentMethodRequest struct {
	Type           string                 `json:"type"`
	Card           *Card                  `json:"card,omitempty"`
	BankAccount    *BankAccount           `json:"bank_account,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type CreateChargeRequest struct {
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	CustomerID      string                 `json:"customer_id"`
	PaymentMethodID string                 `json:"payment_method_id"`
	Description     string                 `json:"description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Capture         bool                   `json:"capture"` // true for immediate capture, false for authorization only
}

type UpdateChargeRequest struct {
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CaptureChargeRequest struct {
	Amount int64 `json:"amount,omitempty"` // if not provided, captures the full amount
}

type ListChargesRequest struct {
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

type ChargeList struct {
	Charges []*Charge `json:"charges"`
	Total   int       `json:"total"`
	HasMore bool      `json:"has_more"`
}

type CreateRefundRequest struct {
	ChargeID string `json:"charge_id"`
	Amount   int64  `json:"amount,omitempty"` // if not provided, refunds the full amount
	Reason   string `json:"reason,omitempty"`
}

type UpdateRefundRequest struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ListRefundsRequest struct {
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	ChargeID string `json:"charge_id,omitempty"`
}

type RefundList struct {
	Refunds []*Refund `json:"refunds"`
	Total   int       `json:"total"`
	HasMore bool      `json:"has_more"`
}

type CreateSubscriptionRequest struct {
	CustomerID string                 `json:"customer_id"`
	PlanID     string                 `json:"plan_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateSubscriptionRequest struct {
	PlanID   string                 `json:"plan_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CancelSubscriptionRequest struct {
	AtPeriodEnd bool `json:"at_period_end"` // true to cancel at period end, false for immediate cancellation
}

type ListSubscriptionsRequest struct {
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

type SubscriptionList struct {
	Subscriptions []*Subscription `json:"subscriptions"`
	Total         int             `json:"total"`
	HasMore       bool            `json:"has_more"`
}

// ProviderFactory creates payment gateway instances
type ProviderFactory interface {
	// CreateGateway creates a new payment gateway instance
	CreateGateway(provider string, config map[string]interface{}) (PaymentGateway, error)
	
	// GetSupportedProviders returns a list of supported providers
	GetSupportedProviders() []string
	
	// ValidateConfig validates provider configuration
	ValidateConfig(provider string, config map[string]interface{}) error
}

// DefaultProviderFactory implements the ProviderFactory interface
type DefaultProviderFactory struct {
	providers map[string]func(map[string]interface{}) (PaymentGateway, error)
}

// NewDefaultProviderFactory creates a new default provider factory
func NewDefaultProviderFactory() *DefaultProviderFactory {
	return &DefaultProviderFactory{
		providers: make(map[string]func(map[string]interface{}) (PaymentGateway, error)),
	}
}

// RegisterProvider registers a provider creation function
func (f *DefaultProviderFactory) RegisterProvider(name string, creator func(map[string]interface{}) (PaymentGateway, error)) {
	f.providers[name] = creator
}

// CreateGateway creates a new payment gateway instance
func (f *DefaultProviderFactory) CreateGateway(provider string, config map[string]interface{}) (PaymentGateway, error) {
	creator, exists := f.providers[provider]
	if !exists {
		return nil, &UnsupportedProviderError{Provider: provider}
	}
	return nil, nil
}

// GetSupportedProviders returns a list of supported providers
func (f *DefaultProviderFactory) GetSupportedProviders() []string {
	providers := make([]string, 0, len(f.providers))
	for provider := range f.providers {
		providers = append(providers, provider)
	}
	return providers
}

// ValidateConfig validates provider configuration
func (f *DefaultProviderFactory) ValidateConfig(provider string, config map[string]interface{}) error {
	// Basic validation - each provider should implement its own validation
	if config == nil {
		return &InvalidConfigError{Message: "config cannot be nil"}
	}
	return nil
}

// Error types

type UnsupportedProviderError struct {
	Provider string
}

func (e *UnsupportedProviderError) Error() string {
	return "unsupported payment provider: " + e.Provider
}

type InvalidConfigError struct {
	Message string
}

func (e *InvalidConfigError) Error() string {
	return "invalid configuration: " + e.Message
}

type PaymentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Provider string `json:"provider"`
}

func (e *PaymentError) Error() string {
	return e.Message
}