package services

import (
	"context"
	"fmt"
	"time"
)

// MockGateway is a mock implementation of PaymentGateway for testing
type MockGateway struct {
	customers      map[string]*Customer
	paymentMethods map[string]*PaymentMethod
	charges        map[string]*Charge
	refunds        map[string]*Refund
	disputes       map[string]*Dispute
}

// NewMockGateway creates a new mock gateway for testing
func NewMockGateway() *MockGateway {
	return &MockGateway{
		customers:      make(map[string]*Customer),
		paymentMethods: make(map[string]*PaymentMethod),
		charges:        make(map[string]*Charge),
		refunds:        make(map[string]*Refund),
		disputes:       make(map[string]*Dispute),
	}
}

// Customer operations
func (m *MockGateway) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
	customer := &Customer{
		ID:          "cus_mock_" + time.Now().Format("20060102150405"),
		Email:       req.Email,
		Name:        req.Name,
		Phone:       req.Phone,
		Description: req.Description,
		Metadata:    req.Metadata,
		Created:     time.Now().Unix(),
		Updated:     time.Now().Unix(),
		ProviderID:  "cus_mock",
	}

	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *MockGateway) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	if customer, exists := m.customers[customerID]; exists {
		return customer, nil
	}
	return nil, fmt.Errorf("customer not found")
}

func (m *MockGateway) UpdateCustomer(ctx context.Context, customerID string, req *CustomerRequest) (*Customer, error) {
	if customer, exists := m.customers[customerID]; exists {
		customer.Email = req.Email
		customer.Name = req.Name
		customer.Phone = req.Phone
		customer.Description = req.Description
		customer.Metadata = req.Metadata
		customer.Updated = time.Now().Unix()
		return customer, nil
	}
	return nil, fmt.Errorf("customer not found")
}

func (m *MockGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	if _, exists := m.customers[customerID]; exists {
		delete(m.customers, customerID)
		return nil
	}
	return fmt.Errorf("customer not found")
}

// Payment method operations
func (m *MockGateway) AddPaymentMethod(ctx context.Context, req *PaymentMethodRequest) (*PaymentMethod, error) {
	paymentMethod := &PaymentMethod{
		ID:         "pm_mock_" + time.Now().Format("20060102150405"),
		Type:       req.Type,
		Customer:   req.Customer,
		Metadata:   req.Metadata,
		Created:    time.Now().Unix(),
		ProviderID: "pm_mock",
	}

	if req.Card != nil {
		paymentMethod.Card = &Card{
			Last4:       "1234",
			Brand:       "visa",
			ExpMonth:    12,
			ExpYear:     2025,
			Fingerprint: "mock_fingerprint",
		}
	}

	m.paymentMethods[paymentMethod.ID] = paymentMethod
	return paymentMethod, nil
}

func (m *MockGateway) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error) {
	if pm, exists := m.paymentMethods[paymentMethodID]; exists {
		return pm, nil
	}
	return nil, fmt.Errorf("payment method not found")
}

func (m *MockGateway) ListPaymentMethods(ctx context.Context, customerID string, limit int) ([]*PaymentMethod, error) {
	var methods []*PaymentMethod
	for _, pm := range m.paymentMethods {
		if pm.Customer == customerID {
			methods = append(methods, pm)
		}
	}
	return methods, nil
}

func (m *MockGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	if _, exists := m.paymentMethods[paymentMethodID]; exists {
		delete(m.paymentMethods, paymentMethodID)
		return nil
	}
	return fmt.Errorf("payment method not found")
}

// Charge operations
func (m *MockGateway) CreateCharge(ctx context.Context, req *ChargeRequest) (*Charge, error) {
	charge := &Charge{
		ID:              "ch_mock_" + time.Now().Format("20060102150405"),
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          "succeeded",
		CustomerID:      req.CustomerID,
		PaymentMethodID: req.PaymentMethod,
		Description:     req.Description,
		Metadata:        req.Metadata,
		Created:         time.Now().Unix(),
		Updated:         time.Now().Unix(),
		ProviderID:      "ch_mock",
	}

	m.charges[charge.ID] = charge
	return charge, nil
}

func (m *MockGateway) GetCharge(ctx context.Context, chargeID string) (*Charge, error) {
	if charge, exists := m.charges[chargeID]; exists {
		return charge, nil
	}
	return nil, fmt.Errorf("charge not found")
}

func (m *MockGateway) ListCharges(ctx context.Context, customerID string, limit int) ([]*Charge, error) {
	var charges []*Charge
	for _, charge := range m.charges {
		if charge.CustomerID == customerID {
			charges = append(charges, charge)
		}
	}
	return charges, nil
}

// Refund operations
func (m *MockGateway) CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error) {
	refund := &Refund{
		ID:         "re_mock_" + time.Now().Format("20060102150405"),
		ChargeID:   req.ChargeID,
		Amount:     req.Amount,
		Currency:   "usd", // Default currency
		Status:     "succeeded",
		Reason:     req.Reason,
		Metadata:   req.Metadata,
		Created:    time.Now().Unix(),
		Updated:    time.Now().Unix(),
		ProviderID: "re_mock",
	}

	m.refunds[refund.ID] = refund
	return refund, nil
}

func (m *MockGateway) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	if refund, exists := m.refunds[refundID]; exists {
		return refund, nil
	}
	return nil, fmt.Errorf("refund not found")
}

func (m *MockGateway) ListRefunds(ctx context.Context, chargeID string, limit int) ([]*Refund, error) {
	var refunds []*Refund
	for _, refund := range m.refunds {
		if refund.ChargeID == chargeID {
			refunds = append(refunds, refund)
		}
	}
	return refunds, nil
}

// Dispute operations
func (m *MockGateway) CreateDispute(ctx context.Context, req *DisputeRequest) (*Dispute, error) {
	dispute := &Dispute{
		ID:         "dp_mock_" + time.Now().Format("20060102150405"),
		ChargeID:   req.ChargeID,
		Amount:     req.Amount,
		Currency:   "usd", // Default currency
		Status:     "open",
		Reason:     req.Reason,
		Evidence:   req.Evidence,
		Metadata:   req.Metadata,
		Created:    time.Now().Unix(),
		Updated:    time.Now().Unix(),
		ProviderID: "dp_mock",
	}

	m.disputes[dispute.ID] = dispute
	return dispute, nil
}

func (m *MockGateway) GetDispute(ctx context.Context, disputeID string) (*Dispute, error) {
	if dispute, exists := m.disputes[disputeID]; exists {
		return dispute, nil
	}
	return nil, fmt.Errorf("dispute not found")
}

func (m *MockGateway) ListDisputes(ctx context.Context, chargeID string, limit int) ([]*Dispute, error) {
	var disputes []*Dispute
	for _, dispute := range m.disputes {
		if dispute.ChargeID == chargeID {
			disputes = append(disputes, dispute)
		}
	}
	return disputes, nil
}

func (m *MockGateway) UpdateDisputeStatus(ctx context.Context, disputeID string, status string) (*Dispute, error) {
	if dispute, exists := m.disputes[disputeID]; exists {
		dispute.Status = status
		dispute.Updated = time.Now().Unix()
		return dispute, nil
	}
	return nil, fmt.Errorf("dispute not found")
}

// Provider information
func (m *MockGateway) GetProviderName() string {
	return "mock"
}

func (m *MockGateway) GetCapabilities() GatewayCapabilities {
	return GatewayCapabilities{
		SupportsSubscriptions: true,
		SupportsConnect:       true,
		SupportsTax:           true,
		SupportsInvoices:      true,
		SupportsPayouts:       true,
		SupportsDisputes:      true,
		SupportsRefunds:       true,
		MaxPaymentAmount:      1000000, // $10,000
		SupportedCurrencies:   []string{"usd", "eur", "gbp"},
		SupportedCountries:    []string{"US", "CA", "GB", "DE"},
	}
}
