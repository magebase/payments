# Payments Service

A modern, scalable payments service built with Go, Go Fiber, and Stripe integration. This service provides comprehensive payment processing capabilities including customer management, payment methods, charges, and more.

## Features

- **Customer Vault**: Create, read, update, and delete customer records
- **Payment Methods**: Add, list, and manage payment methods for customers
- **Charges**: Process payments with comprehensive validation and error handling
- **Refunds**: Process refunds with support for partial refunds and reason tracking
- **RESTful API**: Clean, RESTful endpoints with proper HTTP status codes
- **Validation**: Request validation using go-playground/validator
- **Tracing**: OpenTelemetry integration for observability
- **Middleware**: CORS, logging, and recovery middleware
- **Configuration**: Environment-based configuration management

## Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: Go Fiber v2
- **API Design**: RESTful with JSON
- **Payment Provider**: Stripe
- **Validation**: go-playground/validator
- **Tracing**: OpenTelemetry
- **Testing**: Testify
- **Configuration**: Environment variables

## Architecture

The service follows a clean architecture pattern with:

- **Services Layer**: Business logic for payments, customers, and payment methods
- **API Layer**: HTTP handlers and routing
- **Configuration Layer**: Environment-based configuration management
- **Tracing Layer**: OpenTelemetry integration for observability

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

## API Usage Examples

### Creating a Refund

```bash
curl -X POST http://localhost:8080/api/v1/refunds \
  -H "Content-Type: application/json" \
  -d '{
    "charge_id": "ch_1234567890",
    "amount": 1000,
    "reason": "requested_by_customer",
    "metadata": {
      "note": "Customer requested refund"
    }
  }'
```

### Getting a Refund

```bash
curl http://localhost:8080/api/v1/refunds/re_1234567890
```

### Listing Refunds for a Charge

```bash
curl "http://localhost:8080/api/v1/refunds?charge_id=ch_1234567890"
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Stripe account and API keys
- PostgreSQL (optional, for future database integration)
- Kafka (optional, for future event streaming)

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
- **TRACING_ENABLED**: Enable/disable OpenTelemetry tracing
- **TRACING_ENDPOINT**: OpenTelemetry collector endpoint

## Development

### Project Structure

```
payments/
├── main/           # Application entry point
├── config/         # Configuration management
├── services/       # Business logic services
│   └── stripe/    # Stripe integration
├── test/           # Test files
│   └── unit/      # Unit tests
├── go.mod          # Go module file
├── go.sum          # Go module checksums
├── env.example     # Environment template
└── README.md       # This file
```

### Adding New Features

1. **Write Tests First**: Follow TDD principles - write failing tests first
2. **Implement Feature**: Add the minimal code to make tests pass
3. **Add Validation**: Include proper request validation
4. **Add Tracing**: Include OpenTelemetry spans for observability
5. **Update Documentation**: Keep this README and API docs updated

### Testing Guidelines

- Write unit tests for all business logic
- Use mocks for external dependencies (Stripe API)
- Test both success and failure scenarios
- Ensure high test coverage

## API Examples

### Creating a Customer

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
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
- `500 Internal Server Error`: Unexpected server errors

Error responses include a descriptive error message:

```json
{
  "error": "validation failed: email is required"
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

## Security

- Input validation on all endpoints
- CORS configuration for cross-origin requests
- No sensitive data in logs
- Environment-based configuration for secrets

## Future Enhancements

- [ ] Database integration for persistent storage
- [ ] Kafka integration for event streaming
- [ ] Webhook handling for Stripe events
- [ ] Refunds and disputes API
- [ ] Multi-currency support
- [ ] Rate limiting and throttling
- [ ] Authentication and authorization
- [ ] API versioning
- [ ] Swagger/OpenAPI documentation

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
