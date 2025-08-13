package services

import (
	"context"
	"fmt"
	"log"

	"apis/payments/kafka"
)

// PaymentService provides a unified interface for payment operations
// using the configured payment gateway
type PaymentService struct {
	gateway        PaymentGateway
	factory        *GatewayFactory
	eventPublisher EventPublisher
}

// EventPublisher defines the interface for publishing payment events
type EventPublisher interface {
	Publish(topic string, event *kafka.PaymentEvent) error
}

// NewPaymentService creates a new payment service with the default gateway
func NewPaymentService() (*PaymentService, error) {
	factory, err := NewGatewayFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway factory: %w", err)
	}

	gateway, err := factory.CreateDefaultGateway()
	if err != nil {
		return nil, fmt.Errorf("failed to create default gateway: %w", err)
	}

	return &PaymentService{
		gateway: gateway,
		factory: factory,
	}, nil
}

// NewPaymentServiceWithProvider creates a new payment service with a specific provider
func NewPaymentServiceWithProvider(providerType ProviderType) (*PaymentService, error) {
	factory, err := NewGatewayFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway factory: %w", err)
	}

	gateway, err := factory.CreateGateway(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway for provider %s: %w", providerType, err)
	}

	return &PaymentService{
		gateway: gateway,
		factory: factory,
	}, nil
}

// NewPaymentServiceWithMockGateway creates a new payment service with a mock gateway for testing
func NewPaymentServiceWithMockGateway() *PaymentService {
	return &PaymentService{
		gateway: NewMockGateway(),
		factory: nil,
	}
}

// GetProviderName returns the name of the current payment provider
func (s *PaymentService) GetProviderName() string {
	return s.gateway.GetProviderName()
}

// GetCapabilities returns the capabilities of the current payment provider
func (s *PaymentService) GetCapabilities() GatewayCapabilities {
	return s.gateway.GetCapabilities()
}

// SwitchProvider switches to a different payment provider
func (s *PaymentService) SwitchProvider(providerType ProviderType) error {
	gateway, err := s.factory.CreateGateway(providerType)
	if err != nil {
		return fmt.Errorf("failed to switch to provider %s: %w", providerType, err)
	}

	s.gateway = gateway
	log.Printf("Switched to payment provider: %s", providerType)
	return nil
}

// GetAvailableProviders returns a list of available payment providers
func (s *PaymentService) GetAvailableProviders() []ProviderType {
	return s.factory.GetAvailableProviders()
}

// Customer operations
func (s *PaymentService) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
	customer, err := s.gateway.CreateCustomer(ctx, req)
	if err == nil && customer != nil {
		s.publishEvent("customer.created", "/payments/customers", map[string]interface{}{
			"customer_id": customer.ID,
			"email":       customer.Email,
			"name":        customer.Name,
		})
	}
	return customer, err
}

func (s *PaymentService) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	return s.gateway.GetCustomer(ctx, customerID)
}

func (s *PaymentService) UpdateCustomer(ctx context.Context, customerID string, req *CustomerRequest) (*Customer, error) {
	return s.gateway.UpdateCustomer(ctx, customerID, req)
}

func (s *PaymentService) DeleteCustomer(ctx context.Context, customerID string) error {
	return s.gateway.DeleteCustomer(ctx, customerID)
}

// Payment method operations
func (s *PaymentService) AddPaymentMethod(ctx context.Context, req *PaymentMethodRequest) (*PaymentMethod, error) {
	return s.gateway.AddPaymentMethod(ctx, req)
}

func (s *PaymentService) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error) {
	return s.gateway.GetPaymentMethod(ctx, paymentMethodID)
}

func (s *PaymentService) ListPaymentMethods(ctx context.Context, customerID string, limit int) ([]*PaymentMethod, error) {
	return s.gateway.ListPaymentMethods(ctx, customerID, limit)
}

func (s *PaymentService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return s.gateway.DetachPaymentMethod(ctx, paymentMethodID)
}

// Charge operations
func (s *PaymentService) CreateCharge(ctx context.Context, req *ChargeRequest) (*Charge, error) {
	charge, err := s.gateway.CreateCharge(ctx, req)
	if err == nil && charge != nil {
		s.publishEvent("charge.created", "/payments/charges", map[string]interface{}{
			"charge_id":   charge.ID,
			"amount":      charge.Amount,
			"currency":    charge.Currency,
			"status":      charge.Status,
			"customer_id": charge.CustomerID,
		})
	}
	return charge, err
}

func (s *PaymentService) GetCharge(ctx context.Context, chargeID string) (*Charge, error) {
	return s.gateway.GetCharge(ctx, chargeID)
}

func (s *PaymentService) ListCharges(ctx context.Context, customerID string, limit int) ([]*Charge, error) {
	return s.gateway.ListCharges(ctx, customerID, limit)
}

// Refund operations
func (s *PaymentService) CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error) {
	refund, err := s.gateway.CreateRefund(ctx, req)
	if err == nil && refund != nil {
		s.publishEvent("refund.created", "/payments/refunds", map[string]interface{}{
			"refund_id": refund.ID,
			"charge_id": refund.ChargeID,
			"amount":    refund.Amount,
			"currency":  refund.Currency,
			"status":    refund.Status,
			"reason":    refund.Reason,
		})
	}
	return refund, err
}

func (s *PaymentService) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	return s.gateway.GetRefund(ctx, refundID)
}

func (s *PaymentService) ListRefunds(ctx context.Context, chargeID string, limit int) ([]*Refund, error) {
	return s.gateway.ListRefunds(ctx, chargeID, limit)
}

// Dispute operations
func (s *PaymentService) CreateDispute(ctx context.Context, req *DisputeRequest) (*Dispute, error) {
	dispute, err := s.gateway.CreateDispute(ctx, req)
	if err == nil && dispute != nil {
		s.publishEvent("dispute.created", "/payments/disputes", map[string]interface{}{
			"dispute_id": dispute.ID,
			"charge_id":  dispute.ChargeID,
			"amount":     dispute.Amount,
			"currency":   dispute.Currency,
			"status":     dispute.Status,
			"reason":     dispute.Reason,
		})
	}
	return dispute, err
}

func (s *PaymentService) GetDispute(ctx context.Context, disputeID string) (*Dispute, error) {
	return s.gateway.GetDispute(ctx, disputeID)
}

func (s *PaymentService) ListDisputes(ctx context.Context, chargeID string, limit int) ([]*Dispute, error) {
	return s.gateway.ListDisputes(ctx, chargeID, limit)
}

func (s *PaymentService) UpdateDisputeStatus(ctx context.Context, disputeID string, status string) (*Dispute, error) {
	return s.gateway.UpdateDisputeStatus(ctx, disputeID, status)
}

// Subscription plan methods
func (s *PaymentService) CreateSubscriptionPlan(ctx context.Context, req *SubscriptionPlanRequest) (*SubscriptionPlan, error) {
	return s.gateway.CreateSubscriptionPlan(ctx, req)
}

func (s *PaymentService) GetSubscriptionPlan(ctx context.Context, planID string) (*SubscriptionPlan, error) {
	return s.gateway.GetSubscriptionPlan(ctx, planID)
}

func (s *PaymentService) ListSubscriptionPlans(ctx context.Context, params *SubscriptionPlanListParams) ([]*SubscriptionPlan, error) {
	return s.gateway.ListSubscriptionPlans(ctx, params)
}

func (s *PaymentService) UpdateSubscriptionPlan(ctx context.Context, planID string, req *SubscriptionPlanUpdateRequest) (*SubscriptionPlan, error) {
	return s.gateway.UpdateSubscriptionPlan(ctx, planID, req)
}

func (s *PaymentService) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	return s.gateway.DeleteSubscriptionPlan(ctx, planID)
}

// Subscription methods
func (s *PaymentService) CreateSubscription(ctx context.Context, req *SubscriptionRequest) (*Subscription, error) {
	return s.gateway.CreateSubscription(ctx, req)
}

func (s *PaymentService) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	return s.gateway.GetSubscription(ctx, subscriptionID)
}

func (s *PaymentService) ListSubscriptions(ctx context.Context, params *SubscriptionListParams) ([]*Subscription, error) {
	return s.gateway.ListSubscriptions(ctx, params)
}

func (s *PaymentService) UpdateSubscription(ctx context.Context, subscriptionID string, req *SubscriptionUpdateRequest) (*Subscription, error) {
	return s.gateway.UpdateSubscription(ctx, subscriptionID, req)
}

func (s *PaymentService) CancelSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	return s.gateway.CancelSubscription(ctx, subscriptionID)
}

func (s *PaymentService) ReactivateSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	return s.gateway.ReactivateSubscription(ctx, subscriptionID)
}

// Feature detection and capability checking
func (s *PaymentService) SupportsSubscriptions() bool {
	return s.gateway.GetCapabilities().SupportsSubscriptions
}

func (s *PaymentService) SupportsConnect() bool {
	return s.gateway.GetCapabilities().SupportsConnect
}

func (s *PaymentService) SupportsTax() bool {
	return s.gateway.GetCapabilities().SupportsTax
}

func (s *PaymentService) SupportsInvoices() bool {
	return s.gateway.GetCapabilities().SupportsInvoices
}

func (s *PaymentService) SupportsPayouts() bool {
	return s.gateway.GetCapabilities().SupportsPayouts
}

func (s *PaymentService) SupportsDisputes() bool {
	return s.gateway.GetCapabilities().SupportsDisputes
}

func (s *PaymentService) SupportsRefunds() bool {
	return s.gateway.GetCapabilities().SupportsRefunds
}

// GetSupportedCurrencies returns the currencies supported by the current provider
func (s *PaymentService) GetSupportedCurrencies() []string {
	return s.gateway.GetCapabilities().SupportedCurrencies
}

// GetSupportedCountries returns the countries supported by the current provider
func (s *PaymentService) GetSupportedCountries() []string {
	return s.gateway.GetCapabilities().SupportedCountries
}

// GetMaxPaymentAmount returns the maximum payment amount supported by the current provider
func (s *PaymentService) GetMaxPaymentAmount() int64 {
	return s.gateway.GetCapabilities().MaxPaymentAmount
}

// SetEventPublisher sets the event publisher for the payment service
func (s *PaymentService) SetEventPublisher(publisher EventPublisher) {
	s.eventPublisher = publisher
}

// publishEvent publishes a payment event if an event publisher is configured
func (s *PaymentService) publishEvent(eventType, source string, data map[string]interface{}) {
	if s.eventPublisher == nil {
		return
	}

	event := &kafka.PaymentEvent{
		Type:   eventType,
		Source: source,
		Data:   data,
	}

	if err := s.eventPublisher.Publish("payment-events", event); err != nil {
		log.Printf("Failed to publish event %s: %v", eventType, err)
	}
}

// ValidatePaymentRequest validates a payment request against provider capabilities
func (s *PaymentService) ValidatePaymentRequest(req *ChargeRequest) error {
	capabilities := s.gateway.GetCapabilities()

	// Check currency support
	currencySupported := false
	for _, currency := range capabilities.SupportedCurrencies {
		if currency == req.Currency {
			currencySupported = true
			break
		}
	}
	if !currencySupported {
		return fmt.Errorf("currency %s is not supported by provider %s", req.Currency, s.gateway.GetProviderName())
	}

	// Check amount limits
	if req.Amount > capabilities.MaxPaymentAmount {
		return fmt.Errorf("amount %d exceeds maximum allowed amount %d for provider %s",
			req.Amount, capabilities.MaxPaymentAmount, s.gateway.GetProviderName())
	}

	return nil
}
