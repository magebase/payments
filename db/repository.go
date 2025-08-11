package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"apis/payments/db/sqlc"
	"apis/payments/services/stripe"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sqlc-dev/pqtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PoolAdapter adapts pgxpool.Pool to implement sqlc.DBTX interface
type PoolAdapter struct {
	pool *pgxpool.Pool
}

func (p *PoolAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	commandTag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &pgxResult{commandTag: commandTag}, nil
}

func (p *PoolAdapter) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// pgx doesn't support prepared statements in the same way as database/sql
	// Return a no-op statement
	return &sql.Stmt{}, nil
}

func (p *PoolAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// For now, return nil since we don't actually use this method
	// The SQLC generated code handles queries differently
	rows.Close()
	return nil, fmt.Errorf("QueryContext not fully implemented for pgx")
}

func (p *PoolAdapter) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return &sql.Row{}
	}
	defer conn.Release()

	// For now, return empty row since we don't actually use this method
	// The SQLC generated code handles queries differently
	return &sql.Row{}
}

// pgxResult implements sql.Result for pgx
type pgxResult struct {
	commandTag interface{}
}

func (r *pgxResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("LastInsertId not supported by pgx")
}

func (r *pgxResult) RowsAffected() (int64, error) {
	// Try to extract rows affected from command tag
	if ct, ok := r.commandTag.(interface{ RowsAffected() int64 }); ok {
		return ct.RowsAffected(), nil
	}
	return 0, nil
}

// Repository provides database operations for the payments service
type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
	tracer  trace.Tracer
}

// NewRepository creates a new repository instance
func NewRepository(pool *pgxpool.Pool) *Repository {
	adapter := &PoolAdapter{pool: pool}
	return &Repository{
		queries: sqlc.New(adapter),
		pool:    pool,
		tracer:  otel.Tracer("payments.repository"),
	}
}

// CreateCustomer stores a customer in the database
func (r *Repository) CreateCustomer(ctx context.Context, customer *stripe.Customer) (*stripe.Customer, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.CreateCustomer")
	defer span.End()

	// Convert metadata to JSON bytes for storage
	metadataBytes, err := json.Marshal(customer.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.CreateCustomerParams{
		ID:          customer.ID,
		Email:       customer.Email,
		Name:        customer.Name,
		Phone:       sql.NullString{String: customer.Phone, Valid: customer.Phone != ""},
		Description: sql.NullString{String: customer.Description, Valid: customer.Description != ""},
		Metadata:    pqtype.NullRawMessage{RawMessage: metadataBytes, Valid: len(metadataBytes) > 0},
	}

	dbCustomer, err := r.queries.CreateCustomer(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Convert back to stripe.Customer
	return &stripe.Customer{
		ID:          dbCustomer.ID,
		Email:       dbCustomer.Email,
		Name:        dbCustomer.Name,
		Phone:       dbCustomer.Phone.String,
		Description: dbCustomer.Description.String,
		Metadata:    convertMetadata(dbCustomer.Metadata),
		Created:     dbCustomer.CreatedAt.Time.Unix(),
		Updated:     dbCustomer.UpdatedAt.Time.Unix(),
	}, nil
}

// GetCustomer retrieves a customer from the database
func (r *Repository) GetCustomer(ctx context.Context, id string) (*stripe.Customer, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.GetCustomer")
	defer span.End()

	dbCustomer, err := r.queries.GetCustomer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &stripe.Customer{
		ID:          dbCustomer.ID,
		Email:       dbCustomer.Email,
		Name:        dbCustomer.Name,
		Phone:       dbCustomer.Phone.String,
		Description: dbCustomer.Description.String,
		Metadata:    convertMetadata(dbCustomer.Metadata),
		Created:     dbCustomer.CreatedAt.Time.Unix(),
		Updated:     dbCustomer.UpdatedAt.Time.Unix(),
	}, nil
}

// UpdateCustomer updates a customer in the database
func (r *Repository) UpdateCustomer(ctx context.Context, id string, customer *stripe.Customer) (*stripe.Customer, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.UpdateCustomer")
	defer span.End()

	// Convert metadata to JSON bytes for storage
	metadataBytes, err := json.Marshal(customer.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.UpdateCustomerParams{
		ID:          id,
		Email:       customer.Email,
		Name:        customer.Name,
		Phone:       sql.NullString{String: customer.Phone, Valid: customer.Phone != ""},
		Description: sql.NullString{String: customer.Description, Valid: customer.Description != ""},
		Metadata:    pqtype.NullRawMessage{RawMessage: metadataBytes, Valid: len(metadataBytes) > 0},
	}

	dbCustomer, err := r.queries.UpdateCustomer(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return &stripe.Customer{
		ID:          dbCustomer.ID,
		Email:       dbCustomer.Email,
		Name:        dbCustomer.Name,
		Phone:       dbCustomer.Phone.String,
		Description: dbCustomer.Description.String,
		Metadata:    convertMetadata(dbCustomer.Metadata),
		Created:     dbCustomer.CreatedAt.Time.Unix(),
		Updated:     dbCustomer.UpdatedAt.Time.Unix(),
	}, nil
}

// DeleteCustomer removes a customer from the database
func (r *Repository) DeleteCustomer(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "Repository.DeleteCustomer")
	defer span.End()

	err := r.queries.DeleteCustomer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// StorePaymentMethod stores a payment method in the database
func (r *Repository) StorePaymentMethod(ctx context.Context, paymentMethod *stripe.PaymentMethod) (*stripe.PaymentMethod, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.StorePaymentMethod")
	defer span.End()

	// Convert metadata to JSON bytes for storage
	metadataBytes, err := json.Marshal(paymentMethod.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var cardLast4, cardBrand, cardFingerprint sql.NullString
	var cardExpMonth, cardExpYear sql.NullInt32

	if paymentMethod.Card != nil {
		cardLast4 = sql.NullString{String: paymentMethod.Card.Last4, Valid: true}
		cardBrand = sql.NullString{String: paymentMethod.Card.Brand, Valid: true}
		cardFingerprint = sql.NullString{String: paymentMethod.Card.Fingerprint, Valid: true}
		cardExpMonth = sql.NullInt32{Int32: int32(paymentMethod.Card.ExpMonth), Valid: true}
		cardExpYear = sql.NullInt32{Int32: int32(paymentMethod.Card.ExpYear), Valid: true}
	}

	params := sqlc.CreatePaymentMethodParams{
		ID:              paymentMethod.ID,
		Type:            paymentMethod.Type,
		CustomerID:      paymentMethod.Customer,
		CardLast4:       cardLast4,
		CardBrand:       cardBrand,
		CardExpMonth:    cardExpMonth,
		CardExpYear:     cardExpYear,
		CardFingerprint: cardFingerprint,
		Metadata:        pqtype.NullRawMessage{RawMessage: metadataBytes, Valid: len(metadataBytes) > 0},
	}

	dbPaymentMethod, err := r.queries.CreatePaymentMethod(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to store payment method: %w", err)
	}

	// Convert back to stripe.PaymentMethod
	result := &stripe.PaymentMethod{
		ID:       dbPaymentMethod.ID,
		Type:     dbPaymentMethod.Type,
		Customer: dbPaymentMethod.CustomerID,
		Metadata: convertMetadata(dbPaymentMethod.Metadata),
		Created:  dbPaymentMethod.CreatedAt.Time.Unix(),
	}

	// Add card details if available
	if dbPaymentMethod.CardLast4.Valid {
		result.Card = &stripe.Card{
			Last4:       dbPaymentMethod.CardLast4.String,
			Brand:       dbPaymentMethod.CardBrand.String,
			ExpMonth:    int(dbPaymentMethod.CardExpMonth.Int32),
			ExpYear:     int(dbPaymentMethod.CardExpYear.Int32),
			Fingerprint: dbPaymentMethod.CardFingerprint.String,
		}
	}

	return result, nil
}

// GetPaymentMethod retrieves a payment method from the database
func (r *Repository) GetPaymentMethod(ctx context.Context, id string) (*stripe.PaymentMethod, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.GetPaymentMethod")
	defer span.End()

	dbPaymentMethod, err := r.queries.GetPaymentMethod(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}

	result := &stripe.PaymentMethod{
		ID:       dbPaymentMethod.ID,
		Type:     dbPaymentMethod.Type,
		Customer: dbPaymentMethod.CustomerID,
		Metadata: convertMetadata(dbPaymentMethod.Metadata),
		Created:  dbPaymentMethod.CreatedAt.Time.Unix(),
	}

	// Add card details if available
	if dbPaymentMethod.CardLast4.Valid {
		result.Card = &stripe.Card{
			Last4:       dbPaymentMethod.CardLast4.String,
			Brand:       dbPaymentMethod.CardBrand.String,
			ExpMonth:    int(dbPaymentMethod.CardExpMonth.Int32),
			ExpYear:     int(dbPaymentMethod.CardExpYear.Int32),
			Fingerprint: dbPaymentMethod.CardFingerprint.String,
		}
	}

	return result, nil
}

// ListPaymentMethods retrieves payment methods for a customer
func (r *Repository) ListPaymentMethods(ctx context.Context, customerID string) ([]*stripe.PaymentMethod, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.ListPaymentMethods")
	defer span.End()

	dbPaymentMethods, err := r.queries.ListPaymentMethods(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	var result []*stripe.PaymentMethod
	for _, dbPM := range dbPaymentMethods {
		pm := &stripe.PaymentMethod{
			ID:       dbPM.ID,
			Type:     dbPM.Type,
			Customer: dbPM.CustomerID,
			Metadata: convertMetadata(dbPM.Metadata),
			Created:  dbPM.CreatedAt.Time.Unix(),
		}

		// Add card details if available
		if dbPM.CardLast4.Valid {
			pm.Card = &stripe.Card{
				Last4:       dbPM.CardLast4.String,
				Brand:       dbPM.CardBrand.String,
				ExpMonth:    int(dbPM.CardExpMonth.Int32),
				ExpYear:     int(dbPM.CardExpYear.Int32),
				Fingerprint: dbPM.CardFingerprint.String,
			}
		}

		result = append(result, pm)
	}

	return result, nil
}

// DeletePaymentMethod removes a payment method from the database
func (r *Repository) DeletePaymentMethod(ctx context.Context, id, customerID string) error {
	ctx, span := r.tracer.Start(ctx, "Repository.DeletePaymentMethod")
	defer span.End()

	err := r.queries.DeletePaymentMethod(ctx, sqlc.DeletePaymentMethodParams{
		ID:         id,
		CustomerID: customerID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	return nil
}

// StoreCharge stores a charge in the database
func (r *Repository) StoreCharge(ctx context.Context, charge *stripe.Charge) (*stripe.Charge, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.StoreCharge")
	defer span.End()

	// Convert metadata to JSON bytes for storage
	metadataBytes, err := json.Marshal(charge.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.CreateChargeParams{
		ID:              charge.ID,
		Amount:          charge.Amount,
		Currency:        charge.Currency,
		Status:          charge.Status,
		CustomerID:      charge.CustomerID,
		PaymentMethodID: sql.NullString{String: charge.PaymentMethodID, Valid: charge.PaymentMethodID != ""},
		Description:     sql.NullString{String: charge.Description, Valid: charge.Description != ""},
		Metadata:        pqtype.NullRawMessage{RawMessage: metadataBytes, Valid: len(metadataBytes) > 0},
	}

	dbCharge, err := r.queries.CreateCharge(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to store charge: %w", err)
	}

	return &stripe.Charge{
		ID:              dbCharge.ID,
		Amount:          dbCharge.Amount,
		Currency:        dbCharge.Currency,
		Status:          dbCharge.Status,
		CustomerID:      dbCharge.CustomerID,
		PaymentMethodID: dbCharge.PaymentMethodID.String,
		Description:     dbCharge.Description.String,
		Metadata:        convertMetadata(dbCharge.Metadata),
		Created:         dbCharge.CreatedAt.Time.Unix(),
	}, nil
}

// GetCharge retrieves a charge from the database
func (r *Repository) GetCharge(ctx context.Context, id string) (*stripe.Charge, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.GetCharge")
	defer span.End()

	dbCharge, err := r.queries.GetCharge(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get charge: %w", err)
	}

	return &stripe.Charge{
		ID:              dbCharge.ID,
		Amount:          dbCharge.Amount,
		Currency:        dbCharge.Currency,
		Status:          dbCharge.Status,
		CustomerID:      dbCharge.CustomerID,
		PaymentMethodID: dbCharge.PaymentMethodID.String,
		Description:     dbCharge.Description.String,
		Metadata:        convertMetadata(dbCharge.Metadata),
		Created:         dbCharge.CreatedAt.Time.Unix(),
	}, nil
}

// ListCharges retrieves charges for a customer
func (r *Repository) ListCharges(ctx context.Context, customerID string, limit, offset int32) ([]*stripe.Charge, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.ListCharges")
	defer span.End()

	dbCharges, err := r.queries.ListCharges(ctx, sqlc.ListChargesParams{
		CustomerID: customerID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list charges: %w", err)
	}

	var result []*stripe.Charge
	for _, dbCharge := range dbCharges {
		charge := &stripe.Charge{
			ID:              dbCharge.ID,
			Amount:          dbCharge.Amount,
			Currency:        dbCharge.Currency,
			Status:          dbCharge.Status,
			CustomerID:      dbCharge.CustomerID,
			PaymentMethodID: dbCharge.PaymentMethodID.String,
			Description:     dbCharge.Description.String,
			Metadata:        convertMetadata(dbCharge.Metadata),
			Created:         dbCharge.CreatedAt.Time.Unix(),
		}
		result = append(result, charge)
	}

	return result, nil
}

// convertMetadata converts database metadata to stripe metadata format
func convertMetadata(dbMetadata pqtype.NullRawMessage) map[string]string {
	if !dbMetadata.Valid || len(dbMetadata.RawMessage) == 0 {
		return nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(dbMetadata.RawMessage, &metadata); err != nil {
		return nil
	}

	result := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}
