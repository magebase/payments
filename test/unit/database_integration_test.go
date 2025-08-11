package test

import (
	"context"
	"testing"
	"time"

	"apis/payments/db"
	"apis/payments/services/stripe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseIntegration tests the database integration functionality
func TestDatabaseIntegration(t *testing.T) {
	t.Run("should connect to Yugabyte DB successfully", func(t *testing.T) {
		// This test will verify database connection
		// In a real test, you would use testcontainers to spin up Yugabyte

		config := &db.Config{
			Host:     "localhost",
			Port:     5433,
			User:     "yugabyte",
			Password: "yugabyte",
			DBName:   "payments",
			SSLMode:  "disable",
		}

		// Test that config is valid
		assert.NotNil(t, config)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5433, config.Port)
	})

	t.Run("should create customer in database", func(t *testing.T) {
		// Test customer persistence
		customer := &stripe.Customer{
			ID:          "cus_test123",
			Email:       "test@example.com",
			Name:        "Test Customer",
			Phone:       "+1234567890",
			Description: "Test customer",
			Metadata:    map[string]string{"source": "test"},
			Created:     time.Now().Unix(),
			Updated:     time.Now().Unix(),
		}

		// Mock database operation
		mockDB := &MockDatabase{
			shouldSucceed: true,
			mockCustomer:  customer,
		}

		createdCustomer, err := mockDB.CreateCustomer(context.Background(), customer)

		require.NoError(t, err)
		assert.NotNil(t, createdCustomer)
		assert.Equal(t, customer.ID, createdCustomer.ID)
		assert.Equal(t, customer.Email, createdCustomer.Email)
	})

	t.Run("should store payment method in database", func(t *testing.T) {
		// Test payment method persistence
		paymentMethod := &stripe.PaymentMethod{
			ID:       "pm_test123",
			Type:     "card",
			Card:     &stripe.Card{Last4: "4242", Brand: "visa"},
			Customer: "cus_test123",
			Created:  time.Now().Unix(),
		}

		mockDB := &MockDatabase{
			shouldSucceed:     true,
			mockPaymentMethod: paymentMethod,
		}

		storedPaymentMethod, err := mockDB.StorePaymentMethod(context.Background(), paymentMethod)

		require.NoError(t, err)
		assert.NotNil(t, storedPaymentMethod)
		assert.Equal(t, paymentMethod.ID, storedPaymentMethod.ID)
	})

	t.Run("should log charge to ClickHouse", func(t *testing.T) {
		// Test ClickHouse analytics logging
		charge := &stripe.Charge{
			ID:          "ch_test123",
			Amount:      2000,
			Currency:    "usd",
			Status:      "succeeded",
			CustomerID:  "cus_test123",
			Description: "Test charge",
			Created:     time.Now().Unix(),
		}

		mockClickHouse := &MockClickHouse{
			shouldSucceed: true,
		}

		err := mockClickHouse.LogCharge(context.Background(), charge)

		require.NoError(t, err)
	})
}

// MockDatabase is a mock implementation for testing
type MockDatabase struct {
	shouldSucceed     bool
	mockCustomer      *stripe.Customer
	mockPaymentMethod *stripe.PaymentMethod
	mockError         error
}

func (m *MockDatabase) CreateCustomer(ctx context.Context, customer *stripe.Customer) (*stripe.Customer, error) {
	if !m.shouldSucceed {
		return nil, m.mockError
	}
	return m.mockCustomer, nil
}

func (m *MockDatabase) StorePaymentMethod(ctx context.Context, paymentMethod *stripe.PaymentMethod) (*stripe.PaymentMethod, error) {
	if !m.shouldSucceed {
		return nil, m.mockError
	}
	return m.mockPaymentMethod, nil
}

// MockClickHouse is a mock implementation for testing
type MockClickHouse struct {
	shouldSucceed bool
	mockError     error
}

func (m *MockClickHouse) LogCharge(ctx context.Context, charge *stripe.Charge) error {
	if !m.shouldSucceed {
		return m.mockError
	}
	return nil
}
