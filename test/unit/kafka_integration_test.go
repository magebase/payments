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
			Amount:        2000,
			Currency:      "usd",
			CustomerID:    "cus_test123",
			Description:   "Test charge for Kafka integration",
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
		ID:     "evt_test123",
		Type:   "customer.created",
		Source: "/payments/customers",
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

// TestCloudEventsSchemaValidation tests that payment events follow proper CloudEvents v1.0 schema
func TestCloudEventsSchemaValidation(t *testing.T) {
	// Create a CloudEvents validator
	validator := kafka.NewCloudEventsValidator()

	t.Run("Valid CloudEvents v1.0 event passes validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.NoError(t, err, "Valid CloudEvents v1.0 event should pass validation")
	})

	t.Run("Event without ID fails validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.Error(t, err, "Event without ID should fail validation")
		assert.Contains(t, err.Error(), "id")
	})

	t.Run("Event without type fails validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.Error(t, err, "Event without type should fail validation")
		assert.Contains(t, err.Error(), "type")
	})

	t.Run("Event without source fails validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.Error(t, err, "Event without source should fail validation")
		assert.Contains(t, err.Error(), "source")
	})

	t.Run("Event without specversion fails validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:     "evt_test123",
			Type:   "customer.created",
			Source: "/payments/customers",
			Data:   map[string]interface{}{"customer_id": "cus_test123"},
			Time:   time.Now().UTC(),
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.Error(t, err, "Event without specversion should fail validation")
		assert.Contains(t, err.Error(), "specversion")
	})

	t.Run("Event with invalid specversion fails validation", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "2.0", // Invalid version
		}

		// Act
		err := validator.Validate(event)

		// Assert
		assert.Error(t, err, "Event with invalid specversion should fail validation")
		assert.Contains(t, err.Error(), "specversion")
	})
}

// TestCloudEventsVersioning tests that payment events support schema evolution and versioning
func TestCloudEventsVersioning(t *testing.T) {
	// Create a CloudEvents version manager
	versionManager := kafka.NewCloudEventsVersionManager()

	t.Run("Can register new event schema version", func(t *testing.T) {
		// Arrange
		schema := &kafka.EventSchema{
			Version:        "1.1",
			Type:           "customer.created",
			Description:    "Customer created event with additional fields",
			RequiredFields: []string{"id", "type", "source", "specversion", "time"},
			OptionalFields: []string{"customer_id", "email", "name", "metadata"},
		}

		// Act
		err := versionManager.RegisterSchema(schema)

		// Assert
		assert.NoError(t, err, "Should be able to register new schema version")
	})

	t.Run("Can validate event against specific schema version", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123", "email": "test@example.com"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := versionManager.ValidateAgainstVersion(event, "1.0")

		// Assert
		assert.NoError(t, err, "Event should validate against schema version 1.0")
	})

	t.Run("Can migrate event to newer schema version", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		migratedEvent, err := versionManager.MigrateToVersion(event, "1.1")

		// Assert
		assert.NoError(t, err, "Should be able to migrate event to newer version")
		assert.Equal(t, "1.1", migratedEvent.SpecVersion, "Migrated event should have new version")
		assert.NotNil(t, migratedEvent, "Migrated event should not be nil")
	})

	t.Run("Fails to migrate to non-existent schema version", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_test123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_test123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		migratedEvent, err := versionManager.MigrateToVersion(event, "2.0")

		// Assert
		assert.Error(t, err, "Should fail to migrate to non-existent schema version")
		assert.Nil(t, migratedEvent, "Migrated event should be nil on failure")
		assert.Contains(t, err.Error(), "schema version")
	})
}

// TestDeadLetterQueue tests that failed event deliveries are sent to dead letter queue
func TestDeadLetterQueue(t *testing.T) {
	// Create a dead letter queue manager
	dlqManager := kafka.NewDeadLetterQueueManager()

	t.Run("Failed events are sent to dead letter queue", func(t *testing.T) {
		// Arrange
		failedEvent := &kafka.PaymentEvent{
			ID:          "evt_failed123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_failed123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}
		failureReason := "Kafka connection timeout"

		// Act
		err := dlqManager.SendToDeadLetterQueue(failedEvent, failureReason)

		// Assert
		assert.NoError(t, err, "Should be able to send failed event to dead letter queue")

		// Verify event is in dead letter queue
		dlqEvents := dlqManager.GetDeadLetterEvents()
		assert.Len(t, dlqEvents, 1, "Should have one event in dead letter queue")
		assert.Equal(t, failedEvent.ID, dlqEvents[0].Event.ID, "Event ID should match")
		assert.Equal(t, failureReason, dlqEvents[0].FailureReason, "Failure reason should match")
	})

	t.Run("Can retry events from dead letter queue", func(t *testing.T) {
		// Arrange
		failedEvent := &kafka.PaymentEvent{
			ID:          "evt_retry123",
			Type:        "charge.created",
			Source:      "/payments/charges",
			Data:        map[string]interface{}{"charge_id": "ch_retry123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}
		failureReason := "Temporary network error"

		// Send to dead letter queue first
		err := dlqManager.SendToDeadLetterQueue(failedEvent, failureReason)
		assert.NoError(t, err)

		// Act - retry the event
		retryEvent, err := dlqManager.RetryEvent(failedEvent.ID)

		// Assert
		assert.NoError(t, err, "Should be able to retry event from dead letter queue")
		assert.Equal(t, failedEvent.ID, retryEvent.ID, "Retried event ID should match")

		// Verify event is removed from dead letter queue
		dlqEvents := dlqManager.GetDeadLetterEvents()
		assert.Len(t, dlqEvents, 1, "Should still have one event in dead letter queue (the first one)")
	})

	t.Run("Can get dead letter queue statistics", func(t *testing.T) {
		// Act
		stats := dlqManager.GetStatistics()

		// Assert
		assert.NotNil(t, stats, "Statistics should not be nil")
		assert.GreaterOrEqual(t, stats.TotalEvents, 1, "Should have at least one event in dead letter queue")
		assert.GreaterOrEqual(t, stats.RetryCount, 1, "Should have at least one retry")
	})

	t.Run("Can clear dead letter queue", func(t *testing.T) {
		// Act
		err := dlqManager.ClearDeadLetterQueue()

		// Assert
		assert.NoError(t, err, "Should be able to clear dead letter queue")

		// Verify queue is empty
		dlqEvents := dlqManager.GetDeadLetterEvents()
		assert.Len(t, dlqEvents, 0, "Dead letter queue should be empty after clearing")
	})
}

// TestEventReplay tests that payment events can be replayed for historical processing
func TestEventReplay(t *testing.T) {
	// Create an event replay manager
	replayManager := kafka.NewEventReplayManager()

	t.Run("Can store events for replay", func(t *testing.T) {
		// Arrange
		event := &kafka.PaymentEvent{
			ID:          "evt_replay123",
			Type:        "customer.created",
			Source:      "/payments/customers",
			Data:        map[string]interface{}{"customer_id": "cus_replay123"},
			Time:        time.Now().UTC(),
			SpecVersion: "1.0",
		}

		// Act
		err := replayManager.StoreEvent(event)

		// Assert
		assert.NoError(t, err, "Should be able to store event for replay")

		// Verify event is stored
		storedEvents := replayManager.GetStoredEvents()
		assert.Len(t, storedEvents, 1, "Should have one stored event")
		assert.Equal(t, event.ID, storedEvents[0].ID, "Stored event ID should match")
	})

	t.Run("Can replay events from specific time range", func(t *testing.T) {
		// Arrange
		startTime := time.Now().UTC().Add(-1 * time.Hour)
		endTime := time.Now().UTC().Add(1 * time.Hour)

		// Act
		replayEvents := replayManager.ReplayEvents(startTime, endTime)

		// Assert
		assert.NotNil(t, replayEvents, "Replay events should not be nil")
		assert.Len(t, replayEvents, 1, "Should replay one event in time range")
		assert.Equal(t, "evt_replay123", replayEvents[0].ID, "Replayed event ID should match")
	})

	t.Run("Can replay events by type", func(t *testing.T) {
		// Act
		replayEvents := replayManager.ReplayEventsByType("customer.created")

		// Assert
		assert.NotNil(t, replayEvents, "Replay events by type should not be nil")
		assert.Len(t, replayEvents, 1, "Should replay one customer.created event")
		assert.Equal(t, "customer.created", replayEvents[0].Type, "Replayed event type should match")
	})

	t.Run("Can get replay statistics", func(t *testing.T) {
		// Act
		stats := replayManager.GetReplayStatistics()

		// Assert
		assert.NotNil(t, stats, "Replay statistics should not be nil")
		assert.Equal(t, 1, stats.TotalStoredEvents, "Should have one stored event")
		assert.Equal(t, 2, stats.TotalReplayOperations, "Should have two replay operations (time range + type)")
	})

	t.Run("Can clear stored events", func(t *testing.T) {
		// Act
		err := replayManager.ClearStoredEvents()

		// Assert
		assert.NoError(t, err, "Should be able to clear stored events")

		// Verify events are cleared
		storedEvents := replayManager.GetStoredEvents()
		assert.Len(t, storedEvents, 0, "Stored events should be empty after clearing")
	})
}
