package services

import (
	"fmt"
	"os"
	"strings"

	"github.com/magebase/payments/services/stripe"
)

// Global factory instance
var globalFactory *DefaultProviderFactory

// InitializeFactory initializes the global payment gateway factory
func InitializeFactory() {
	globalFactory = NewDefaultProviderFactory()
	
	// Register Stripe provider
	globalFactory.RegisterProvider("stripe", func(config map[string]interface{}) (PaymentGateway, error) {
		return stripe.NewStripeGateway(config)
	})
	
	// TODO: Register additional providers as they are implemented
	// globalFactory.RegisterProvider("paddle", paddle.NewPaddleGateway)
	// globalFactory.RegisterProvider("square", square.NewSquareGateway)
}

// GetFactory returns the global payment gateway factory
func GetFactory() *DefaultProviderFactory {
	if globalFactory == nil {
		InitializeFactory()
	}
	return globalFactory
}

// CreateGatewayFromEnv creates a payment gateway from environment variables
func CreateGatewayFromEnv() (PaymentGateway, error) {
	factory := GetFactory()
	
	// Get provider from environment
	provider := strings.ToLower(os.Getenv("PAYMENT_PROVIDER"))
	if provider == "" {
		provider = "stripe" // Default to Stripe
	}
	
	// Build configuration from environment
	config := buildConfigFromEnv(provider)
	
	// Validate configuration
	if err := factory.ValidateConfig(provider, config); err != nil {
		return nil, fmt.Errorf("invalid configuration for provider %s: %w", provider, err)
	}
	
	// Create gateway
	gateway, err := factory.CreateGateway(provider, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway for provider %s: %w", provider, err)
	}
	
	return gateway, nil
}

// buildConfigFromEnv builds provider configuration from environment variables
func buildConfigFromEnv(provider string) map[string]interface{} {
	config := make(map[string]interface{})
	
	switch provider {
	case "stripe":
		config["api_key"] = os.Getenv("STRIPE_API_KEY")
		config["webhook_secret"] = os.Getenv("STRIPE_WEBHOOK_SECRET")
		config["publishable_key"] = os.Getenv("STRIPE_PUBLISHABLE_KEY")
		
	case "paddle":
		config["vendor_id"] = os.Getenv("PADDLE_VENDOR_ID")
		config["vendor_auth_code"] = os.Getenv("PADDLE_VENDOR_AUTH_CODE")
		config["environment"] = os.Getenv("PADDLE_ENVIRONMENT") // sandbox or production
		
	case "square":
		config["application_id"] = os.Getenv("SQUARE_APPLICATION_ID")
		config["access_token"] = os.Getenv("SQUARE_ACCESS_TOKEN")
		config["environment"] = os.Getenv("SQUARE_ENVIRONMENT") // sandbox or production
		
	default:
		// For unknown providers, try to get generic config
		config["api_key"] = os.Getenv("PAYMENT_API_KEY")
		config["secret_key"] = os.Getenv("PAYMENT_SECRET_KEY")
		config["environment"] = os.Getenv("PAYMENT_ENVIRONMENT")
	}
	
	return config
}

// ValidateProviderConfig validates configuration for a specific provider
func ValidateProviderConfig(provider string, config map[string]interface{}) error {
	switch provider {
	case "stripe":
		return validateStripeConfig(config)
	case "paddle":
		return validatePaddleConfig(config)
	case "square":
		return validateSquareConfig(config)
	default:
		return &UnsupportedProviderError{Provider: provider}
	}
}

// validateStripeConfig validates Stripe configuration
func validateStripeConfig(config map[string]interface{}) error {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return &InvalidConfigError{Message: "stripe api_key is required"}
	}
	
	// Validate API key format (starts with sk_ for secret keys)
	if !strings.HasPrefix(apiKey, "sk_") {
		return &InvalidConfigError{Message: "stripe api_key must start with 'sk_'"}
	}
	
	return nil
}

// validatePaddleConfig validates Paddle configuration
func validatePaddleConfig(config map[string]interface{}) error {
	vendorID, ok := config["vendor_id"].(string)
	if !ok || vendorID == "" {
		return &InvalidConfigError{Message: "paddle vendor_id is required"}
	}
	
	vendorAuthCode, ok := config["vendor_auth_code"].(string)
	if !ok || vendorAuthCode == "" {
		return &InvalidConfigError{Message: "paddle vendor_auth_code is required"}
	}
	
	environment, ok := config["environment"].(string)
	if !ok || environment == "" {
		return &InvalidConfigError{Message: "paddle environment is required"}
	}
	
	if environment != "sandbox" && environment != "production" {
		return &InvalidConfigError{Message: "paddle environment must be 'sandbox' or 'production'"}
	}
	
	return nil
}

// validateSquareConfig validates Square configuration
func validateSquareConfig(config map[string]interface{}) error {
	applicationID, ok := config["application_id"].(string)
	if !ok || applicationID == "" {
		return &InvalidConfigError{Message: "square application_id is required"}
	}
	
	accessToken, ok := config["access_token"].(string)
	if !ok || accessToken == "" {
		return &InvalidConfigError{Message: "square access_token is required"}
	}
	
	environment, ok := config["environment"].(string)
	if !ok || environment == "" {
		return &InvalidConfigError{Message: "square environment is required"}
	}
	
	if environment != "sandbox" && environment != "production" {
		return &InvalidConfigError{Message: "square environment must be 'sandbox' or 'production'"}
	}
	
	return nil
}

// GetSupportedProviders returns a list of all supported payment providers
func GetSupportedProviders() []string {
	factory := GetFactory()
	return factory.GetSupportedProviders()
}

// IsProviderSupported checks if a provider is supported
func IsProviderSupported(provider string) bool {
	factory := GetFactory()
	providers := factory.GetSupportedProviders()
	
	for _, p := range providers {
		if p == provider {
			return true
		}
	}
	
	return false
}

// GetProviderCapabilities returns the capabilities of a specific provider
func GetProviderCapabilities(provider string) (*GatewayCapabilities, error) {
	gateway, err := CreateGatewayFromEnv()
	if err != nil {
		return nil, err
	}
	
	// Check if the gateway supports the requested provider
	if gateway.GetProvider() != provider {
		return nil, &UnsupportedProviderError{Provider: provider}
	}
	
	capabilities := gateway.GetCapabilities()
	return &capabilities, nil
}

// Feature detection helpers

// SupportsFeature checks if a provider supports a specific feature
func SupportsFeature(provider, feature string) bool {
	capabilities, err := GetProviderCapabilities(provider)
	if err != nil {
		return false
	}
	
	switch feature {
	case "customers":
		return capabilities.SupportsCustomers
	case "charges":
		return capabilities.SupportsCharges
	case "refunds":
		return capabilities.SupportsRefunds
	case "subscriptions":
		return capabilities.SupportsSubscriptions
	case "disputes":
		return capabilities.SupportsDisputes
	case "connect":
		return capabilities.SupportsConnect
	case "tax":
		return capabilities.SupportsTax
	default:
		return false
	}
}

// GetProviderLimits returns the limits for a specific provider
func GetProviderLimits(provider string) (minAmount, maxAmount int64, currencies []string, err error) {
	capabilities, err := GetProviderCapabilities(provider)
	if err != nil {
		return 0, 0, nil, err
	}
	
	return capabilities.MinChargeAmount, capabilities.MaxChargeAmount, capabilities.SupportedCurrencies, nil
}