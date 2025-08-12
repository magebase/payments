# Payments Service

A modern, scalable payments service built with Go, Go Fiber, and Stripe integration. This service provides comprehensive payment processing capabilities including customer management, payment methods, charges, and more, with optional architecture to support different payment gateways.

## Features

- **Customer Vault**: Secure storage and management of customer information
- **Payment Methods**: Add, retrieve, and manage payment methods (cards, SEPA, iDEAL, etc.)
- **Charges**: Process one-time payments with comprehensive error handling
- **Refunds**: Process full or partial refunds with reason tracking
- **Disputes**: Handle chargeback disputes and evidence management
- **Multi-tenant Support**: Isolated data and operations per tenant
- **Event Publishing**: All payment events published to Kafka using Knative CloudEvents
- **Payment Gateway Abstraction**: Support for multiple payment providers (Stripe, Paddle, Square)

## Payment Gateway Abstraction Layer ✅ **COMPLETED**

The payments service now includes a comprehensive abstraction layer that allows you to switch between different payment providers without changing your business logic.

## Kafka Integration with Knative CloudEvents ✅ **COMPLETED**

The payments service now includes comprehensive Kafka integration for event streaming using the CloudEvents v1.0 specification. All payment operations automatically publish events to Kafka for downstream processing and analytics.

### What's Implemented

- **CloudEvents v1.0 Compliance**: All events follow the CloudEvents specification
- **Automatic Event Publishing**: Customer, charge, refund, and dispute operations automatically publish events
- **Kafka Producer**: Reliable message publishing with retry logic and error handling
- **Event Schema**: Structured event data with consistent naming conventions
- **Mock Producer**: Testing support with mock event publisher
- **Environment Configuration**: Kafka brokers and topics configurable via environment variables

### Event Types Published

- **Customer Events**: `customer.created`, `customer.updated`, `customer.deleted`
- **Charge Events**: `charge.created`, `charge.succeeded`, `charge.failed`
- **Refund Events**: `refund.created`, `refund.processed`
- **Dispute Events**: `dispute.created`, `dispute.updated`

### Event Schema

```json
{
  "id": "evt_uuid",
  "type": "customer.created",
  "source": "/payments/customers",
  "data": {
    "customer_id": "cus_123",
    "email": "customer@example.com",
    "name": "John Doe"
  },
  "time": "2025-01-01T00:00:00Z",
  "specversion": "1.0"
}
```

### Configuration

Set the following environment variables to enable Kafka integration:

```bash
export KAFKA_BROKERS=localhost:9092,localhost:9093
export KAFKA_TOPIC=payment-events
```

### Testing

The Kafka integration includes comprehensive testing with mock producers:

```bash
go test ./test/unit/kafka_integration_test.go -v
```

### What's Implemented

- **Abstract Interfaces**: Clean interfaces for all payment operations
- **Provider Factory**: Environment-based provider selection and configuration
- **Stripe Gateway**: Full implementation with all Stripe features
- **Paddle Gateway**: Placeholder implementation ready for full integration
- **Square Gateway**: Placeholder implementation ready for full integration
- **Unified Service Layer**: Single service interface for all providers
- **Capability Detection**: Runtime feature detection and validation
- **Provider Switching**: Dynamic switching between providers at runtime

### Key Benefits

- **Provider Agnostic**: Write code once, use with any supported gateway
- **Easy Migration**: Switch from Stripe to Paddle or Square with minimal code changes
- **Feature Detection**: Automatically detect what each provider supports
- **Consistent API**: Same interface regardless of underlying provider
- **Extensible**: Easy to add new payment providers

### Usage Example

```go
// Create service with default provider (Stripe)
paymentService, err := services.NewPaymentService()

// Switch to Paddle
err = paymentService.SwitchProvider(services.ProviderPaddle)

// Check capabilities
if paymentService.SupportsSubscriptions() {
    // Create subscription
}

// All operations work the same way
customer, err := paymentService.CreateCustomer(ctx, req)
```

See `services/README.md` for comprehensive documentation and examples.

## Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: Go Fiber v2
- **API Design**: RESTful with JSON
- **Payment Provider**: Stripe (with abstraction layer for other providers)
- **Validation**: go-playground/validator
- **Tracing**: OpenTelemetry
- **Testing**: Testify
- **Configuration**: Environment variables
- **Event Streaming**: Kafka with Knative CloudEvents
- **Database**: PostgreSQL with multi-tenant support

## Architecture

The service follows a clean architecture pattern with:

- **Gateway Abstraction Layer**: Abstract interfaces for payment operations
- **Provider Implementations**: Stripe and future payment gateway implementations
- **Services Layer**: Business logic for payments, customers, and payment methods
- **API Layer**: HTTP handlers and routing with tenant isolation
- **Configuration Layer**: Environment-based configuration management
- **Tracing Layer**: OpenTelemetry integration for observability
- **Event Layer**: Kafka integration for event publishing

## API Endpoints

### Health Check

- `GET /health` - Service health status

### Customers

- `POST /api/v1/customers` - Create a new customer
- `GET /api/v1/customers/:id` - Get customer by ID
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

### Payment Methods

- `POST /api/v1/customers/:customerId/payment-methods` - Add payment method
- `GET /api/v1/customers/:customerId/payment-methods` - List payment methods
- `GET /api/v1/customers/:customerId/payment-methods/:id` - Get payment method
- `DELETE /api/v1/customers/:customerId/payment-methods/:id` - Remove payment method

### Charges

- `POST /api/v1/charges` - Create a charge
- `GET /api/v1/charges/:id` - Get charge by ID
- `GET /api/v1/charges` - List charges (with optional customer filter)

### Refunds

- `POST /api/v1/refunds` - Create a refund for a charge
- `GET /api/v1/refunds/:id` - Get refund by ID
- `GET /api/v1/refunds` - List refunds for a specific charge

### Disputes

- `POST /api/v1/disputes` - Create a dispute for a charge
- `GET /api/v1/disputes/:id` - Get dispute by ID
- `GET /api/v1/disputes` - List disputes for a specific charge
- `PUT /api/v1/disputes/:id/status` - Update dispute status

### Webhooks

- `POST /api/v1/webhooks/stripe` - Stripe webhook endpoint

## Planned Features

### Phase 1: Core Infrastructure

- [x] Payment Gateway Abstraction Layer (PAY-003) ✅ **COMPLETED**
- [x] Kafka Integration with Knative CloudEvents (PAY-004) ✅ **COMPLETED**
- [ ] Multi-tenant Architecture and Tenant Isolation (PAY-007)

### Phase 2: Advanced Stripe Features

- [ ] Stripe Subscriptions and Recurring Billing (PAY-005)
- [ ] Stripe Connect and Payouts (PAY-006)
- [ ] Stripe Tax and Invoice Management (PAY-009)

### Phase 3: Enterprise Features

- [ ] Advanced Security and Compliance Features (PAY-008)
- [ ] Advanced Webhook and Event Processing (PAY-010)
- [ ] Advanced Analytics and Reporting (PAY-012)

### Phase 4: Developer Experience

- [ ] OpenAPI v3 Documentation and SDK Generation (PAY-011)
- [ ] Idempotency and Rate Limiting (PAY-013)
- [ ] Testing Framework and CI/CD Pipeline (PAY-014)

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Stripe account and API keys
- PostgreSQL database
- Kafka cluster (for event streaming)
- Knative eventing infrastructure

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd payments
```

2. Install dependencies:

```bash
go mod tidy
```

3. Set up environment variables:

```bash
cp env.example .env
# Edit .env with your actual values
```

4. Set your Stripe API keys:

```bash
export STRIPE_SECRET_KEY=sk_test_your_key_here
export STRIPE_PUBLISHABLE_KEY=pk_test_your_key_here
export STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
```

5. Set up Kafka and Knative:

```bash
# Configure Kafka connection
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=payment-events

# Configure Knative eventing
export KNATIVE_EVENTING_ENABLED=true
```

### Running the Service

#### Development

```bash
go run main/main.go
```

#### Production

```bash
go build -o payments main/main.go
./payments
```

The service will start on port 8080 by default (configurable via PORT environment variable).

### Running Tests

```bash
# Run all tests
go test ./test/...

# Run specific test files
go test ./test/unit/customer_vault_test.go -v
go test ./test/unit/stripe_charges_test.go -v

# Run with coverage
go test ./test/... -cover
```

## Configuration

The service uses environment variables for configuration. See `env.example` for all available options.

### Key Configuration Options

- **PORT**: Server port (default: 8080)
- **STRIPE_SECRET_KEY**: Your Stripe secret key
- **STRIPE_PUBLISHABLE_KEY**: Your Stripe publishable key
- **STRIPE_WEBHOOK_SECRET**: Your Stripe webhook secret
- **TRACING_ENABLED**: Enable/disable OpenTelemetry tracing
- **TRACING_ENDPOINT**: OpenTelemetry collector endpoint
- **KAFKA_BROKERS**: Kafka broker addresses
- **KAFKA_TOPIC**: Kafka topic for payment events
- **KNATIVE_EVENTING_ENABLED**: Enable Knative eventing

## Development

### Project Structure

```
payments/
├── main/           # Application entry point
├── config/         # Configuration management
├── services/       # Business logic services
│   └── stripe/    # Stripe integration
├── db/            # Database layer
│   ├── migrations/ # Database migrations
│   └── sqlc/      # Generated SQL code
├── test/          # Test files
│   └── unit/      # Unit tests
├── go.mod         # Go module file
├── go.sum         # Go module checksums
├── env.example    # Environment template
├── knative.yaml   # Knative configuration
└── README.md      # This file
```

### Adding New Features

1. **Write Tests First**: Follow TDD principles - write failing tests first
2. **Implement Feature**: Add the minimal code to make tests pass
3. **Add Validation**: Include proper request validation
4. **Add Tracing**: Include OpenTelemetry spans for observability
5. **Add Event Publishing**: Publish events to Kafka for downstream processing
6. **Update Documentation**: Keep this README and API docs updated

### Testing Guidelines

- Write unit tests for all business logic
- Use mocks for external dependencies (Stripe API)
- Test both success and failure scenarios
- Ensure high test coverage (target: 90%+)
- Test tenant isolation for multi-tenant features

## API Examples

### Creating a Customer

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant123" \
  -d '{
    "email": "customer@example.com",
    "name": "John Doe",
    "phone": "+1234567890",
    "description": "Test customer"
  }'
```

### Adding a Payment Method

```bash
curl -X POST http://localhost:8080/api/v1/customers/cus_123/payment-methods \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant123" \
  -d '{
    "type": "card",
    "card": {
      "token": "tok_visa"
    }
  }'
```

### Creating a Charge

```bash
curl -X POST http://localhost:8080/api/v1/charges \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant123" \
  -d '{
    "amount": 2000,
    "currency": "usd",
    "customer_id": "cus_123",
    "description": "Test charge",
    "source": "pm_123"
  }'
```

## Error Handling

The service returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Invalid request data or validation errors
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Unexpected server errors

Error responses include a descriptive error message:

```json
{
  "error": "validation failed: email is required",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "email",
    "message": "email is required"
  }
}
```

## Observability

### Tracing

The service integrates with OpenTelemetry for distributed tracing. When enabled, all operations create spans that can be collected and visualized in tools like Jaeger or Zipkin.

### Logging

Structured logging is provided via Go Fiber's logger middleware, including:

- Request/response logging
- Error logging
- Performance metrics
- Tenant context information

### Event Publishing

All payment operations publish events to Kafka using CloudEvents format:

- Customer events (created, updated, deleted)
- Payment method events (added, removed)
- Charge events (created, succeeded, failed)
- Refund events (created, processed)
- Dispute events (created, updated)

## Security

- Input validation on all endpoints
- CORS configuration for cross-origin requests
- No sensitive data in logs
- Environment-based configuration for secrets
- Tenant isolation and data partitioning
- Rate limiting and abuse prevention
- PCI DSS compliance measures

## Multi-tenant Support

The service supports multiple tenants with complete data isolation:

- **Tenant Identification**: Via JWT claims, API keys, or headers
- **Data Isolation**: Database-level tenant separation
- **Rate Limiting**: Per-tenant quotas and limits
- **Configuration**: Tenant-specific settings and branding
- **Audit Logging**: Tenant-aware audit trails

## Future Enhancements

- [ ] Support for additional payment gateways (Paddle, Square, etc.)
- [ ] Advanced fraud detection and prevention
- [ ] Machine learning-based risk scoring
- [ ] Automated compliance reporting
- [ ] Real-time analytics and dashboards
- [ ] Advanced webhook management
- [ ] Subscription and recurring billing
- [ ] Marketplace payment support
- [ ] Tax calculation and reporting
- [ ] Advanced invoice management

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow TDD principles
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license information here]

## Support

For support and questions, please [create an issue](link-to-issues) or contact the development team.
