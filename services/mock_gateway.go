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
	subscriptionPlans map[string]*SubscriptionPlan
	subscriptions map[string]*Subscription
	counter       int64 // Add counter for unique IDs
}

// NewMockGateway creates a new mock gateway for testing
func NewMockGateway() *MockGateway {
	return &MockGateway{
		customers:      make(map[string]*Customer),
		paymentMethods: make(map[string]*PaymentMethod),
		charges:        make(map[string]*Charge),
		refunds:        make(map[string]*Refund),
		disputes:       make(map[string]*Dispute),
		subscriptionPlans: make(map[string]*SubscriptionPlan),
		subscriptions: make(map[string]*Subscription),
		counter:       0, // Initialize counter
	}
}

// Customer operations
func (m *MockGateway) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
	m.counter++
	customer := &Customer{
		ID:          "cus_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
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
	m.counter++
	paymentMethod := &PaymentMethod{
		ID:         "pm_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
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
	m.counter++
	charge := &Charge{
		ID:              "ch_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
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
	m.counter++
	refund := &Refund{
		ID:         "re_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
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
	m.counter++
	dispute := &Dispute{
		ID:         "dp_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
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

// Subscription plan operations
func (m *MockGateway) CreateSubscriptionPlan(ctx context.Context, req *SubscriptionPlanRequest) (*SubscriptionPlan, error) {
	m.counter++
	plan := &SubscriptionPlan{
		ID:             "sp_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
		Name:           req.Name,
		Description:    req.Description,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Interval:       req.Interval,
		IntervalCount:  req.IntervalCount,
		TrialPeriodDays: req.TrialPeriodDays,
		Metadata:       req.Metadata,
		Created:        time.Now().Unix(),
		Updated:        time.Now().Unix(),
		ProviderID:     "sp_mock",
	}

	m.subscriptionPlans[plan.ID] = plan
	
	return plan, nil
}

func (m *MockGateway) GetSubscriptionPlan(ctx context.Context, planID string) (*SubscriptionPlan, error) {
	if plan, exists := m.subscriptionPlans[planID]; exists {
		return plan, nil
	}
	return nil, fmt.Errorf("subscription plan not found")
}

func (m *MockGateway) ListSubscriptionPlans(ctx context.Context, params *SubscriptionPlanListParams) ([]*SubscriptionPlan, error) {
	var plans []*SubscriptionPlan
	for _, plan := range m.subscriptionPlans {
		plans = append(plans, plan)
	}
	
	// Apply limit and offset if specified
	if params.Limit > 0 && len(plans) > params.Limit {
		if params.Offset >= len(plans) {
			return []*SubscriptionPlan{}, nil
		}
		end := params.Offset + params.Limit
		if end > len(plans) {
			end = len(plans)
		}
		plans = plans[params.Offset:end]
	} else if params.Offset > 0 && params.Offset < len(plans) {
		plans = plans[params.Offset:]
	}
	
	return plans, nil
}

func (m *MockGateway) UpdateSubscriptionPlan(ctx context.Context, planID string, req *SubscriptionPlanUpdateRequest) (*SubscriptionPlan, error) {
	if plan, exists := m.subscriptionPlans[planID]; exists {
		// Store the original updated timestamp for comparison
		originalUpdated := plan.Updated
		
		if req.Name != nil {
			plan.Name = *req.Name
		}
		if req.Description != nil {
			plan.Description = *req.Description
		}
		if req.Amount != nil {
			plan.Amount = *req.Amount
		}
		if req.TrialPeriodDays != nil {
			plan.TrialPeriodDays = req.TrialPeriodDays
		}
		if req.Metadata != nil {
			plan.Metadata = req.Metadata
		}
		// Ensure update timestamp is newer than the original updated timestamp
		plan.Updated = originalUpdated + 1
		
		return plan, nil
	}
	return nil, fmt.Errorf("subscription plan not found")
}

func (m *MockGateway) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	if _, exists := m.subscriptionPlans[planID]; exists {
		delete(m.subscriptionPlans, planID)
		return nil
	}
	return fmt.Errorf("subscription plan not found")
}

// Subscription operations
func (m *MockGateway) CreateSubscription(ctx context.Context, req *SubscriptionRequest) (*Subscription, error) {
	m.counter++
	subscription := &Subscription{
		ID:                 "sub_mock_" + time.Now().Format("20060102150405") + fmt.Sprintf("%d", m.counter),
		CustomerID:         req.CustomerID,
		PlanID:             req.PlanID,
		Status:             "active",
		CurrentPeriodStart: time.Now().Unix(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0).Unix(), // 1 month from now
		TrialStart:         req.TrialEnd,
		TrialEnd:           req.TrialEnd,
		Metadata:           req.Metadata,
		Created:            time.Now().Unix(),
		Updated:            time.Now().Unix(),
		ProviderID:         "sub_mock",
	}

	m.subscriptions[subscription.ID] = subscription
	return subscription, nil
}

func (m *MockGateway) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

func (m *MockGateway) ListSubscriptions(ctx context.Context, params *SubscriptionListParams) ([]*Subscription, error) {
	var subscriptions []*Subscription
	for _, subscription := range m.subscriptions {
		if params.CustomerID != "" && subscription.CustomerID != params.CustomerID {
			continue
		}
		if params.Status != "" && subscription.Status != params.Status {
			continue
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (m *MockGateway) UpdateSubscription(ctx context.Context, subscriptionID string, req *SubscriptionUpdateRequest) (*Subscription, error) {
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		if req.PlanID != nil {
			subscription.PlanID = *req.PlanID
		}
		if req.PaymentMethod != nil {
			// Mock implementation doesn't store payment method
		}
		if req.TrialEnd != nil {
			subscription.TrialEnd = req.TrialEnd
		}
		if req.Metadata != nil {
			subscription.Metadata = req.Metadata
		}
		subscription.Updated = time.Now().Unix()
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

func (m *MockGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		subscription.Status = "canceled"
		subscription.CanceledAt = &[]int64{time.Now().Unix()}[0]
		subscription.Updated = time.Now().Unix()
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

func (m *MockGateway) ReactivateSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		subscription.Status = "active"
		subscription.CanceledAt = nil
		subscription.Updated = time.Now().Unix()
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
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
