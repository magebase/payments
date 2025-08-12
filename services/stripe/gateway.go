package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/magebase/payments/services"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/charge"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/paymentmethod"
	"github.com/stripe/stripe-go/v78/refund"
	"github.com/stripe/stripe-go/v78/subscription"
)

// StripeGateway implements the PaymentGateway interface for Stripe
type StripeGateway struct {
	apiKey string
	config map[string]interface{}
}

// NewStripeGateway creates a new Stripe payment gateway instance
func NewStripeGateway(config map[string]interface{}) (*StripeGateway, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, &services.InvalidConfigError{Message: "stripe api_key is required"}
	}

	// Set the Stripe API key
	stripe.Key = apiKey

	return &StripeGateway{
		apiKey: apiKey,
		config: config,
	}, nil
}

// GetProvider returns the provider name
func (g *StripeGateway) GetProvider() string {
	return "stripe"
}

// GetCapabilities returns the capabilities supported by Stripe
func (g *StripeGateway) GetCapabilities() services.GatewayCapabilities {
	return services.GatewayCapabilities{
		SupportsCustomers:     true,
		SupportsCharges:       true,
		SupportsRefunds:       true,
		SupportsSubscriptions: true,
		SupportsDisputes:      true,
		SupportsConnect:       true,
		SupportsTax:           true,
		MaxChargeAmount:       99999999, // $999,999.99 in cents
		MinChargeAmount:       50,       // $0.50 in cents
		SupportedCurrencies:   []string{"usd", "eur", "gbp", "cad", "aud", "jpy"},
		SupportedCountries:    []string{"US", "CA", "GB", "DE", "FR", "AU", "JP"},
	}
}

// Customer management implementation

func (g *StripeGateway) CreateCustomer(ctx context.Context, req services.CreateCustomerRequest) (*services.Customer, error) {
	// Convert to Stripe customer params
	params := &stripe.CustomerParams{
		Email:    stripe.String(req.Email),
		Name:     stripe.String(req.Name),
		Phone:    stripe.String(req.Phone),
		Metadata: req.Metadata,
	}

	// Add address if provided
	if req.Address != nil {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(req.Address.Line1),
			Line2:      stripe.String(req.Address.Line2),
			City:       stripe.String(req.Address.City),
			State:      stripe.String(req.Address.State),
			PostalCode: stripe.String(req.Address.PostalCode),
			Country:    stripe.String(req.Address.Country),
		}
	}

	// Create customer in Stripe
	stripeCustomer, err := customer.New(params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "customer_creation_failed",
			Message:  fmt.Sprintf("failed to create customer: %v", err),
			Provider: "stripe",
		}
	}

	// Convert back to our interface
	return g.convertStripeCustomer(stripeCustomer), nil
}

func (g *StripeGateway) GetCustomer(ctx context.Context, customerID string) (*services.Customer, error) {
	stripeCustomer, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "customer_retrieval_failed",
			Message:  fmt.Sprintf("failed to retrieve customer: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCustomer(stripeCustomer), nil
}

func (g *StripeGateway) UpdateCustomer(ctx context.Context, customerID string, req services.UpdateCustomerRequest) (*services.Customer, error) {
	params := &stripe.CustomerParams{}

	if req.Email != "" {
		params.Email = stripe.String(req.Email)
	}
	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}
	if req.Phone != "" {
		params.Phone = stripe.String(req.Phone)
	}
	if req.Address != nil {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(req.Address.Line1),
			Line2:      stripe.String(req.Address.Line2),
			City:       stripe.String(req.Address.City),
			State:      stripe.String(req.Address.State),
			PostalCode: stripe.String(req.Address.PostalCode),
			Country:    stripe.String(req.Address.Country),
		}
	}
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	stripeCustomer, err := customer.Update(customerID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "customer_update_failed",
			Message:  fmt.Sprintf("failed to update customer: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCustomer(stripeCustomer), nil
}

func (g *StripeGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID, nil)
	if err != nil {
		return &services.PaymentError{
			Code:     "customer_deletion_failed",
			Message:  fmt.Sprintf("failed to delete customer: %v", err),
			Provider: "stripe",
		}
	}
	return nil
}

func (g *StripeGateway) ListCustomers(ctx context.Context, req services.ListCustomersRequest) (*services.CustomerList, error) {
	params := &stripe.CustomerListParams{
		Limit: stripe.Int64(int64(req.Limit)),
	}

	if req.Email != "" {
		params.Filters.AddFilter("email", "", req.Email)
	}

	iter := customer.List(params)
	var customers []*services.Customer
	total := 0

	for iter.Next() {
		customers = append(customers, g.convertStripeCustomer(iter.Customer()))
		total++
	}

	if err := iter.Err(); err != nil {
		return nil, &services.PaymentError{
			Code:     "customer_list_failed",
			Message:  fmt.Sprintf("failed to list customers: %v", err),
			Provider: "stripe",
		}
	}

	return &services.CustomerList{
		Customers: customers,
		Total:     total,
		HasMore:   iter.Meta().HasMore,
	}, nil
}

func (g *StripeGateway) AddPaymentMethod(ctx context.Context, customerID string, req services.AddPaymentMethodRequest) (*services.PaymentMethod, error) {
	// Create payment method in Stripe
	pmParams := &stripe.PaymentMethodParams{
		Type: stripe.String(req.Type),
		Customer: stripe.String(customerID),
		Metadata: req.Metadata,
	}

	if req.Card != nil {
		pmParams.Card = &stripe.PaymentMethodCardParams{
			Token: stripe.String("tok_visa"), // In real implementation, you'd get this from frontend
		}
	}

	stripePM, err := paymentmethod.New(pmParams)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "payment_method_creation_failed",
			Message:  fmt.Sprintf("failed to create payment method: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripePaymentMethod(stripePM), nil
}

func (g *StripeGateway) RemovePaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return &services.PaymentError{
			Code:     "payment_method_removal_failed",
			Message:  fmt.Sprintf("failed to remove payment method: %v", err),
			Provider: "stripe",
		}
	}
	return nil
}

func (g *StripeGateway) ListPaymentMethods(ctx context.Context, customerID string) ([]*services.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)
	var paymentMethods []*services.PaymentMethod

	for iter.Next() {
		paymentMethods = append(paymentMethods, g.convertStripePaymentMethod(iter.PaymentMethod()))
	}

	if err := iter.Err(); err != nil {
		return nil, &services.PaymentError{
			Code:     "payment_method_list_failed",
			Message:  fmt.Sprintf("failed to list payment methods: %v", err),
			Provider: "stripe",
		}
	}

	return paymentMethods, nil
}

// Payment processing implementation

func (g *StripeGateway) CreateCharge(ctx context.Context, req services.CreateChargeRequest) (*services.Charge, error) {
	params := &stripe.ChargeParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(req.Currency),
		Customer:    stripe.String(req.CustomerID),
		Description: stripe.String(req.Description),
		Metadata:    req.Metadata,
		Capture:     stripe.Bool(req.Capture),
	}

	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
	}

	stripeCharge, err := charge.New(params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "charge_creation_failed",
			Message:  fmt.Sprintf("failed to create charge: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCharge(stripeCharge), nil
}

func (g *StripeGateway) GetCharge(ctx context.Context, chargeID string) (*services.Charge, error) {
	stripeCharge, err := charge.Get(chargeID, nil)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "charge_retrieval_failed",
			Message:  fmt.Sprintf("failed to retrieve charge: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCharge(stripeCharge), nil
}

func (g *StripeGateway) UpdateCharge(ctx context.Context, chargeID string, req services.UpdateChargeRequest) (*services.Charge, error) {
	params := &stripe.ChargeParams{}

	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	stripeCharge, err := charge.Update(chargeID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "charge_update_failed",
			Message:  fmt.Sprintf("failed to update charge: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCharge(stripeCharge), nil
}

func (g *StripeGateway) CaptureCharge(ctx context.Context, chargeID string, req services.CaptureChargeRequest) (*services.Charge, error) {
	params := &stripe.CaptureParams{}

	if req.Amount > 0 {
		params.Amount = stripe.Int64(req.Amount)
	}

	stripeCharge, err := charge.Capture(chargeID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "charge_capture_failed",
			Message:  fmt.Sprintf("failed to capture charge: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeCharge(stripeCharge), nil
}

func (g *StripeGateway) ListCharges(ctx context.Context, req services.ListChargesRequest) (*services.ChargeList, error) {
	params := &stripe.ChargeListParams{
		Limit: stripe.Int64(int64(req.Limit)),
	}

	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}
	if req.Status != "" {
		params.Filters.AddFilter("status", "", req.Status)
	}

	iter := charge.List(params)
	var charges []*services.Charge
	total := 0

	for iter.Next() {
		charges = append(charges, g.convertStripeCharge(iter.Charge()))
		total++
	}

	if err := iter.Err(); err != nil {
		return nil, &services.PaymentError{
			Code:     "charge_list_failed",
			Message:  fmt.Sprintf("failed to list charges: %v", err),
			Provider: "stripe",
		}
	}

	return &services.ChargeList{
		Charges: charges,
		Total:   total,
		HasMore: iter.Meta().HasMore,
	}, nil
}

// Refund processing implementation

func (g *StripeGateway) CreateRefund(ctx context.Context, req services.CreateRefundRequest) (*services.Refund, error) {
	params := &stripe.RefundParams{
		Charge: stripe.String(req.ChargeID),
		Reason: stripe.String(req.Reason),
	}

	if req.Amount > 0 {
		params.Amount = stripe.Int64(req.Amount)
	}

	stripeRefund, err := refund.New(params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "refund_creation_failed",
			Message:  fmt.Sprintf("failed to create refund: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeRefund(stripeRefund), nil
}

func (g *StripeGateway) GetRefund(ctx context.Context, refundID string) (*services.Refund, error) {
	stripeRefund, err := refund.Get(refundID, nil)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "refund_retrieval_failed",
			Message:  fmt.Sprintf("failed to retrieve refund: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeRefund(stripeRefund), nil
}

func (g *StripeGateway) UpdateRefund(ctx context.Context, refundID string, req services.UpdateRefundRequest) (*services.Refund, error) {
	params := &stripe.RefundParams{}

	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	stripeRefund, err := refund.Update(refundID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "refund_update_failed",
			Message:  fmt.Sprintf("failed to update refund: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeRefund(stripeRefund), nil
}

func (g *StripeGateway) ListRefunds(ctx context.Context, req services.ListRefundsRequest) (*services.RefundList, error) {
	params := &stripe.RefundListParams{
		Limit: stripe.Int64(int64(req.Limit)),
	}

	if req.ChargeID != "" {
		params.Charge = stripe.String(req.ChargeID)
	}

	iter := refund.List(params)
	var refunds []*services.Refund
	total := 0

	for iter.Next() {
		refunds = append(refunds, g.convertStripeRefund(iter.Refund()))
		total++
	}

	if err := iter.Err(); err != nil {
		return nil, &services.PaymentError{
			Code:     "refund_list_failed",
			Message:  fmt.Sprintf("failed to list refunds: %v", err),
			Provider: "stripe",
		}
	}

	return &services.RefundList{
		Refunds: refunds,
		Total:   total,
		HasMore: iter.Meta().HasMore,
	}, nil
}

// Subscription management implementation

func (g *StripeGateway) CreateSubscription(ctx context.Context, req services.CreateSubscriptionRequest) (*services.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PlanID),
			},
		},
		Metadata: req.Metadata,
	}

	stripeSubscription, err := subscription.New(params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "subscription_creation_failed",
			Message:  fmt.Sprintf("failed to create subscription: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeSubscription(stripeSubscription), nil
}

func (g *StripeGateway) GetSubscription(ctx context.Context, subscriptionID string) (*services.Subscription, error) {
	stripeSubscription, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "subscription_retrieval_failed",
			Message:  fmt.Sprintf("failed to retrieve subscription: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeSubscription(stripeSubscription), nil
}

func (g *StripeGateway) UpdateSubscription(ctx context.Context, subscriptionID string, req services.UpdateSubscriptionRequest) (*services.Subscription, error) {
	params := &stripe.SubscriptionParams{}

	if req.PlanID != "" {
		params.Items = []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PlanID),
			},
		}
	}
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	stripeSubscription, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "subscription_update_failed",
			Message:  fmt.Sprintf("failed to update subscription: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeSubscription(stripeSubscription), nil
}

func (g *StripeGateway) CancelSubscription(ctx context.Context, subscriptionID string, req services.CancelSubscriptionRequest) (*services.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(req.AtPeriodEnd),
	}

	stripeSubscription, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, &services.PaymentError{
			Code:     "subscription_cancellation_failed",
			Message:  fmt.Sprintf("failed to cancel subscription: %v", err),
			Provider: "stripe",
		}
	}

	return g.convertStripeSubscription(stripeSubscription), nil
}

func (g *StripeGateway) ListSubscriptions(ctx context.Context, req services.ListSubscriptionsRequest) (*services.SubscriptionList, error) {
	params := &stripe.SubscriptionListParams{
		Limit: stripe.Int64(int64(req.Limit)),
	}

	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}
	if req.Status != "" {
		params.Status = stripe.String(req.Status)
	}

	iter := subscription.List(params)
	var subscriptions []*services.Subscription
	total := 0

	for iter.Next() {
		subscriptions = append(subscriptions, g.convertStripeSubscription(iter.Subscription()))
		total++
	}

	if err := iter.Err(); err != nil {
		return nil, &services.PaymentError{
			Code:     "subscription_list_failed",
			Message:  fmt.Sprintf("failed to list subscriptions: %v", err),
			Provider: "stripe",
		}
	}

	return &services.SubscriptionList{
		Subscriptions: subscriptions,
		Total:         total,
		HasMore:       iter.Meta().HasMore,
	}, nil
}

// Conversion helper methods

func (g *StripeGateway) convertStripeCustomer(sc *stripe.Customer) *services.Customer {
	c := &services.Customer{
		ID:         sc.ID,
		Email:      sc.Email,
		Name:       sc.Name,
		Phone:      sc.Phone,
		Metadata:   sc.Metadata,
		CreatedAt:  time.Unix(sc.Created, 0),
		UpdatedAt:  time.Unix(sc.Created, 0), // Stripe doesn't provide updated_at
		ProviderID: sc.ID,
		Provider:   "stripe",
	}

	if sc.Address != nil {
		c.Address = &services.Address{
			Line1:      sc.Address.Line1,
			Line2:      sc.Address.Line2,
			City:       sc.Address.City,
			State:      sc.Address.State,
			PostalCode: sc.Address.PostalCode,
			Country:    sc.Address.Country,
		}
	}

	return c
}

func (g *StripeGateway) convertStripePaymentMethod(spm *stripe.PaymentMethod) *services.PaymentMethod {
	pm := &services.PaymentMethod{
		ID:         spm.ID,
		CustomerID: spm.Customer.ID,
		Type:       string(spm.Type),
		Metadata:   spm.Metadata,
		CreatedAt:  time.Unix(spm.Created, 0),
		ProviderID: spm.ID,
		Provider:   "stripe",
	}

	if spm.Card != nil {
		pm.Card = &services.Card{
			Brand:       string(spm.Card.Brand),
			Last4:       spm.Card.Last4,
			ExpMonth:    int(spm.Card.ExpMonth),
			ExpYear:     int(spm.Card.ExpYear),
			Fingerprint: spm.Card.Fingerprint,
			Country:     spm.Card.Country,
		}
	}

	return pm
}

func (g *StripeGateway) convertStripeCharge(sc *stripe.Charge) *services.Charge {
	c := &services.Charge{
		ID:              sc.ID,
		Amount:          sc.Amount,
		Currency:        string(sc.Currency),
		CustomerID:      sc.Customer.ID,
		PaymentMethodID: sc.PaymentMethod.ID,
		Status:          string(sc.Status),
		Description:     sc.Description,
		Metadata:        sc.Metadata,
		CreatedAt:       time.Unix(sc.Created, 0),
		UpdatedAt:       time.Unix(sc.Created, 0), // Stripe doesn't provide updated_at
		ProviderID:      sc.ID,
		Provider:        "stripe",
	}

	return c
}

func (g *StripeGateway) convertStripeRefund(sr *stripe.Refund) *services.Refund {
	r := &services.Refund{
		ID:         sr.ID,
		ChargeID:   sr.Charge.ID,
		Amount:     sr.Amount,
		Currency:   string(sr.Currency),
		Reason:     string(sr.Reason),
		Status:     string(sr.Status),
		Metadata:   sr.Metadata,
		CreatedAt:  time.Unix(sr.Created, 0),
		UpdatedAt:  time.Unix(sr.Created, 0), // Stripe doesn't provide updated_at
		ProviderID: sr.ID,
		Provider:   "stripe",
	}

	return r
}

func (g *StripeGateway) convertStripeSubscription(ss *stripe.Subscription) *services.Subscription {
	s := &services.Subscription{
		ID:           ss.ID,
		CustomerID:   ss.Customer.ID,
		PlanID:       ss.Items.Data[0].Price.ID,
		Status:       string(ss.Status),
		Metadata:     ss.Metadata,
		CreatedAt:    time.Unix(ss.Created, 0),
		UpdatedAt:    time.Unix(ss.Created, 0), // Stripe doesn't provide updated_at
		ProviderID:   ss.ID,
		Provider:     "stripe",
	}

	if ss.CurrentPeriodStart > 0 {
		s.CurrentPeriodStart = time.Unix(ss.CurrentPeriodStart, 0)
	}
	if ss.CurrentPeriodEnd > 0 {
		s.CurrentPeriodEnd = time.Unix(ss.CurrentPeriodEnd, 0)
	}

	return s
}