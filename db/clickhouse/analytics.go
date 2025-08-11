package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"apis/payments/services/stripe"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AnalyticsService provides ClickHouse analytics operations
type AnalyticsService struct {
	conn   clickhouse.Conn
	tracer trace.Tracer
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(conn clickhouse.Conn) *AnalyticsService {
	return &AnalyticsService{
		conn:   conn,
		tracer: otel.Tracer("payments.analytics"),
	}
}

// LogCharge logs a charge event to ClickHouse for analytics
func (a *AnalyticsService) LogCharge(ctx context.Context, charge *stripe.Charge) error {
	ctx, span := a.tracer.Start(ctx, "AnalyticsService.LogCharge")
	defer span.End()

	// Prepare the charge data for ClickHouse
	chargeData := map[string]interface{}{
		"id":                charge.ID,
		"amount":            charge.Amount,
		"currency":          charge.Currency,
		"status":            charge.Status,
		"customer_id":       charge.CustomerID,
		"payment_method_id": charge.PaymentMethodID,
		"description":       charge.Description,
		"created_at":        time.Unix(charge.Created, 0),
		"event_type":        "charge_created",
		"timestamp":         time.Now(),
	}

	// Convert metadata to JSON
	if charge.Metadata != nil {
		if metadataJSON, err := json.Marshal(charge.Metadata); err == nil {
			chargeData["metadata"] = string(metadataJSON)
		}
	}

	// Insert into ClickHouse
	query := `
		INSERT INTO payment_events (
			event_id, event_type, customer_id, amount, currency, status, 
			metadata, created_at, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	err := a.conn.Exec(ctx, query,
		charge.ID,
		"charge_created",
		charge.CustomerID,
		charge.Amount,
		charge.Currency,
		charge.Status,
		chargeData["metadata"],
		chargeData["created_at"],
		chargeData["timestamp"],
	)

	if err != nil {
		return fmt.Errorf("failed to log charge to ClickHouse: %w", err)
	}

	return nil
}

// LogCustomerEvent logs a customer event to ClickHouse
func (a *AnalyticsService) LogCustomerEvent(ctx context.Context, eventType string, customer *stripe.Customer) error {
	ctx, span := a.tracer.Start(ctx, "AnalyticsService.LogCustomerEvent")
	defer span.End()

	// Prepare the customer data for ClickHouse
	customerData := map[string]interface{}{
		"id":          customer.ID,
		"email":       customer.Email,
		"name":        customer.Name,
		"phone":       customer.Phone,
		"description": customer.Description,
		"created_at":  time.Unix(customer.Created, 0),
		"event_type":  eventType,
		"timestamp":   time.Now(),
	}

	// Convert metadata to JSON
	var metadataJSON string
	if customer.Metadata != nil {
		if jsonData, err := json.Marshal(customer.Metadata); err == nil {
			metadataJSON = string(jsonData)
		}
	}

	// Insert into ClickHouse
	query := `
		INSERT INTO customer_events (
			event_id, event_type, customer_id, email, name, phone, 
			description, metadata, created_at, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	err := a.conn.Exec(ctx, query,
		customer.ID,
		eventType,
		customer.ID,
		customer.Email,
		customer.Name,
		customer.Phone,
		customer.Description,
		metadataJSON,
		customerData["created_at"],
		customerData["timestamp"],
	)

	if err != nil {
		return fmt.Errorf("failed to log customer event to ClickHouse: %w", err)
	}

	return nil
}

// LogPaymentMethodEvent logs a payment method event to ClickHouse
func (a *AnalyticsService) LogPaymentMethodEvent(ctx context.Context, eventType string, paymentMethod *stripe.PaymentMethod) error {
	ctx, span := a.tracer.Start(ctx, "AnalyticsService.LogPaymentMethodEvent")
	defer span.End()

	// Prepare the payment method data for ClickHouse
	paymentMethodData := map[string]interface{}{
		"id":          paymentMethod.ID,
		"type":        paymentMethod.Type,
		"customer_id": paymentMethod.Customer,
		"created_at":  time.Unix(paymentMethod.Created, 0),
		"event_type":  eventType,
		"timestamp":   time.Now(),
	}

	// Add card details if available
	if paymentMethod.Card != nil {
		paymentMethodData["card_last4"] = paymentMethod.Card.Last4
		paymentMethodData["card_brand"] = paymentMethod.Card.Brand
		paymentMethodData["card_exp_month"] = paymentMethod.Card.ExpMonth
		paymentMethodData["card_exp_year"] = paymentMethod.Card.ExpYear
	}

	// Convert metadata to JSON
	var metadataJSON string
	if paymentMethod.Metadata != nil {
		if jsonData, err := json.Marshal(paymentMethod.Metadata); err == nil {
			metadataJSON = string(jsonData)
		}
	}

	// Insert into ClickHouse
	query := `
		INSERT INTO payment_method_events (
			event_id, event_type, payment_method_id, customer_id, type,
			card_last4, card_brand, card_exp_month, card_exp_year,
			metadata, created_at, timestamp
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	err := a.conn.Exec(ctx, query,
		paymentMethod.ID,
		eventType,
		paymentMethod.ID,
		paymentMethod.Customer,
		paymentMethod.Type,
		paymentMethodData["card_last4"],
		paymentMethodData["card_brand"],
		paymentMethodData["card_exp_month"],
		paymentMethodData["card_exp_year"],
		metadataJSON,
		paymentMethodData["created_at"],
		paymentMethodData["timestamp"],
	)

	if err != nil {
		return fmt.Errorf("failed to log payment method event to ClickHouse: %w", err)
	}

	return nil
}

// GetChargeMetrics retrieves charge metrics from ClickHouse
func (a *AnalyticsService) GetChargeMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	ctx, span := a.tracer.Start(ctx, "AnalyticsService.GetChargeMetrics")
	defer span.End()

	query := `
		SELECT 
			count() as total_charges,
			sum(amount) as total_amount,
			avg(amount) as avg_amount,
			countIf(status = 'succeeded') as successful_charges,
			sumIf(amount, status = 'succeeded') as successful_amount
		FROM payment_events 
		WHERE event_type = 'charge_created' 
		AND timestamp >= now() - INTERVAL ? DAY
	`

	var result struct {
		TotalCharges      uint64  `ch:"total_charges"`
		TotalAmount       uint64  `ch:"total_amount"`
		AvgAmount         float64 `ch:"avg_amount"`
		SuccessfulCharges uint64  `ch:"successful_charges"`
		SuccessfulAmount  uint64  `ch:"successful_amount"`
	}

	err := a.conn.QueryRow(ctx, query, days).ScanStruct(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get charge metrics: %w", err)
	}

	metrics := map[string]interface{}{
		"total_charges":      result.TotalCharges,
		"total_amount":       result.TotalAmount,
		"avg_amount":         result.AvgAmount,
		"successful_charges": result.SuccessfulCharges,
		"successful_amount":  result.SuccessfulAmount,
		"success_rate":       float64(result.SuccessfulCharges) / float64(result.TotalCharges) * 100,
		"period_days":        days,
		"timestamp":          time.Now(),
	}

	return metrics, nil
}

// GetCustomerMetrics retrieves customer metrics from ClickHouse
func (a *AnalyticsService) GetCustomerMetrics(ctx context.Context, days int) (map[string]interface{}, error) {
	ctx, span := a.tracer.Start(ctx, "AnalyticsService.GetCustomerMetrics")
	defer span.End()

	query := `
		SELECT 
			count() as total_customers,
			countIf(event_type = 'customer_created') as new_customers
		FROM customer_events 
		WHERE timestamp >= now() - INTERVAL ? DAY
	`

	var result struct {
		TotalCustomers uint64 `ch:"total_customers"`
		NewCustomers   uint64 `ch:"new_customers"`
	}

	err := a.conn.QueryRow(ctx, query, days).ScanStruct(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer metrics: %w", err)
	}

	metrics := map[string]interface{}{
		"total_customers": result.TotalCustomers,
		"new_customers":   result.NewCustomers,
		"period_days":     days,
		"timestamp":       time.Now(),
	}

	return metrics, nil
}
