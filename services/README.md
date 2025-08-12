# Payment Gateway Abstraction Layer

This package provides a clean abstraction layer for multiple payment gateways, allowing you to switch between different payment providers (Stripe, Paddle, Square) without changing your business logic.

## Features

- **Provider Agnostic**: Write code once, use with any supported payment gateway
- **Easy Switching**: Switch between providers at runtime or via configuration
- **Capability Detection**: Automatically detect what features each provider supports
- **Unified Interface**: Consistent API across all payment providers
- **Extensible**: Easy to add new payment providers

## Supported Providers

### Stripe

- ✅ Full implementation with all features
- ✅ Subscriptions, Connect, Tax, Invoices, Payouts
- ✅ Disputes and Refunds
- ✅ Comprehensive error handling

### Paddle

- 🔄 Placeholder implementation (ready for full integration)
- ✅ Subscriptions, Tax, Invoices
- ❌ Connect, Payouts (not supported by Paddle)

### Square

- 🔄 Placeholder implementation (ready for full integration)
- ✅ Connect, Invoices, Payouts
- ❌ Subscriptions, Tax (not supported by Square)

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "context"
    "log"

    "your-project/services"
)

func main() {
    // Create a payment service with the default provider (Stripe)
    paymentService, err := services.NewPaymentService()
    if err != nil {
        log.Fatal(err)
    }

    // Create a customer
    customer, err := paymentService.CreateCustomer(context.Background(), &services.CustomerRequest{
        Email: "customer@example.com",
        Name:  "John Doe",
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created customer: %s", customer.ID)
}
```

### 2. Switch Providers

```go
// Switch to Paddle
err := paymentService.SwitchProvider(services.ProviderPaddle)
if err != nil {
    log.Printf("Failed to switch to Paddle: %v", err)
}

// Check capabilities
if paymentService.SupportsSubscriptions() {
    log.Println("Current provider supports subscriptions")
}
```

### 3. Provider-Specific Service

```go
// Create a service with a specific provider
paddleService, err := services.NewPaymentServiceWithProvider(services.ProviderPaddle)
if err != nil {
    log.Fatal(err)
}
```

## Configuration

### Environment Variables

```bash
# Payment Gateway Configuration
PAYMENT_DEFAULT_PROVIDER=stripe

# Stripe Configuration (default provider)
STRIPE_ENABLED=true
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret_here

# Paddle Configuration (optional)
PADDLE_ENABLED=false
PADDLE_API_KEY=your_paddle_api_key_here
PADDLE_WEBHOOK_SECRET=your_paddle_webhook_secret_here

# Square Configuration (optional)
SQUARE_ENABLED=false
SQUARE_ACCESS_TOKEN=your_square_access_token_here
SQUARE_WEBHOOK_SIGNATURE_KEY=your_square_webhook_signature_key_here

# Environment
ENVIRONMENT=development
```

### Configuration Priority

1. Environment variables
2. Default values (Stripe enabled by default)
3. Validation ensures at least one provider is configured

## API Reference

### Core Interfaces

#### PaymentGateway

The main interface that all payment providers must implement:

```go
type PaymentGateway interface {
    CustomerGateway
    PaymentMethodGateway
    ChargeGateway
    RefundGateway
    DisputeGateway
    GetProviderName() string
    GetCapabilities() GatewayCapabilities
}
```

#### GatewayCapabilities

Describes what features a payment gateway supports:

```go
type GatewayCapabilities struct {
    SupportsSubscriptions bool
    SupportsConnect       bool
    SupportsTax           bool
    SupportsInvoices      bool
    SupportsPayouts       bool
    SupportsDisputes      bool
    SupportsRefunds       bool
    MaxPaymentAmount      int64
    SupportedCurrencies   []string
    SupportedCountries    []string
}
```

### Common Types

All payment operations use unified request/response types:

```go
type CustomerRequest struct {
    Email       string            `json:"email" validate:"required,email"`
    Name        string            `json:"name" validate:"required,min=1"`
    Phone       string            `json:"phone,omitempty"`
    Description string            `json:"description,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type Customer struct {
    ID          string            `json:"id"`
    Email       string            `json:"email"`
    Name        string            `json:"name"`
    Phone       string            `json:"phone,omitempty"`
    Description string            `json:"description,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    Created     int64             `json:"created"`
    Updated     int64             `json:"updated"`
    ProviderID  string            `json:"provider_id"`
}
```

## Adding New Providers

### 1. Implement the Interface

Create a new file `your_provider_gateway.go`:

```go
package services

import (
    "context"
    "fmt"

    "github.com/go-playground/validator/v10"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

type YourProviderGateway struct {
    config   *ProviderConfig
    validator *validator.Validate
    tracer    trace.Tracer
}

func NewYourProviderGateway(config *ProviderConfig) (*YourProviderGateway, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid YourProvider configuration: %w", err)
    }

    return &YourProviderGateway{
        config:    config,
        validator: validator.New(),
        tracer:    otel.Tracer("payments.yourprovider"),
    }, nil
}

// Implement all required methods...
func (g *YourProviderGateway) GetProviderName() string {
    return "yourprovider"
}

func (g *YourProviderGateway) GetCapabilities() GatewayCapabilities {
    return GatewayCapabilities{
        SupportsSubscriptions: true,
        SupportsConnect:       false,
        // ... set other capabilities
    }
}

// Implement CustomerGateway methods
func (g *YourProviderGateway) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
    // Your implementation here
}

// ... implement all other required methods
```

### 2. Add to Factory

Update `factory.go` to include your new provider:

```go
const (
    ProviderStripe ProviderType = "stripe"
    ProviderPaddle ProviderType = "paddle"
    ProviderSquare ProviderType = "square"
    ProviderYourProvider ProviderType = "yourprovider" // Add this
)

// In CreateGateway method, add:
case ProviderYourProvider:
    return NewYourProviderGateway(&providerConfig)
```

### 3. Add Configuration

Update `factory.go` to load your provider configuration:

```go
// In loadGatewayConfig function, add:
yourProviderEnabled := os.Getenv("YOURPROVIDER_ENABLED") == "true"
if yourProviderEnabled {
    config.Providers[ProviderYourProvider] = ProviderConfig{
        Enabled:       true,
        APIKey:        os.Getenv("YOURPROVIDER_API_KEY"),
        WebhookSecret: os.Getenv("YOURPROVIDER_WEBHOOK_SECRET"),
        Environment:   getEnvironment(),
    }
}
```

### 4. Add Environment Variables

Update `env.example`:

```bash
# YourProvider Configuration
YOURPROVIDER_ENABLED=false
YOURPROVIDER_API_KEY=your_api_key_here
YOURPROVIDER_WEBHOOK_SECRET=your_webhook_secret_here
```

## Testing

### Run All Tests

```bash
go test -v
```

### Run Specific Test

```bash
go test -v -run TestStripeGateway
```

### Test Coverage

```bash
go test -cover
```

## Examples

See `example_usage.go` for comprehensive examples of:

- Basic usage
- Provider switching
- Capability checking
- Error handling

## Best Practices

### 1. Always Check Capabilities

```go
if paymentService.SupportsSubscriptions() {
    // Create subscription
} else {
    // Handle gracefully or show error
}
```

### 2. Validate Requests

```go
err := paymentService.ValidatePaymentRequest(chargeReq)
if err != nil {
    // Handle validation error
    return err
}
```

### 3. Handle Provider-Specific Errors

```go
customer, err := paymentService.CreateCustomer(ctx, req)
if err != nil {
    if strings.Contains(err.Error(), "stripe") {
        // Handle Stripe-specific error
    } else if strings.Contains(err.Error(), "paddle") {
        // Handle Paddle-specific error
    }
    return err
}
```

### 4. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

customer, err := paymentService.CreateCustomer(ctx, req)
```

## Error Handling

The abstraction layer provides consistent error handling across all providers:

- **Validation Errors**: Request validation failures
- **Provider Errors**: Provider-specific API errors
- **Configuration Errors**: Missing or invalid configuration
- **Capability Errors**: Feature not supported by provider

## Performance Considerations

- **Provider Switching**: Minimal overhead when switching providers
- **Capability Checking**: Fast in-memory checks
- **Validation**: Efficient request validation
- **Tracing**: OpenTelemetry integration for observability

## Security

- **API Key Management**: Secure configuration loading
- **Validation**: Input validation for all requests
- **Tracing**: Secure tracing with sensitive data filtering
- **Provider Isolation**: Complete isolation between providers

## Contributing

1. Follow the existing code patterns
2. Add comprehensive tests for new providers
3. Update documentation and examples
4. Ensure all tests pass before submitting

## License

This package is part of the Magebase payments service.
