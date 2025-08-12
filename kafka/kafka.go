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
