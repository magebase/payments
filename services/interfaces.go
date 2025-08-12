package services

import (
	"context"
)

// PaymentGateway defines the interface for all payment operations
// This allows the service to switch between different payment providers
// without changing the core business logic
type PaymentGateway interface {
	// Customer operations
	CustomerGateway
	// Payment method operations
	PaymentMethodGateway
	// Charge operations
	ChargeGateway
	// Refund operations
	RefundGateway
	// Dispute operations
	DisputeGateway
	// Provider information
	GetProviderName() string
	GetCapabilities() GatewayCapabilities
}

// CustomerGateway defines customer-related operations
type CustomerGateway interface {
	CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error)
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, req *CustomerRequest) (*Customer, error)
	DeleteCustomer(ctx context.Context, customerID string) error
}

// PaymentMethodGateway defines payment method operations
type PaymentMethodGateway interface {
	AddPaymentMethod(ctx context.Context, req *PaymentMethodRequest) (*PaymentMethod, error)
	GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error)
	ListPaymentMethods(ctx context.Context, customerID string, limit int) ([]*PaymentMethod, error)
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
}

// ChargeGateway defines charge operations
type ChargeGateway interface {
	CreateCharge(ctx context.Context, req *ChargeRequest) (*Charge, error)
	GetCharge(ctx context.Context, chargeID string) (*Charge, error)
	ListCharges(ctx context.Context, customerID string, limit int) ([]*Charge, error)
}

// RefundGateway defines refund operations
type RefundGateway interface {
	CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error)
	GetRefund(ctx context.Context, refundID string) (*Refund, error)
	ListRefunds(ctx context.Context, chargeID string, limit int) ([]*Refund, error)
}

// DisputeGateway defines dispute operations
type DisputeGateway interface {
	CreateDispute(ctx context.Context, req *DisputeRequest) (*Dispute, error)
	GetDispute(ctx context.Context, disputeID string) (*Dispute, error)
	ListDisputes(ctx context.Context, chargeID string, limit int) ([]*Dispute, error)
	UpdateDisputeStatus(ctx context.Context, disputeID string, status string) (*Dispute, error)
}

// GatewayCapabilities defines what features a payment gateway supports
type GatewayCapabilities struct {
	SupportsSubscriptions bool
	SupportsConnect       bool
	SupportsTax           bool
	SupportsInvoices      bool
	SupportsPayouts       bool
	SupportsDisputes      bool
	SupportsRefunds       bool
	MaxPaymentAmount      int64
	SupportedCurrencies   []string
	SupportedCountries    []string
}

// Common request and response types that work across all providers
type CustomerRequest struct {
	Email       string            `json:"email" validate:"required,email"`
	Name        string            `json:"name" validate:"required,min=1"`
	Phone       string            `json:"phone,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Customer struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Phone       string            `json:"phone,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Created     int64             `json:"created"`
	Updated     int64             `json:"updated"`
	ProviderID  string            `json:"provider_id"` // Original provider ID
}

type PaymentMethodRequest struct {
	Type     string            `json:"type" validate:"required,oneof=card sepa_debit ideal sofort"`
	Card     *CardRequest      `json:"card,omitempty"`
	Customer string            `json:"customer" validate:"required"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CardRequest struct {
	Token string `json:"token" validate:"required"`
}

type PaymentMethod struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Card       *Card             `json:"card,omitempty"`
	Customer   string            `json:"customer"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Created    int64             `json:"created"`
	ProviderID string            `json:"provider_id"` // Original provider ID
}

type Card struct {
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint"`
}

type ChargeRequest struct {
	Amount        int64             `json:"amount" validate:"required,min=1"`
	Currency      string            `json:"currency" validate:"required,len=3"`
	CustomerID    string            `json:"customer_id" validate:"required"`
	PaymentMethod string            `json:"payment_method,omitempty"`
	Description   string            `json:"description,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type Charge struct {
	ID              string            `json:"id"`
	Amount          int64             `json:"amount"`
	Currency        string            `json:"currency"`
	Status          string            `json:"status"`
	CustomerID      string            `json:"customer_id"`
	PaymentMethodID string            `json:"payment_method_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Created         int64             `json:"created"`
	Updated         int64             `json:"updated"`
	ProviderID      string            `json:"provider_id"` // Original provider ID
}

type RefundRequest struct {
	ChargeID string            `json:"charge_id" validate:"required"`
	Amount   int64             `json:"amount,omitempty"`
	Reason   string            `json:"reason,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Refund struct {
	ID         string            `json:"id"`
	ChargeID   string            `json:"charge_id"`
	Amount     int64             `json:"amount"`
	Currency   string            `json:"currency"`
	Status     string            `json:"status"`
	Reason     string            `json:"reason,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Created    int64             `json:"created"`
	Updated    int64             `json:"updated"`
	ProviderID string            `json:"provider_id"` // Original provider ID
}

type DisputeRequest struct {
	ChargeID string            `json:"charge_id" validate:"required"`
	Amount   int64             `json:"amount" validate:"required,min=1"`
	Reason   string            `json:"reason" validate:"required"`
	Evidence map[string]string `json:"evidence,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Dispute struct {
	ID         string            `json:"id"`
	ChargeID   string            `json:"charge_id"`
	Amount     int64             `json:"amount"`
	Currency   string            `json:"currency"`
	Status     string            `json:"status"`
	Reason     string            `json:"reason"`
	Evidence   map[string]string `json:"evidence,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Created    int64             `json:"created"`
	Updated    int64             `json:"updated"`
	ProviderID string            `json:"provider_id"` // Original provider ID
}
