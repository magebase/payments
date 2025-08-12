package services

import (
	"fmt"
	"os"
	"strings"
)

// ProviderType represents the type of payment provider
type ProviderType string

const (
	ProviderStripe ProviderType = "stripe"
	ProviderPaddle ProviderType = "paddle"
	ProviderSquare ProviderType = "square"
)

// GatewayFactory creates payment gateway instances based on configuration
type GatewayFactory struct {
	config *GatewayConfig
}

// GatewayConfig holds configuration for payment gateways
type GatewayConfig struct {
	DefaultProvider ProviderType
	Providers       map[ProviderType]ProviderConfig
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Enabled       bool
	APIKey        string
	WebhookSecret string
	Environment   string // "test" or "live"
}

// NewGatewayFactory creates a new gateway factory with configuration
func NewGatewayFactory() (*GatewayFactory, error) {
	config, err := loadGatewayConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load gateway config: %w", err)
	}

	return &GatewayFactory{
		config: config,
	}, nil
}

// CreateGateway creates a payment gateway instance for the specified provider
func (f *GatewayFactory) CreateGateway(providerType ProviderType) (PaymentGateway, error) {
	providerConfig, exists := f.config.Providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured", providerType)
	}

	if !providerConfig.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", providerType)
	}

	switch providerType {
	case ProviderStripe:
		return NewStripeGateway(&providerConfig)
	case ProviderPaddle:
		return NewPaddleGateway(&providerConfig)
	case ProviderSquare:
		return NewSquareGateway(&providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// CreateDefaultGateway creates a payment gateway instance using the default provider
func (f *GatewayFactory) CreateDefaultGateway() (PaymentGateway, error) {
	return f.CreateGateway(f.config.DefaultProvider)
}

// GetAvailableProviders returns a list of available and enabled providers
func (f *GatewayFactory) GetAvailableProviders() []ProviderType {
	var providers []ProviderType
	for providerType, config := range f.config.Providers {
		if config.Enabled {
			providers = append(providers, providerType)
		}
	}
	return providers
}

// loadGatewayConfig loads gateway configuration from environment variables
func loadGatewayConfig() (*GatewayConfig, error) {
	config := &GatewayConfig{
		Providers: make(map[ProviderType]ProviderConfig),
	}

	// Default provider
	defaultProvider := os.Getenv("PAYMENT_DEFAULT_PROVIDER")
	if defaultProvider == "" {
		defaultProvider = string(ProviderStripe)
	}
	config.DefaultProvider = ProviderType(strings.ToLower(defaultProvider))

	// Stripe configuration
	stripeEnabled := os.Getenv("STRIPE_ENABLED") != "false" // Default to true
	if stripeEnabled {
		config.Providers[ProviderStripe] = ProviderConfig{
			Enabled:       true,
			APIKey:        os.Getenv("STRIPE_SECRET_KEY"),
			WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
			Environment:   getEnvironment(),
		}
	}

	// Paddle configuration
	paddleEnabled := os.Getenv("PADDLE_ENABLED") == "true"
	if paddleEnabled {
		config.Providers[ProviderPaddle] = ProviderConfig{
			Enabled:       true,
			APIKey:        os.Getenv("PADDLE_API_KEY"),
			WebhookSecret: os.Getenv("PADDLE_WEBHOOK_SECRET"),
			Environment:   getEnvironment(),
		}
	}

	// Square configuration
	squareEnabled := os.Getenv("SQUARE_ENABLED") == "true"
	if squareEnabled {
		config.Providers[ProviderSquare] = ProviderConfig{
			Enabled:       true,
			APIKey:        os.Getenv("SQUARE_ACCESS_TOKEN"),
			WebhookSecret: os.Getenv("SQUARE_WEBHOOK_SIGNATURE_KEY"),
			Environment:   getEnvironment(),
		}
	}

	// Validate that at least one provider is configured
	if len(config.Providers) == 0 {
		return nil, fmt.Errorf("no payment providers configured")
	}

	// Validate that default provider is configured
	if _, exists := config.Providers[config.DefaultProvider]; !exists {
		return nil, fmt.Errorf("default provider %s is not configured", config.DefaultProvider)
	}

	return config, nil
}

// getEnvironment determines the environment from environment variables
func getEnvironment() string {
	if os.Getenv("ENVIRONMENT") == "production" {
		return "live"
	}
	return "test"
}

// ValidateProviderConfig validates that a provider configuration is complete
func (c *ProviderConfig) Validate() error {
	if !c.Enabled {
		return nil // Disabled providers don't need validation
	}

	if c.APIKey == "" {
		return fmt.Errorf("API key is required for enabled provider")
	}

	if c.Environment == "" {
		return fmt.Errorf("environment is required for enabled provider")
	}

	return nil
}
