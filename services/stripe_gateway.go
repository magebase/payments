package services

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/charge"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/dispute"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/subscription"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StripeGateway implements the PaymentGateway interface for Stripe
type StripeGateway struct {
	config    *ProviderConfig
	validator *validator.Validate
	tracer    trace.Tracer
}

// NewStripeGateway creates a new Stripe gateway instance
func NewStripeGateway(config *ProviderConfig) (*StripeGateway, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Stripe configuration: %w", err)
	}

	// Set Stripe API key
	stripe.Key = config.APIKey

	return &StripeGateway{
		config:    config,
		validator: validator.New(),
		tracer:    otel.Tracer("payments.stripe"),
	}, nil
}

// GetProviderName returns the name of this payment provider
func (g *StripeGateway) GetProviderName() string {
	return "stripe"
}

// GetCapabilities returns the capabilities supported by Stripe
func (g *StripeGateway) GetCapabilities() GatewayCapabilities {
	return GatewayCapabilities{
		SupportsSubscriptions: true,
		SupportsConnect:       true,
		SupportsTax:           true,
		SupportsInvoices:      true,
		SupportsPayouts:       true,
		SupportsDisputes:      true,
		SupportsRefunds:       true,
		MaxPaymentAmount:      99999999, // $999,999.99 in cents
		SupportedCurrencies:   []string{"usd", "eur", "gbp", "cad", "aud", "jpy"},
		SupportedCountries:    []string{"US", "CA", "GB", "DE", "FR", "AU", "JP"},
	}
}

// CreateCustomer creates a new customer using Stripe
func (g *StripeGateway) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateCustomer")
	defer span.End()

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe customer params
	params := &stripe.CustomerParams{
		Email:       stripe.String(req.Email),
		Name:        stripe.String(req.Name),
		Phone:       stripe.String(req.Phone),
		Description: stripe.String(req.Description),
		Metadata:    req.Metadata,
	}

	// Create customer in Stripe
	stripeCustomer, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Convert to common Customer type
	return &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     stripeCustomer.Created, // Stripe doesn't have updated field
		ProviderID:  stripeCustomer.ID,
	}, nil
}

// GetCustomer retrieves a customer from Stripe
func (g *StripeGateway) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetCustomer")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Get customer from Stripe
	stripeCustomer, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe customer: %w", err)
	}

	// Convert to common Customer type
	return &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     stripeCustomer.Created, // Stripe doesn't have updated field
		ProviderID:  stripeCustomer.ID,
	}, nil
}

// UpdateCustomer updates a customer in Stripe
func (g *StripeGateway) UpdateCustomer(ctx context.Context, customerID string, req *CustomerRequest) (*Customer, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.UpdateCustomer")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe customer params
	params := &stripe.CustomerParams{
		Email:       stripe.String(req.Email),
		Name:        stripe.String(req.Name),
		Phone:       stripe.String(req.Phone),
		Description: stripe.String(req.Description),
		Metadata:    req.Metadata,
	}

	// Update customer in Stripe
	stripeCustomer, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update Stripe customer: %w", err)
	}

	// Convert to common Customer type
	return &Customer{
		ID:          stripeCustomer.ID,
		Email:       stripeCustomer.Email,
		Name:        stripeCustomer.Name,
		Phone:       stripeCustomer.Phone,
		Description: stripeCustomer.Description,
		Metadata:    stripeCustomer.Metadata,
		Created:     stripeCustomer.Created,
		Updated:     stripeCustomer.Created, // Stripe doesn't have updated field
		ProviderID:  stripeCustomer.ID,
	}, nil
}

// DeleteCustomer deletes a customer from Stripe
func (g *StripeGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.DeleteCustomer")
	defer span.End()

	if customerID == "" {
		return fmt.Errorf("customer ID is required")
	}

	// Delete customer in Stripe
	_, err := customer.Del(customerID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Stripe customer: %w", err)
	}

	return nil
}

// AddPaymentMethod adds a payment method to a customer
func (g *StripeGateway) AddPaymentMethod(ctx context.Context, req *PaymentMethodRequest) (*PaymentMethod, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.AddPaymentMethod")
	defer span.End()

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe payment method params
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(req.Type),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(req.Card.Token),
		},
		Customer: stripe.String(req.Customer),
		Metadata: req.Metadata,
	}

	// Create payment method in Stripe
	stripePaymentMethod, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment method: %w", err)
	}

	// Convert to common PaymentMethod type
	return &PaymentMethod{
		ID:         stripePaymentMethod.ID,
		Type:       string(stripePaymentMethod.Type),
		Customer:   req.Customer,
		Metadata:   stripePaymentMethod.Metadata,
		Created:    stripePaymentMethod.Created,
		ProviderID: stripePaymentMethod.ID,
		Card: &Card{
			Last4:       stripePaymentMethod.Card.Last4,
			Brand:       string(stripePaymentMethod.Card.Brand),
			ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
			ExpYear:     int(stripePaymentMethod.Card.ExpYear),
			Fingerprint: stripePaymentMethod.Card.Fingerprint,
		},
	}, nil
}

// GetPaymentMethod retrieves a payment method from Stripe
func (g *StripeGateway) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethod, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetPaymentMethod")
	defer span.End()

	if paymentMethodID == "" {
		return nil, fmt.Errorf("payment method ID is required")
	}

	// Get payment method from Stripe
	stripePaymentMethod, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe payment method: %w", err)
	}

	// Convert to common PaymentMethod type
	return &PaymentMethod{
		ID:         stripePaymentMethod.ID,
		Type:       string(stripePaymentMethod.Type),
		Customer:   stripePaymentMethod.Customer.ID,
		Metadata:   stripePaymentMethod.Metadata,
		Created:    stripePaymentMethod.Created,
		ProviderID: stripePaymentMethod.ID,
		Card: &Card{
			Last4:       stripePaymentMethod.Card.Last4,
			Brand:       string(stripePaymentMethod.Card.Brand),
			ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
			ExpYear:     int(stripePaymentMethod.Card.ExpYear),
			Fingerprint: stripePaymentMethod.Card.Fingerprint,
		},
	}, nil
}

// ListPaymentMethods lists payment methods for a customer
func (g *StripeGateway) ListPaymentMethods(ctx context.Context, customerID string, limit int) ([]*PaymentMethod, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListPaymentMethods")
	defer span.End()

	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	// List payment methods from Stripe
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	params.Limit = stripe.Int64(int64(limit))

	iter := paymentmethod.List(params)
	var paymentMethods []*PaymentMethod

	for iter.Next() {
		stripePaymentMethod := iter.PaymentMethod()
		paymentMethod := &PaymentMethod{
			ID:         stripePaymentMethod.ID,
			Type:       string(stripePaymentMethod.Type),
			Customer:   customerID,
			Metadata:   stripePaymentMethod.Metadata,
			Created:    stripePaymentMethod.Created,
			ProviderID: stripePaymentMethod.ID,
			Card: &Card{
				Last4:       stripePaymentMethod.Card.Last4,
				Brand:       string(stripePaymentMethod.Card.Brand),
				ExpMonth:    int(stripePaymentMethod.Card.ExpMonth),
				ExpYear:     int(stripePaymentMethod.Card.ExpYear),
				Fingerprint: stripePaymentMethod.Card.Fingerprint,
			},
		}
		paymentMethods = append(paymentMethods, paymentMethod)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe payment methods: %w", err)
	}

	return paymentMethods, nil
}

// DetachPaymentMethod detaches a payment method from a customer
func (g *StripeGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.DetachPaymentMethod")
	defer span.End()

	if paymentMethodID == "" {
		return fmt.Errorf("payment method ID is required")
	}

	// Detach payment method in Stripe
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach Stripe payment method: %w", err)
	}

	return nil
}

// CreateCharge creates a charge using Stripe
func (g *StripeGateway) CreateCharge(ctx context.Context, req *ChargeRequest) (*Charge, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateCharge")
	defer span.End()

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe charge params
	params := &stripe.ChargeParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(req.Currency),
		Customer:    stripe.String(req.CustomerID),
		Description: stripe.String(req.Description),
		Metadata:    req.Metadata,
	}

	if req.PaymentMethod != "" {
		params.Source = &stripe.PaymentSourceSourceParams{
			Token: stripe.String(req.PaymentMethod),
		}
	}

	// Create charge in Stripe
	stripeCharge, err := charge.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe charge: %w", err)
	}

	// Convert to common Charge type
	paymentMethodID := ""
	// For simplicity, we'll use the source ID directly
	if stripeCharge.Source != nil {
		paymentMethodID = stripeCharge.Source.ID
	}

	return &Charge{
		ID:              stripeCharge.ID,
		Amount:          stripeCharge.Amount,
		Currency:        string(stripeCharge.Currency),
		Status:          string(stripeCharge.Status),
		CustomerID:      stripeCharge.Customer.ID,
		PaymentMethodID: paymentMethodID,
		Description:     stripeCharge.Description,
		Metadata:        stripeCharge.Metadata,
		Created:         stripeCharge.Created,
		Updated:         stripeCharge.Created, // Stripe doesn't have updated field
		ProviderID:      stripeCharge.ID,
	}, nil
}

// GetCharge retrieves a charge from Stripe
func (g *StripeGateway) GetCharge(ctx context.Context, chargeID string) (*Charge, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetCharge")
	defer span.End()

	if chargeID == "" {
		return nil, fmt.Errorf("charge ID is required")
	}

	// Get charge from Stripe
	stripeCharge, err := charge.Get(chargeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe charge: %w", err)
	}

	// Convert to common Charge type
	paymentMethodID := ""
	if stripeCharge.Source != nil {
		paymentMethodID = stripeCharge.Source.ID
	}

	return &Charge{
		ID:              stripeCharge.ID,
		Amount:          stripeCharge.Amount,
		Currency:        string(stripeCharge.Currency),
		Status:          string(stripeCharge.Status),
		CustomerID:      stripeCharge.Customer.ID,
		PaymentMethodID: paymentMethodID,
		Description:     stripeCharge.Description,
		Metadata:        stripeCharge.Metadata,
		Created:         stripeCharge.Created,
		Updated:         stripeCharge.Created,
		ProviderID:      stripeCharge.ID,
	}, nil
}

// ListCharges lists charges for a customer
func (g *StripeGateway) ListCharges(ctx context.Context, customerID string, limit int) ([]*Charge, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListCharges")
	defer span.End()

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	// List charges from Stripe
	params := &stripe.ChargeListParams{}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}

	iter := charge.List(params)
	var charges []*Charge

	for iter.Next() {
		stripeCharge := iter.Charge()

		paymentMethodID := ""
		if stripeCharge.Source != nil {
			paymentMethodID = stripeCharge.Source.ID
		}

		charge := &Charge{
			ID:              stripeCharge.ID,
			Amount:          stripeCharge.Amount,
			Currency:        string(stripeCharge.Currency),
			Status:          string(stripeCharge.Status),
			CustomerID:      stripeCharge.Customer.ID,
			PaymentMethodID: paymentMethodID,
			Description:     stripeCharge.Description,
			Metadata:        stripeCharge.Metadata,
			Created:         stripeCharge.Created,
			Updated:         stripeCharge.Created,
			ProviderID:      stripeCharge.ID,
		}
		charges = append(charges, charge)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe charges: %w", err)
	}

	return charges, nil
}

// CreateRefund creates a refund using Stripe
func (g *StripeGateway) CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateRefund")
	defer span.End()

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to Stripe refund params
	params := &stripe.RefundParams{
		Charge:   stripe.String(req.ChargeID),
		Reason:   stripe.String(req.Reason),
		Metadata: req.Metadata,
	}

	if req.Amount > 0 {
		params.Amount = stripe.Int64(req.Amount)
	}

	// Create refund in Stripe
	stripeRefund, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	// Convert to common Refund type
	return &Refund{
		ID:         stripeRefund.ID,
		ChargeID:   stripeRefund.Charge.ID,
		Amount:     stripeRefund.Amount,
		Currency:   string(stripeRefund.Currency),
		Status:     string(stripeRefund.Status),
		Reason:     string(stripeRefund.Reason),
		Metadata:   stripeRefund.Metadata,
		Created:    stripeRefund.Created,
		Updated:    stripeRefund.Created,
		ProviderID: stripeRefund.ID,
	}, nil
}

// GetRefund retrieves a refund from Stripe
func (g *StripeGateway) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetRefund")
	defer span.End()

	if refundID == "" {
		return nil, fmt.Errorf("refund ID is required")
	}

	// Get refund from Stripe
	stripeRefund, err := refund.Get(refundID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe refund: %w", err)
	}

	// Convert to common Refund type
	return &Refund{
		ID:         stripeRefund.ID,
		ChargeID:   stripeRefund.Charge.ID,
		Amount:     stripeRefund.Amount,
		Currency:   string(stripeRefund.Currency),
		Status:     string(stripeRefund.Status),
		Reason:     string(stripeRefund.Reason),
		Metadata:   stripeRefund.Metadata,
		Created:    stripeRefund.Created,
		Updated:    stripeRefund.Created,
		ProviderID: stripeRefund.ID,
	}, nil
}

// ListRefunds lists refunds for a charge
func (g *StripeGateway) ListRefunds(ctx context.Context, chargeID string, limit int) ([]*Refund, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListRefunds")
	defer span.End()

	if chargeID == "" {
		return nil, fmt.Errorf("charge ID is required")
	}

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	// List refunds from Stripe
	params := &stripe.RefundListParams{
		Charge: stripe.String(chargeID),
	}

	iter := refund.List(params)
	var refunds []*Refund

	for iter.Next() {
		stripeRefund := iter.Refund()
		refund := &Refund{
			ID:         stripeRefund.ID,
			ChargeID:   stripeRefund.Charge.ID,
			Amount:     stripeRefund.Amount,
			Currency:   string(stripeRefund.Currency),
			Status:     string(stripeRefund.Status),
			Reason:     string(stripeRefund.Reason),
			Metadata:   stripeRefund.Metadata,
			Created:    stripeRefund.Created,
			Updated:    stripeRefund.Created,
			ProviderID: stripeRefund.ID,
		}
		refunds = append(refunds, refund)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe refunds: %w", err)
	}

	return refunds, nil
}

// CreateDispute creates a dispute using Stripe
func (g *StripeGateway) CreateDispute(ctx context.Context, req *DisputeRequest) (*Dispute, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateDispute")
	defer span.End()

	// Validate the request
	if err := g.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Note: Stripe doesn't support creating disputes directly
	// Disputes are created automatically when customers dispute charges
	// This method would typically be used for internal dispute tracking
	return nil, fmt.Errorf("Stripe does not support creating disputes directly")
}

// GetDispute retrieves a dispute from Stripe
func (g *StripeGateway) GetDispute(ctx context.Context, disputeID string) (*Dispute, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetDispute")
	defer span.End()

	if disputeID == "" {
		return nil, fmt.Errorf("dispute ID is required")
	}

	// Get dispute from Stripe
	stripeDispute, err := dispute.Get(disputeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe dispute: %w", err)
	}

	// Convert to common Dispute type
	evidence := make(map[string]string)
	if stripeDispute.Evidence != nil {
		// Add basic evidence information
		evidence["has_evidence"] = "true"
	}

	return &Dispute{
		ID:         stripeDispute.ID,
		ChargeID:   stripeDispute.Charge.ID,
		Amount:     stripeDispute.Amount,
		Currency:   string(stripeDispute.Currency),
		Status:     string(stripeDispute.Status),
		Reason:     string(stripeDispute.Reason),
		Evidence:   evidence,
		Metadata:   stripeDispute.Metadata,
		Created:    stripeDispute.Created,
		Updated:    stripeDispute.Created,
		ProviderID: stripeDispute.ID,
	}, nil
}

// ListDisputes lists disputes for a charge
func (g *StripeGateway) ListDisputes(ctx context.Context, chargeID string, limit int) ([]*Dispute, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListDisputes")
	defer span.End()

	// Set default limit
	if limit <= 0 {
		limit = 100
	}

	// List disputes from Stripe
	params := &stripe.DisputeListParams{}
	if chargeID != "" {
		params.Charge = stripe.String(chargeID)
	}

	iter := dispute.List(params)
	var disputes []*Dispute

	for iter.Next() {
		stripeDispute := iter.Dispute()

		evidence := make(map[string]string)
		if stripeDispute.Evidence != nil {
			// Add basic evidence information
			evidence["has_evidence"] = "true"
		}

		dispute := &Dispute{
			ID:         stripeDispute.ID,
			ChargeID:   stripeDispute.Charge.ID,
			Amount:     stripeDispute.Amount,
			Currency:   string(stripeDispute.Currency),
			Status:     string(stripeDispute.Status),
			Reason:     string(stripeDispute.Reason),
			Evidence:   evidence,
			Metadata:   stripeDispute.Metadata,
			Created:    stripeDispute.Created,
			Updated:    stripeDispute.Created,
			ProviderID: stripeDispute.ID,
		}
		disputes = append(disputes, dispute)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe disputes: %w", err)
	}

	return disputes, nil
}

// UpdateDisputeStatus updates a dispute status
func (g *StripeGateway) UpdateDisputeStatus(ctx context.Context, disputeID string, status string) (*Dispute, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.UpdateDisputeStatus")
	defer span.End()

	if disputeID == "" {
		return nil, fmt.Errorf("dispute ID is required")
	}

	if status == "" {
		return nil, fmt.Errorf("status is required")
	}

	// Note: Stripe has limited dispute status update capabilities
	// This method would typically be used for internal dispute tracking
	return nil, fmt.Errorf("Stripe has limited dispute status update capabilities")
}

// Subscription plan operations
func (g *StripeGateway) CreateSubscriptionPlan(ctx context.Context, req *SubscriptionPlanRequest) (*SubscriptionPlan, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateSubscriptionPlan")
	defer span.End()

	if req == nil {
		return nil, fmt.Errorf("subscription plan request is required")
	}

	// Convert to Stripe product params
	productParams := &stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
		Metadata:    req.Metadata,
	}

	// Create product in Stripe
	stripeProduct, err := product.New(productParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe product: %w", err)
	}

	// Convert to Stripe price params
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(stripeProduct.ID),
		UnitAmount: stripe.Int64(req.Amount),
		Currency:   stripe.String(req.Currency),
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String(req.Interval),
			IntervalCount: stripe.Int64(int64(req.IntervalCount)),
		},
		Metadata: req.Metadata,
	}

	// Create price in Stripe
	stripePrice, err := price.New(priceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe price: %w", err)
	}

	// Convert to common SubscriptionPlan type
	return &SubscriptionPlan{
		ID:             stripePrice.ID,
		Name:           req.Name,
		Description:    req.Description,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Interval:       req.Interval,
		IntervalCount:  req.IntervalCount,
		TrialPeriodDays: req.TrialPeriodDays,
		Metadata:       req.Metadata,
		Created:        stripePrice.Created,
		Updated:        stripePrice.Created,
		ProviderID:     stripePrice.ID,
	}, nil
}

func (g *StripeGateway) GetSubscriptionPlan(ctx context.Context, planID string) (*SubscriptionPlan, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetSubscriptionPlan")
	defer span.End()

	if planID == "" {
		return nil, fmt.Errorf("plan ID is required")
	}

	// Get price from Stripe
	stripePrice, err := price.Get(planID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe price: %w", err)
	}

	// Get product from Stripe
	stripeProduct, err := product.Get(stripePrice.Product.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe product: %w", err)
	}

	// Convert to common SubscriptionPlan type
	return &SubscriptionPlan{
		ID:             stripePrice.ID,
		Name:           stripeProduct.Name,
		Description:    stripeProduct.Description,
		Amount:         stripePrice.UnitAmount,
		Currency:       string(stripePrice.Currency),
		Interval:       string(stripePrice.Recurring.Interval),
		IntervalCount:  int(stripePrice.Recurring.IntervalCount),
		Metadata:       stripePrice.Metadata,
		Created:        stripePrice.Created,
		Updated:        stripePrice.Created,
		ProviderID:     stripePrice.ID,
	}, nil
}

func (g *StripeGateway) ListSubscriptionPlans(ctx context.Context, params *SubscriptionPlanListParams) ([]*SubscriptionPlan, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListSubscriptionPlans")
	defer span.End()

	// List prices from Stripe
	priceParams := &stripe.PriceListParams{
		Active: stripe.Bool(true),
	}
	if params.Limit > 0 {
		priceParams.Limit = stripe.Int64(int64(params.Limit))
	}

	iter := price.List(priceParams)
	var plans []*SubscriptionPlan

	for iter.Next() {
		stripePrice := iter.Price()
		if stripePrice.Recurring == nil {
			continue // Skip non-recurring prices
		}

		// Get product for this price
		stripeProduct, err := product.Get(stripePrice.Product.ID, nil)
		if err != nil {
			continue // Skip if product not found
		}

		plan := &SubscriptionPlan{
			ID:             stripePrice.ID,
			Name:           stripeProduct.Name,
			Description:    stripeProduct.Description,
			Amount:         stripePrice.UnitAmount,
			Currency:       string(stripePrice.Currency),
			Interval:       string(stripePrice.Recurring.Interval),
			IntervalCount:  int(stripePrice.Recurring.IntervalCount),
			Metadata:       stripePrice.Metadata,
			Created:        stripePrice.Created,
			Updated:        stripePrice.Created,
			ProviderID:     stripePrice.ID,
		}
		plans = append(plans, plan)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe prices: %w", err)
	}

	return plans, nil
}

func (g *StripeGateway) UpdateSubscriptionPlan(ctx context.Context, planID string, req *SubscriptionPlanUpdateRequest) (*SubscriptionPlan, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.UpdateSubscriptionPlan")
	defer span.End()

	if planID == "" {
		return nil, fmt.Errorf("plan ID is required")
	}

	// Get current price
	stripePrice, err := price.Get(planID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe price: %w", err)
	}

	// Update product if name/description changed
	if req.Name != nil || req.Description != nil {
		productParams := &stripe.ProductParams{}
		if req.Name != nil {
			productParams.Name = stripe.String(*req.Name)
		}
		if req.Description != nil {
			productParams.Description = stripe.String(*req.Description)
		}

		_, err = product.Update(stripePrice.Product.ID, productParams)
		if err != nil {
			return nil, fmt.Errorf("failed to update Stripe product: %w", err)
		}
	}

	// Update metadata if provided
	if req.Metadata != nil {
		priceParams := &stripe.PriceParams{
			Metadata: req.Metadata,
		}
		_, err = price.Update(planID, priceParams)
		if err != nil {
			return nil, fmt.Errorf("failed to update Stripe price: %w", err)
		}
	}

	// Return updated plan
	return g.GetSubscriptionPlan(ctx, planID)
}

func (g *StripeGateway) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.DeleteSubscriptionPlan")
	defer span.End()

	if planID == "" {
		return fmt.Errorf("plan ID is required")
	}

	// Archive the price in Stripe (soft delete)
	priceParams := &stripe.PriceParams{
		Active: stripe.Bool(false),
	}
	_, err := price.Update(planID, priceParams)
	if err != nil {
		return fmt.Errorf("failed to archive Stripe price: %w", err)
	}

	return nil
}

// Subscription operations
func (g *StripeGateway) CreateSubscription(ctx context.Context, req *SubscriptionRequest) (*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CreateSubscription")
	defer span.End()

	if req == nil {
		return nil, fmt.Errorf("subscription request is required")
	}

	// Convert to Stripe subscription params
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PlanID),
			},
		},
		Metadata: req.Metadata,
	}

	if req.TrialEnd != nil {
		subscriptionParams.TrialEnd = stripe.Int64(*req.TrialEnd)
	}

	// Create subscription in Stripe
	stripeSubscription, err := subscription.New(subscriptionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}

	// Convert to common Subscription type
	return &Subscription{
		ID:                 stripeSubscription.ID,
		CustomerID:         req.CustomerID,
		PlanID:             req.PlanID,
		Status:             string(stripeSubscription.Status),
		CurrentPeriodStart: stripeSubscription.CurrentPeriodStart,
		CurrentPeriodEnd:   stripeSubscription.CurrentPeriodEnd,
		TrialStart:         &stripeSubscription.TrialStart,
		TrialEnd:           &stripeSubscription.TrialEnd,
		CanceledAt:         &stripeSubscription.CanceledAt,
		EndedAt:            &stripeSubscription.EndedAt,
		Metadata:           req.Metadata,
		Created:            stripeSubscription.Created,
		Updated:            stripeSubscription.Created,
		ProviderID:         stripeSubscription.ID,
	}, nil
}

func (g *StripeGateway) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.GetSubscription")
	defer span.End()

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Get subscription from Stripe
	stripeSubscription, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe subscription: %w", err)
	}

	// Convert to common Subscription type
	return &Subscription{
		ID:                 stripeSubscription.ID,
		CustomerID:         stripeSubscription.Customer.ID,
		PlanID:             stripeSubscription.Items.Data[0].Price.ID,
		Status:             string(stripeSubscription.Status),
		CurrentPeriodStart: stripeSubscription.CurrentPeriodStart,
		CurrentPeriodEnd:   stripeSubscription.CurrentPeriodEnd,
		TrialStart:         &stripeSubscription.TrialStart,
		TrialEnd:           &stripeSubscription.TrialEnd,
		CanceledAt:         &stripeSubscription.CanceledAt,
		EndedAt:            &stripeSubscription.EndedAt,
		Metadata:           stripeSubscription.Metadata,
		Created:            stripeSubscription.Created,
		Updated:            stripeSubscription.Created,
		ProviderID:         stripeSubscription.ID,
	}, nil
}

func (g *StripeGateway) ListSubscriptions(ctx context.Context, params *SubscriptionListParams) ([]*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ListSubscriptions")
	defer span.End()

	// List subscriptions from Stripe
	subscriptionParams := &stripe.SubscriptionListParams{}
	if params.CustomerID != "" {
		subscriptionParams.Customer = stripe.String(params.CustomerID)
	}
	if params.Status != "" {
		subscriptionParams.Status = stripe.String(params.Status)
	}
	if params.Limit > 0 {
		subscriptionParams.Limit = stripe.Int64(int64(params.Limit))
	}

	iter := subscription.List(subscriptionParams)
	var subscriptions []*Subscription

	for iter.Next() {
		stripeSubscription := iter.Subscription()
		if len(stripeSubscription.Items.Data) == 0 {
			continue
		}

		subscription := &Subscription{
			ID:                 stripeSubscription.ID,
			CustomerID:         stripeSubscription.Customer.ID,
			PlanID:             stripeSubscription.Items.Data[0].Price.ID,
			Status:             string(stripeSubscription.Status),
			CurrentPeriodStart: stripeSubscription.CurrentPeriodStart,
			CurrentPeriodEnd:   stripeSubscription.CurrentPeriodEnd,
			TrialStart:         &stripeSubscription.TrialStart,
			TrialEnd:           &stripeSubscription.TrialEnd,
			CanceledAt:         &stripeSubscription.CanceledAt,
			EndedAt:            &stripeSubscription.EndedAt,
			Metadata:           stripeSubscription.Metadata,
			Created:            stripeSubscription.Created,
			Updated:            stripeSubscription.Created,
			ProviderID:         stripeSubscription.ID,
		}
		subscriptions = append(subscriptions, subscription)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list Stripe subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (g *StripeGateway) UpdateSubscription(ctx context.Context, subscriptionID string, req *SubscriptionUpdateRequest) (*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.UpdateSubscription")
	defer span.End()

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Convert to Stripe subscription update params
	subscriptionParams := &stripe.SubscriptionParams{}
	if req.PlanID != nil {
		subscriptionParams.Items = []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(*req.PlanID),
			},
		}
	}
	if req.TrialEnd != nil {
		subscriptionParams.TrialEnd = stripe.Int64(*req.TrialEnd)
	}
	if req.Metadata != nil {
		subscriptionParams.Metadata = req.Metadata
	}

	// Update subscription in Stripe
	_, err := subscription.Update(subscriptionID, subscriptionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update Stripe subscription: %w", err)
	}

	// Return updated subscription
	return g.GetSubscription(ctx, subscriptionID)
}

func (g *StripeGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.CancelSubscription")
	defer span.End()

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Cancel subscription in Stripe
	subscriptionParams := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	_, err := subscription.Update(subscriptionID, subscriptionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel Stripe subscription: %w", err)
	}

	// Return canceled subscription
	return g.GetSubscription(ctx, subscriptionID)
}

func (g *StripeGateway) ReactivateSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	ctx, span := g.tracer.Start(ctx, "StripeGateway.ReactivateSubscription")
	defer span.End()

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Reactivate subscription in Stripe
	subscriptionParams := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	_, err := subscription.Update(subscriptionID, subscriptionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate Stripe subscription: %w", err)
	}

	// Return reactivated subscription
	return g.GetSubscription(ctx, subscriptionID)
}
