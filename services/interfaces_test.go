package services

import (
	"os"
	"testing"
)

func TestGatewayFactory(t *testing.T) {
	// Test environment setup
	os.Setenv("STRIPE_ENABLED", "true")
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_123")
	os.Setenv("ENVIRONMENT", "test")
	defer func() {
		os.Unsetenv("STRIPE_ENABLED")
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_WEBHOOK_SECRET")
		os.Unsetenv("ENVIRONMENT")
	}()

	t.Run("Create Gateway Factory", func(t *testing.T) {
		factory, err := NewGatewayFactory()
		if err != nil {
			t.Fatalf("Failed to create gateway factory: %v", err)
		}

		if factory == nil {
			t.Fatal("Gateway factory should not be nil")
		}

		if factory.config == nil {
			t.Fatal("Gateway factory config should not be nil")
		}
	})

	t.Run("Create Default Gateway", func(t *testing.T) {
		factory, err := NewGatewayFactory()
		if err != nil {
			t.Fatalf("Failed to create gateway factory: %v", err)
		}

		gateway, err := factory.CreateDefaultGateway()
		if err != nil {
			t.Fatalf("Failed to create default gateway: %v", err)
		}

		if gateway == nil {
			t.Fatal("Default gateway should not be nil")
		}

		providerName := gateway.GetProviderName()
		if providerName != "stripe" {
			t.Errorf("Expected provider name 'stripe', got '%s'", providerName)
		}
	})

	t.Run("Create Stripe Gateway", func(t *testing.T) {
		factory, err := NewGatewayFactory()
		if err != nil {
			t.Fatalf("Failed to create gateway factory: %v", err)
		}

		gateway, err := factory.CreateGateway(ProviderStripe)
		if err != nil {
			t.Fatalf("Failed to create Stripe gateway: %v", err)
		}

		if gateway == nil {
			t.Fatal("Stripe gateway should not be nil")
		}

		providerName := gateway.GetProviderName()
		if providerName != "stripe" {
			t.Errorf("Expected provider name 'stripe', got '%s'", providerName)
		}
	})

	t.Run("Get Available Providers", func(t *testing.T) {
		factory, err := NewGatewayFactory()
		if err != nil {
			t.Fatalf("Failed to create gateway factory: %v", err)
		}

		providers := factory.GetAvailableProviders()
		if len(providers) == 0 {
			t.Fatal("Should have at least one available provider")
		}

		foundStripe := false
		for _, provider := range providers {
			if provider == ProviderStripe {
				foundStripe = true
				break
			}
		}

		if !foundStripe {
			t.Error("Stripe should be in available providers")
		}
	})
}

func TestStripeGateway(t *testing.T) {
	// Test environment setup
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_123")
	os.Setenv("ENVIRONMENT", "test")
	defer func() {
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_WEBHOOK_SECRET")
		os.Unsetenv("ENVIRONMENT")
	}()

	t.Run("Create Stripe Gateway", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "sk_test_123",
			WebhookSecret: "whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewStripeGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Stripe gateway: %v", err)
		}

		if gateway == nil {
			t.Fatal("Stripe gateway should not be nil")
		}
	})

	t.Run("Get Provider Name", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "sk_test_123",
			WebhookSecret: "whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewStripeGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Stripe gateway: %v", err)
		}

		providerName := gateway.GetProviderName()
		if providerName != "stripe" {
			t.Errorf("Expected provider name 'stripe', got '%s'", providerName)
		}
	})

	t.Run("Get Capabilities", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "sk_test_123",
			WebhookSecret: "whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewStripeGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Stripe gateway: %v", err)
		}

		capabilities := gateway.GetCapabilities()
		if !capabilities.SupportsSubscriptions {
			t.Error("Stripe should support subscriptions")
		}

		if !capabilities.SupportsConnect {
			t.Error("Stripe should support Connect")
		}

		if !capabilities.SupportsTax {
			t.Error("Stripe should support Tax")
		}

		if !capabilities.SupportsInvoices {
			t.Error("Stripe should support Invoices")
		}

		if !capabilities.SupportsPayouts {
			t.Error("Stripe should support Payouts")
		}

		if !capabilities.SupportsDisputes {
			t.Error("Stripe should support Disputes")
		}

		if !capabilities.SupportsRefunds {
			t.Error("Stripe should support Refunds")
		}

		if capabilities.MaxPaymentAmount <= 0 {
			t.Error("Stripe should have a positive max payment amount")
		}

		if len(capabilities.SupportedCurrencies) == 0 {
			t.Error("Stripe should support multiple currencies")
		}

		if len(capabilities.SupportedCountries) == 0 {
			t.Error("Stripe should support multiple countries")
		}
	})
}

func TestPaddleGateway(t *testing.T) {
	t.Run("Create Paddle Gateway", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "paddle_test_123",
			WebhookSecret: "paddle_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewPaddleGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Paddle gateway: %v", err)
		}

		if gateway == nil {
			t.Fatal("Paddle gateway should not be nil")
		}
	})

	t.Run("Get Provider Name", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "paddle_test_123",
			WebhookSecret: "paddle_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewPaddleGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Paddle gateway: %v", err)
		}

		providerName := gateway.GetProviderName()
		if providerName != "paddle" {
			t.Errorf("Expected provider name 'paddle', got '%s'", providerName)
		}
	})

	t.Run("Get Capabilities", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "paddle_test_123",
			WebhookSecret: "paddle_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewPaddleGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Paddle gateway: %v", err)
		}

		capabilities := gateway.GetCapabilities()
		if !capabilities.SupportsSubscriptions {
			t.Error("Paddle should support subscriptions")
		}

		if capabilities.SupportsConnect {
			t.Error("Paddle should not support Connect")
		}

		if !capabilities.SupportsTax {
			t.Error("Paddle should support Tax")
		}

		if !capabilities.SupportsInvoices {
			t.Error("Paddle should support Invoices")
		}

		if capabilities.SupportsPayouts {
			t.Error("Paddle should not support Payouts")
		}

		if !capabilities.SupportsDisputes {
			t.Error("Paddle should support Disputes")
		}

		if !capabilities.SupportsRefunds {
			t.Error("Paddle should support Refunds")
		}
	})
}

func TestSquareGateway(t *testing.T) {
	t.Run("Create Square Gateway", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "square_test_123",
			WebhookSecret: "square_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewSquareGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Square gateway: %v", err)
		}

		if gateway == nil {
			t.Fatal("Square gateway should not be nil")
		}
	})

	t.Run("Get Provider Name", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "square_test_123",
			WebhookSecret: "square_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewSquareGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Square gateway: %v", err)
		}

		providerName := gateway.GetProviderName()
		if providerName != "square" {
			t.Errorf("Expected provider name 'square', got '%s'", providerName)
		}
	})

	t.Run("Get Capabilities", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "square_test_123",
			WebhookSecret: "square_whsec_test_123",
			Environment:   "test",
		}

		gateway, err := NewSquareGateway(config)
		if err != nil {
			t.Fatalf("Failed to create Square gateway: %v", err)
		}

		capabilities := gateway.GetCapabilities()
		if capabilities.SupportsSubscriptions {
			t.Error("Square should not support subscriptions")
		}

		if !capabilities.SupportsConnect {
			t.Error("Square should support Connect")
		}

		if capabilities.SupportsTax {
			t.Error("Square should not support Tax")
		}

		if !capabilities.SupportsInvoices {
			t.Error("Square should support Invoices")
		}

		if !capabilities.SupportsPayouts {
			t.Error("Square should support Payouts")
		}

		if !capabilities.SupportsDisputes {
			t.Error("Square should support Disputes")
		}

		if !capabilities.SupportsRefunds {
			t.Error("Square should support Refunds")
		}
	})
}

func TestPaymentService(t *testing.T) {
	// Test environment setup
	os.Setenv("STRIPE_ENABLED", "true")
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_123")
	os.Setenv("ENVIRONMENT", "test")
	defer func() {
		os.Unsetenv("STRIPE_ENABLED")
		os.Unsetenv("STRIPE_SECRET_KEY")
		os.Unsetenv("STRIPE_WEBHOOK_SECRET")
		os.Unsetenv("ENVIRONMENT")
	}()

	t.Run("Create Payment Service", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		if service == nil {
			t.Fatal("Payment service should not be nil")
		}

		if service.gateway == nil {
			t.Fatal("Payment service gateway should not be nil")
		}

		if service.factory == nil {
			t.Fatal("Payment service factory should not be nil")
		}
	})

	t.Run("Get Provider Name", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		providerName := service.GetProviderName()
		if providerName != "stripe" {
			t.Errorf("Expected provider name 'stripe', got '%s'", providerName)
		}
	})

	t.Run("Get Capabilities", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		capabilities := service.GetCapabilities()
		if !capabilities.SupportsSubscriptions {
			t.Error("Service should support subscriptions")
		}

		if !capabilities.SupportsConnect {
			t.Error("Service should support Connect")
		}

		if !capabilities.SupportsTax {
			t.Error("Service should support Tax")
		}
	})

	t.Run("Feature Detection", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		if !service.SupportsSubscriptions() {
			t.Error("Service should support subscriptions")
		}

		if !service.SupportsConnect() {
			t.Error("Service should support Connect")
		}

		if !service.SupportsTax() {
			t.Error("Service should support Tax")
		}

		if !service.SupportsInvoices() {
			t.Error("Service should support Invoices")
		}

		if !service.SupportsPayouts() {
			t.Error("Service should support Payouts")
		}

		if !service.SupportsDisputes() {
			t.Error("Service should support Disputes")
		}

		if !service.SupportsRefunds() {
			t.Error("Service should support Refunds")
		}
	})

	t.Run("Get Available Providers", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		providers := service.GetAvailableProviders()
		if len(providers) == 0 {
			t.Fatal("Should have at least one available provider")
		}

		foundStripe := false
		for _, provider := range providers {
			if provider == ProviderStripe {
				foundStripe = true
				break
			}
		}

		if !foundStripe {
			t.Error("Stripe should be in available providers")
		}
	})

	t.Run("Validate Payment Request", func(t *testing.T) {
		service, err := NewPaymentService()
		if err != nil {
			t.Fatalf("Failed to create payment service: %v", err)
		}

		// Valid request
		validReq := &ChargeRequest{
			Amount:     1000, // $10.00
			Currency:   "usd",
			CustomerID: "cus_test_123",
		}

		err = service.ValidatePaymentRequest(validReq)
		if err != nil {
			t.Errorf("Valid payment request should not fail validation: %v", err)
		}

		// Invalid currency
		invalidCurrencyReq := &ChargeRequest{
			Amount:     1000,
			Currency:   "invalid",
			CustomerID: "cus_test_123",
		}

		err = service.ValidatePaymentRequest(invalidCurrencyReq)
		if err == nil {
			t.Error("Invalid currency should fail validation")
		}

		// Amount too high
		invalidAmountReq := &ChargeRequest{
			Amount:     999999999999, // Very high amount
			Currency:   "usd",
			CustomerID: "cus_test_123",
		}

		err = service.ValidatePaymentRequest(invalidAmountReq)
		if err == nil {
			t.Error("Amount too high should fail validation")
		}
	})
}

func TestProviderConfigValidation(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "test_key_123",
			WebhookSecret: "test_secret_123",
			Environment:   "test",
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Valid config should not fail validation: %v", err)
		}
	})

	t.Run("Disabled Config", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       false,
			APIKey:        "",
			WebhookSecret: "",
			Environment:   "",
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Disabled config should not fail validation: %v", err)
		}
	})

	t.Run("Missing API Key", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "",
			WebhookSecret: "test_secret_123",
			Environment:   "test",
		}

		err := config.Validate()
		if err == nil {
			t.Error("Config missing API key should fail validation")
		}
	})

	t.Run("Missing Environment", func(t *testing.T) {
		config := &ProviderConfig{
			Enabled:       true,
			APIKey:        "test_key_123",
			WebhookSecret: "test_secret_123",
			Environment:   "",
		}

		err := config.Validate()
		if err == nil {
			t.Error("Config missing environment should fail validation")
		}
	})
}
