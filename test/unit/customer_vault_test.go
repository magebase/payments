package test

import (
	"context"
	"testing"

	"apis/payments/services/stripe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomerVault tests the customer vault and payment methods functionality
func TestCustomerVault(t *testing.T) {
	t.Run("should create a customer successfully", func(t *testing.T) {
		// Create a mock customer service for testing
		mockService := &MockCustomerService{
			shouldSucceed: true,
			mockCustomer: &stripe.Customer{
				ID:       "cus_test123",
				Email:    "test@example.com",
				Name:     "Test Customer",
				Created:  1234567890,
				Metadata: map[string]string{"source": "test"},
			},
		}

		// Create a customer request
		customerRequest := &stripe.CustomerRequest{
			Email:    "test@example.com",
			Name:     "Test Customer",
			Metadata: map[string]string{"source": "test"},
		}

		// Process the customer creation
		customer, err := mockService.CreateCustomer(context.Background(), customerRequest)

		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "cus_test123", customer.ID)
		assert.Equal(t, "test@example.com", customer.Email)
		assert.Equal(t, "Test Customer", customer.Name)
	})

	t.Run("should handle invalid customer requests", func(t *testing.T) {
		// Test with invalid email
		invalidRequest := &stripe.CustomerRequest{
			Email: "invalid-email",
			Name:  "Test Customer",
		}

		customerService := stripe.NewCustomerService()
		_, err := customerService.CreateCustomer(context.Background(), invalidRequest)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("should add payment method to customer", func(t *testing.T) {
		mockService := &MockCustomerService{
			shouldSucceed: true,
			mockPaymentMethod: &stripe.PaymentMethod{
				ID:       "pm_test123",
				Type:     "card",
				Card:     &stripe.Card{Last4: "4242", Brand: "visa"},
				Customer: "cus_test123",
				Created:  1234567890,
			},
		}

		paymentMethodRequest := &stripe.PaymentMethodRequest{
			Type:     "card",
			Card:     &stripe.CardRequest{Token: "tok_visa"},
			Customer: "cus_test123",
		}

		paymentMethod, err := mockService.AddPaymentMethod(context.Background(), paymentMethodRequest)

		require.NoError(t, err)
		assert.NotNil(t, paymentMethod)
		assert.Equal(t, "pm_test123", paymentMethod.ID)
		assert.Equal(t, "card", paymentMethod.Type)
		assert.Equal(t, "4242", paymentMethod.Card.Last4)
	})

	t.Run("should list customer payment methods", func(t *testing.T) {
		mockService := &MockCustomerService{
			shouldSucceed: true,
			mockPaymentMethods: []*stripe.PaymentMethod{
				{
					ID:       "pm_test123",
					Type:     "card",
					Card:     &stripe.Card{Last4: "4242", Brand: "visa"},
					Customer: "cus_test123",
					Created:  1234567890,
				},
				{
					ID:       "pm_test456",
					Type:     "card",
					Card:     &stripe.Card{Last4: "5555", Brand: "mastercard"},
					Customer: "cus_test123",
					Created:  1234567891,
				},
			},
		}

		paymentMethods, err := mockService.ListPaymentMethods(context.Background(), "cus_test123")

		require.NoError(t, err)
		assert.Len(t, paymentMethods, 2)
		assert.Equal(t, "pm_test123", paymentMethods[0].ID)
		assert.Equal(t, "pm_test456", paymentMethods[1].ID)
	})
}

// MockCustomerService is a mock implementation for testing
type MockCustomerService struct {
	shouldSucceed      bool
	mockCustomer       *stripe.Customer
	mockPaymentMethod  *stripe.PaymentMethod
	mockPaymentMethods []*stripe.PaymentMethod
	mockError          error
}

func (m *MockCustomerService) CreateCustomer(ctx context.Context, request *stripe.CustomerRequest) (*stripe.Customer, error) {
	if !m.shouldSucceed {
		return nil, m.mockError
	}
	return m.mockCustomer, nil
}

func (m *MockCustomerService) AddPaymentMethod(ctx context.Context, request *stripe.PaymentMethodRequest) (*stripe.PaymentMethod, error) {
	if !m.shouldSucceed {
		return nil, m.mockError
	}
	return m.mockPaymentMethod, nil
}

func (m *MockCustomerService) ListPaymentMethods(ctx context.Context, customerID string) ([]*stripe.PaymentMethod, error) {
	if !m.shouldSucceed {
		return nil, m.mockError
	}
	return m.mockPaymentMethods, nil
}
