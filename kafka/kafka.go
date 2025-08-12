package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// PaymentEvent represents a payment event in CloudEvents format
type PaymentEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Time        time.Time              `json:"time"`
	SpecVersion string                 `json:"specversion"`
}

// EventPublisher defines the interface for publishing payment events
type EventPublisher interface {
	Publish(topic string, event *PaymentEvent) error
}

// KafkaProducer handles publishing events to Kafka
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

// Publish publishes a payment event to Kafka
func (k *KafkaProducer) Publish(topic string, event *PaymentEvent) error {
	// Ensure event has required fields
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	if event.SpecVersion == "" {
		event.SpecVersion = "1.0"
	}

	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(eventData),
		Headers: []sarama.RecordHeader{
			{Key: []byte("ce-specversion"), Value: []byte(event.SpecVersion)},
			{Key: []byte("ce-type"), Value: []byte(event.Type)},
			{Key: []byte("ce-source"), Value: []byte(event.Source)},
			{Key: []byte("ce-id"), Value: []byte(event.ID)},
			{Key: []byte("ce-time"), Value: []byte(event.Time.Format(time.RFC3339))},
		},
	}

	// Publish message
	partition, offset, err := k.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published event %s to topic %s (partition: %d, offset: %d)",
		event.ID, topic, partition, offset)

	return nil
}

// Close closes the Kafka producer
func (k *KafkaProducer) Close() error {
	return k.producer.Close()
}

// MockProducer is a mock implementation for testing
type MockProducer struct {
	events []*PaymentEvent
	errors map[string]error
}

// NewMockProducer creates a new mock producer for testing
func NewMockProducer() *MockProducer {
	return &MockProducer{
		events: make([]*PaymentEvent, 0),
		errors: make(map[string]error),
	}
}

// Publish records the event for testing purposes
func (m *MockProducer) Publish(topic string, event *PaymentEvent) error {
	// Check if we should return an error for this event type
	if err, exists := m.errors[event.Type]; exists {
		return err
	}

	m.events = append(m.events, event)
	return nil
}

// GetEvents returns all published events
func (m *MockProducer) GetEvents() []*PaymentEvent {
	return m.events
}

// SetError sets an error to be returned for a specific event type
func (m *MockProducer) SetError(eventType string, err error) {
	m.errors[eventType] = err
}

// ClearEvents clears all recorded events
func (m *MockProducer) ClearEvents() {
	m.events = make([]*PaymentEvent, 0)
}

// ClearErrors clears all error configurations
func (m *MockProducer) ClearErrors() {
	m.errors = make(map[string]error)
}

// CloudEventsValidator validates that events follow CloudEvents v1.0 specification
type CloudEventsValidator struct{}

// NewCloudEventsValidator creates a new CloudEvents validator
func NewCloudEventsValidator() *CloudEventsValidator {
	return &CloudEventsValidator{}
}

// Validate validates that a payment event follows CloudEvents specification
func (v *CloudEventsValidator) Validate(event *PaymentEvent) error {
	// Check required fields according to CloudEvents spec
	if event.ID == "" {
		return fmt.Errorf("event id is required")
	}
	if event.Type == "" {
		return fmt.Errorf("event type is required")
	}
	if event.Source == "" {
		return fmt.Errorf("event source is required")
	}
	if event.SpecVersion == "" {
		return fmt.Errorf("event specversion is required")
	}
	
	// Validate specversion is supported (1.0 or 1.1)
	if event.SpecVersion != "1.0" && event.SpecVersion != "1.1" {
		return fmt.Errorf("event specversion must be '1.0' or '1.1', got '%s'", event.SpecVersion)
	}
	
	// Validate time is not zero (optional but should be set)
	if event.Time.IsZero() {
		return fmt.Errorf("event time should be set")
	}
	
	return nil
}

// EventSchema defines the schema for a specific event type and version
type EventSchema struct {
	Version        string   `json:"version"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	RequiredFields []string `json:"required_fields"`
	OptionalFields []string `json:"optional_fields"`
}

// CloudEventsVersionManager manages event schema versions and provides migration capabilities
type CloudEventsVersionManager struct {
	schemas map[string]*EventSchema // key: "type:version"
}

// NewCloudEventsVersionManager creates a new CloudEvents version manager
func NewCloudEventsVersionManager() *CloudEventsVersionManager {
	vm := &CloudEventsVersionManager{
		schemas: make(map[string]*EventSchema),
	}
	
	// Register default schemas
	vm.registerDefaultSchemas()
	
	return vm
}

// registerDefaultSchemas registers the default event schemas
func (vm *CloudEventsVersionManager) registerDefaultSchemas() {
	// Default customer.created schema v1.0
	customerV1 := &EventSchema{
		Version:        "1.0",
		Type:           "customer.created",
		Description:    "Customer created event v1.0",
		RequiredFields: []string{"id", "type", "source", "specversion", "time"},
		OptionalFields: []string{"customer_id", "email", "name"},
	}
	vm.RegisterSchema(customerV1)
	
	// Default customer.created schema v1.1
	customerV11 := &EventSchema{
		Version:        "1.1",
		Type:           "customer.created",
		Description:    "Customer created event v1.1 with additional fields",
		RequiredFields: []string{"id", "type", "source", "specversion", "time", "customer_id", "email"},
		OptionalFields: []string{"name", "metadata"},
	}
	vm.RegisterSchema(customerV11)
}

// RegisterSchema registers a new event schema version
func (vm *CloudEventsVersionManager) RegisterSchema(schema *EventSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if schema.Version == "" {
		return fmt.Errorf("schema version is required")
	}
	if schema.Type == "" {
		return fmt.Errorf("schema type is required")
	}
	
	key := fmt.Sprintf("%s:%s", schema.Type, schema.Version)
	vm.schemas[key] = schema
	return nil
}

// ValidateAgainstVersion validates an event against a specific schema version
func (vm *CloudEventsVersionManager) ValidateAgainstVersion(event *PaymentEvent, version string) error {
	key := fmt.Sprintf("%s:%s", event.Type, version)
	schema, exists := vm.schemas[key]
	if !exists {
		return fmt.Errorf("schema version %s for type %s not found", version, event.Type)
	}
	
	// Basic CloudEvents validation first
	validator := NewCloudEventsValidator()
	if err := validator.Validate(event); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}
	
	// Validate required fields are present in data
	for _, field := range schema.RequiredFields {
		if field == "id" || field == "type" || field == "source" || field == "specversion" || field == "time" {
			continue // These are handled by basic validation
		}
		if _, exists := event.Data[field]; !exists {
			return fmt.Errorf("required field '%s' not found in event data", field)
		}
	}
	
	return nil
}

// MigrateToVersion migrates an event to a newer schema version
func (vm *CloudEventsVersionManager) MigrateToVersion(event *PaymentEvent, targetVersion string) (*PaymentEvent, error) {
	key := fmt.Sprintf("%s:%s", event.Type, targetVersion)
	schema, exists := vm.schemas[key]
	if !exists {
		return nil, fmt.Errorf("target schema version %s for type %s not found", targetVersion, event.Type)
	}
	
	// Create a copy of the event
	migratedEvent := &PaymentEvent{
		ID:          event.ID,
		Type:        event.Type,
		Source:      event.Source,
		Data:        make(map[string]interface{}),
		Time:        event.Time,
		SpecVersion: targetVersion,
	}
	
	// Copy data fields
	for k, v := range event.Data {
		migratedEvent.Data[k] = v
	}
	
	// Validate the migrated event against the target schema
	if err := vm.ValidateAgainstVersion(migratedEvent, targetVersion); err != nil {
		return nil, fmt.Errorf("migrated event validation failed against schema %s: %w", schema.Version, err)
	}
	
	return migratedEvent, nil
}

// DeadLetterEvent represents a failed event in the dead letter queue
type DeadLetterEvent struct {
	Event         *PaymentEvent `json:"event"`
	FailureReason string        `json:"failure_reason"`
	FailureTime   time.Time     `json:"failure_time"`
	RetryCount    int           `json:"retry_count"`
}

// DeadLetterQueueStatistics provides metrics about the dead letter queue
type DeadLetterQueueStatistics struct {
	TotalEvents int `json:"total_events"`
	RetryCount  int `json:"retry_count"`
}

// DeadLetterQueueManager manages failed event deliveries and provides retry capabilities
type DeadLetterQueueManager struct {
	events      []*DeadLetterEvent
	totalRetries int
}

// NewDeadLetterQueueManager creates a new dead letter queue manager
func NewDeadLetterQueueManager() *DeadLetterQueueManager {
	return &DeadLetterQueueManager{
		events: make([]*DeadLetterEvent, 0),
	}
}

// SendToDeadLetterQueue sends a failed event to the dead letter queue
func (dlq *DeadLetterQueueManager) SendToDeadLetterQueue(event *PaymentEvent, failureReason string) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	if failureReason == "" {
		return fmt.Errorf("failure reason is required")
	}
	
	dlqEvent := &DeadLetterEvent{
		Event:         event,
		FailureReason: failureReason,
		FailureTime:   time.Now().UTC(),
		RetryCount:    0,
	}
	
	dlq.events = append(dlq.events, dlqEvent)
	return nil
}

// GetDeadLetterEvents returns all events in the dead letter queue
func (dlq *DeadLetterQueueManager) GetDeadLetterEvents() []*DeadLetterEvent {
	return dlq.events
}

// RetryEvent retrieves an event from the dead letter queue for retry
func (dlq *DeadLetterQueueManager) RetryEvent(eventID string) (*PaymentEvent, error) {
	for i, dlqEvent := range dlq.events {
		if dlqEvent.Event.ID == eventID {
			// Increment retry count
			dlqEvent.RetryCount++
			
			// Track total retries
			dlq.totalRetries++
			
			// Remove the event from the dead letter queue
			dlq.events = append(dlq.events[:i], dlq.events[i+1:]...)
			
			// Return the event for retry
			return dlqEvent.Event, nil
		}
	}
	
	return nil, fmt.Errorf("event with ID %s not found in dead letter queue", eventID)
}

// GetStatistics returns statistics about the dead letter queue
func (dlq *DeadLetterQueueManager) GetStatistics() *DeadLetterQueueStatistics {
	totalEvents := len(dlq.events)
	
	return &DeadLetterQueueStatistics{
		TotalEvents: totalEvents,
		RetryCount:  dlq.totalRetries,
	}
}

// ClearDeadLetterQueue removes all events from the dead letter queue
func (dlq *DeadLetterQueueManager) ClearDeadLetterQueue() error {
	dlq.events = make([]*DeadLetterEvent, 0)
	return nil
}

// ReplayStatistics provides metrics about event replay operations
type ReplayStatistics struct {
	TotalStoredEvents    int `json:"total_stored_events"`
	TotalReplayOperations int `json:"total_replay_operations"`
}

// EventReplayManager manages event storage and replay capabilities
type EventReplayManager struct {
	events []*PaymentEvent
	replayCount int
}

// NewEventReplayManager creates a new event replay manager
func NewEventReplayManager() *EventReplayManager {
	return &EventReplayManager{
		events: make([]*PaymentEvent, 0),
		replayCount: 0,
	}
}

// StoreEvent stores an event for future replay
func (erm *EventReplayManager) StoreEvent(event *PaymentEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	
	erm.events = append(erm.events, event)
	return nil
}

// GetStoredEvents returns all stored events
func (erm *EventReplayManager) GetStoredEvents() []*PaymentEvent {
	return erm.events
}

// ReplayEvents replays events within a specific time range
func (erm *EventReplayManager) ReplayEvents(startTime, endTime time.Time) []*PaymentEvent {
	var replayEvents []*PaymentEvent
	
	for _, event := range erm.events {
		if event.Time.After(startTime) && event.Time.Before(endTime) {
			replayEvents = append(replayEvents, event)
		}
	}
	
	erm.replayCount++
	return replayEvents
}

// ReplayEventsByType replays events of a specific type
func (erm *EventReplayManager) ReplayEventsByType(eventType string) []*PaymentEvent {
	var replayEvents []*PaymentEvent
	
	for _, event := range erm.events {
		if event.Type == eventType {
			replayEvents = append(replayEvents, event)
		}
	}
	
	erm.replayCount++
	return replayEvents
}

// GetReplayStatistics returns statistics about replay operations
func (erm *EventReplayManager) GetReplayStatistics() *ReplayStatistics {
	return &ReplayStatistics{
		TotalStoredEvents:     len(erm.events),
		TotalReplayOperations: erm.replayCount,
	}
}

// ClearStoredEvents removes all stored events
func (erm *EventReplayManager) ClearStoredEvents() error {
	erm.events = make([]*PaymentEvent, 0)
	return nil
}
