package test

import (
	"context"
	"testing"
	"time"

	"apis/payments/kafka"
	"apis/payments/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



// TestKafkaIntegration tests that payment operations publish events to Kafka
func TestKafkaIntegration(t *testing.T) {
	// Create mock Kafka producer
	mockProducer := kafka.NewMockProducer()
	
	// Create payment service with mock gateway and Kafka integration
	paymentService := services.NewPaymentServiceWithMockGateway()
	paymentService.SetEventPublisher(mockProducer)

	t.Run("CreateCustomer publishes customer.created event", func(t *testing.T) {
		// Arrange
		req := &services.CustomerRequest{
			Email:       "test@example.com",
			Name:        "Test Customer",
			Description: "Test customer for Kafka integration",
		}

		// Act
		customer, err := paymentService.CreateCustomer(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, customer)
		
		// Verify event was published
		events := mockProducer.GetEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "customer.created", events[0].Type)
		assert.Equal(t, "/payments/customers", events[0].Source)
		assert.Equal(t, "test@example.com", events[0].Data["email"])
		
		// Clear events for next test
		mockProducer.ClearEvents()
	})

	t.Run("CreateCharge publishes charge.created event", func(t *testing.T) {
		// Arrange
		req := &services.ChargeRequest{
			Amount:       2000,
			Currency:     "usd",
			CustomerID:   "cus_test123",
			Description:  "Test charge for Kafka integration",
			PaymentMethod: "pm_test123",
		}

		// Act
		charge, err := paymentService.CreateCharge(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, charge)
		
		// Verify event was published
		events := mockProducer.GetEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "charge.created", events[0].Type)
		assert.Equal(t, "/payments/charges", events[0].Source)
		assert.Equal(t, int64(2000), events[0].Data["amount"])
		assert.Equal(t, "usd", events[0].Data["currency"])
		
		// Clear events for next test
		mockProducer.ClearEvents()
	})

	t.Run("CreateRefund publishes refund.created event", func(t *testing.T) {
		// Arrange
		req := &services.RefundRequest{
			ChargeID: "ch_test123",
			Amount:   1000,
			Reason:   "requested_by_customer",
		}

		// Act
		refund, err := paymentService.CreateRefund(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, refund)
		
		// Verify event was published
		events := mockProducer.GetEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "refund.created", events[0].Type)
		assert.Equal(t, "/payments/refunds", events[0].Source)
		assert.Equal(t, "ch_test123", events[0].Data["charge_id"])
		assert.Equal(t, int64(1000), events[0].Data["amount"])
		
		// Clear events for next test
		mockProducer.ClearEvents()
	})

	t.Run("CreateDispute publishes dispute.created event", func(t *testing.T) {
		// Arrange
		req := &services.DisputeRequest{
			ChargeID: "ch_test123",
			Amount:   2000,
			Reason:   "fraudulent",
		}

		// Act
		dispute, err := paymentService.CreateDispute(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, dispute)
		
		// Verify event was published
		events := mockProducer.GetEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "dispute.created", events[0].Type)
		assert.Equal(t, "/payments/disputes", events[0].Source)
		assert.Equal(t, "ch_test123", events[0].Data["charge_id"])
		assert.Equal(t, "fraudulent", events[0].Data["reason"])
		
		// Clear events for next test
		mockProducer.ClearEvents()
	})
}

// TestPaymentEventStructure tests that payment events follow CloudEvents specification
func TestPaymentEventStructure(t *testing.T) {
	// Create a sample payment event
	event := &kafka.PaymentEvent{
		ID:          "evt_test123",
		Type:        "customer.created",
		Source:      "/payments/customers",
		Data: map[string]interface{}{
			"customer_id": "cus_test123",
			"email":       "test@example.com",
			"name":        "Test Customer",
		},
		Time: time.Now().UTC(),
	}

	// Verify CloudEvents v1.0 specification compliance
	assert.NotEmpty(t, event.ID, "Event ID is required")
	assert.NotEmpty(t, event.Type, "Event type is required")
	assert.NotEmpty(t, event.Source, "Event source is required")
	assert.NotNil(t, event.Data, "Event data is required")
	assert.NotZero(t, event.Time, "Event time is required")

	// Verify event type follows naming convention
	assert.Regexp(t, `^[a-z]+\.[a-z]+(?:\.(?:created|updated|deleted|succeeded|failed))?$`, event.Type, "Event type should follow naming convention")

	// Verify source follows URI format
	assert.Regexp(t, `^/payments/[a-z-]+$`, event.Source, "Event source should follow URI format")
}

// TestEventPublishingFailure tests that payment operations handle Kafka publishing failures gracefully
func TestEventPublishingFailure(t *testing.T) {
	// Create mock Kafka producer that fails
	mockProducer := kafka.NewMockProducer()
	mockProducer.SetError("customer.created", assert.AnError)

	// Create payment service with mock gateway and failing Kafka producer
	paymentService := services.NewPaymentServiceWithMockGateway()
	paymentService.SetEventPublisher(mockProducer)

	// Test that customer creation still works even if event publishing fails
	req := &services.CustomerRequest{
		Email: "test@example.com",
		Name:  "Test Customer",
	}

	// Act - should not fail due to Kafka publishing failure
	customer, err := paymentService.CreateCustomer(context.Background(), req)

	// Assert - customer creation should succeed even if event publishing fails
	require.NoError(t, err)
	assert.NotNil(t, customer)
	
	// Verify that the error was recorded
	events := mockProducer.GetEvents()
	assert.Len(t, events, 0, "No events should be published when Kafka fails")
}
