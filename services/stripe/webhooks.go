package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// WebhookService handles Stripe webhook events
type WebhookService struct {
	validator     *validator.Validate
	webhookSecret string
	tracer        trace.Tracer
}

// NewWebhookService creates a new webhook service
func NewWebhookService(webhookSecret string) *WebhookService {
	return &WebhookService{
		validator:     validator.New(),
		webhookSecret: webhookSecret,
		tracer:        otel.Tracer("payments.webhook"),
	}
}

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Created int64                  `json:"created"`
	Data    map[string]interface{} `json:"data"`
}

// WebhookRequest represents a webhook request
type WebhookRequest struct {
	Signature string `json:"-" header:"Stripe-Signature"`
	Body      []byte `json:"-"`
}

// ProcessWebhook processes a Stripe webhook event
func (s *WebhookService) ProcessWebhook(ctx context.Context, request *WebhookRequest) error {
	ctx, span := s.tracer.Start(ctx, "ProcessWebhook")
	defer span.End()

	// Verify webhook signature
	if err := s.verifySignature(request.Signature, request.Body); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Parse the webhook event
	var event WebhookEvent
	if err := json.Unmarshal(request.Body, &event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("failed to parse webhook event: %w", err)
	}

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", event.Type),
		attribute.Int64("event.created", event.Created),
	)

	// Handle the event based on its type
	if err := s.handleEvent(ctx, &event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("failed to handle event %s: %w", event.Type, err)
	}

	return nil
}

// verifySignature verifies the Stripe webhook signature
func (s *WebhookService) verifySignature(signature string, payload []byte) error {
	if signature == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	// Parse the signature header
	parts := strings.Split(signature, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid signature format")
	}

	var timestamp, sig string
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			sig = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || sig == "" {
		return fmt.Errorf("missing timestamp or signature in header")
	}

	// Check if timestamp is within tolerance (5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	now := time.Now().Unix()
	if now-ts > 300 { // 5 minutes
		return fmt.Errorf("webhook timestamp too old")
	}

	// Create expected signature
	message := fmt.Sprintf("%s.%s", timestamp, string(payload))
	expectedSig := hmac.New(sha256.New, []byte(s.webhookSecret))
	expectedSig.Write([]byte(message))
	expectedSigHex := hex.EncodeToString(expectedSig.Sum(nil))

	// Compare signatures
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedSigHex)) != 1 {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// handleEvent routes the webhook event to the appropriate handler
func (s *WebhookService) handleEvent(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleEvent")
	defer span.End()

	span.SetAttributes(attribute.String("event.type", event.Type))

	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(ctx, event)
	case "charge.succeeded":
		return s.handleChargeSucceeded(ctx, event)
	case "charge.failed":
		return s.handleChargeFailed(ctx, event)
	case "charge.refunded":
		return s.handleChargeRefunded(ctx, event)
	case "charge.dispute.created":
		return s.handleDisputeCreated(ctx, event)
	case "charge.dispute.closed":
		return s.handleDisputeClosed(ctx, event)
	default:
		// Log unknown event types but don't fail
		span.SetAttributes(attribute.String("warning", "unknown event type"))
		return nil
	}
}

// handlePaymentIntentSucceeded handles successful payment intents
func (s *WebhookService) handlePaymentIntentSucceeded(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handlePaymentIntentSucceeded")
	defer span.End()

	// TODO: Implement payment intent success logic
	// This could include updating order status, sending confirmation emails, etc.
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handlePaymentIntentFailed handles failed payment intents
func (s *WebhookService) handlePaymentIntentFailed(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handlePaymentIntentFailed")
	defer span.End()

	// TODO: Implement payment intent failure logic
	// This could include updating order status, sending failure notifications, etc.
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handleChargeSucceeded handles successful charges
func (s *WebhookService) handleChargeSucceeded(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleChargeSucceeded")
	defer span.End()

	// TODO: Implement charge success logic
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handleChargeFailed handles failed charges
func (s *WebhookService) handleChargeFailed(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleChargeFailed")
	defer span.End()

	// TODO: Implement charge failure logic
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handleChargeRefunded handles refunded charges
func (s *WebhookService) handleChargeRefunded(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleChargeRefunded")
	defer span.End()

	// TODO: Implement charge refund logic
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handleDisputeCreated handles created disputes
func (s *WebhookService) handleDisputeCreated(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleDisputeCreated")
	defer span.End()

	// TODO: Implement dispute creation logic
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// handleDisputeClosed handles closed disputes
func (s *WebhookService) handleDisputeClosed(ctx context.Context, event *WebhookEvent) error {
	ctx, span := s.tracer.Start(ctx, "handleDisputeClosed")
	defer span.End()

	// TODO: Implement dispute closure logic
	span.SetAttributes(attribute.String("status", "processed"))
	return nil
}

// ValidateWebhookRequest validates a webhook request
func (s *WebhookService) ValidateWebhookRequest(request *WebhookRequest) error {
	if err := s.validator.Struct(request); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if request.Signature == "" {
		return fmt.Errorf("Stripe-Signature header is required")
	}

	if len(request.Body) == 0 {
		return fmt.Errorf("webhook body cannot be empty")
	}

	return nil
}

// ParseWebhookRequest parses a webhook request from a Fiber context
func (s *WebhookService) ParseWebhookRequest(c *fiber.Ctx) (*WebhookRequest, error) {
	body := c.Body()
	if len(body) == 0 {
		return nil, fmt.Errorf("webhook body cannot be empty")
	}

	signature := c.Get("Stripe-Signature")
	if signature == "" {
		return nil, fmt.Errorf("missing Stripe-Signature header")
	}

	return &WebhookRequest{
		Signature: signature,
		Body:      body,
	}, nil
}
