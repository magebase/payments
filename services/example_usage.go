package services

import (
	"context"
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the payment gateway abstraction layer
func ExampleUsage() {
	// Create a payment service with the default provider (Stripe)
	paymentService, err := NewPaymentService()
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	fmt.Printf("Current payment provider: %s\n", paymentService.GetProviderName())

	// Check provider capabilities
	capabilities := paymentService.GetCapabilities()
	fmt.Printf("Supports subscriptions: %t\n", capabilities.SupportsSubscriptions)
	fmt.Printf("Supports Connect: %t\n", capabilities.SupportsConnect)
	fmt.Printf("Supports Tax: %t\n", capabilities.SupportsTax)
	fmt.Printf("Max payment amount: %d cents\n", capabilities.MaxPaymentAmount)

	// Create a customer
	customerReq := &CustomerRequest{
		Email:       "customer@example.com",
		Name:        "John Doe",
		Phone:       "+1234567890",
		Description: "Example customer",
		Metadata: map[string]string{
			"source": "example",
		},
	}

	customer, err := paymentService.CreateCustomer(context.Background(), customerReq)
	if err != nil {
		log.Printf("Failed to create customer: %v", err)
	} else {
		fmt.Printf("Created customer: %s\n", customer.ID)
	}

	// Switch to a different provider (if available)
	availableProviders := paymentService.GetAvailableProviders()
	fmt.Printf("Available providers: %v\n", availableProviders)

	// Example: Switch to Paddle if available
	for _, provider := range availableProviders {
		if provider == ProviderPaddle {
			err := paymentService.SwitchProvider(ProviderPaddle)
			if err != nil {
				log.Printf("Failed to switch to Paddle: %v", err)
			} else {
				fmt.Printf("Switched to provider: %s\n", paymentService.GetProviderName())

				// Check Paddle capabilities
				paddleCapabilities := paymentService.GetCapabilities()
				fmt.Printf("Paddle supports Connect: %t\n", paddleCapabilities.SupportsConnect)
				fmt.Printf("Paddle supports Tax: %t\n", paddleCapabilities.SupportsTax)
			}
			break
		}
	}

	// Example: Create a charge with validation
	chargeReq := &ChargeRequest{
		Amount:      2500, // $25.00
		Currency:    "usd",
		CustomerID:  customer.ID,
		Description: "Example charge",
		Metadata: map[string]string{
			"source": "example",
		},
	}

	// Validate the charge request
	err = paymentService.ValidatePaymentRequest(chargeReq)
	if err != nil {
		log.Printf("Charge request validation failed: %v", err)
		return
	}

	// Create the charge
	charge, err := paymentService.CreateCharge(context.Background(), chargeReq)
	if err != nil {
		log.Printf("Failed to create charge: %v", err)
	} else {
		fmt.Printf("Created charge: %s for $%.2f\n", charge.ID, float64(charge.Amount)/100)
	}
}

// ExampleProviderSwitching demonstrates how to switch between providers
func ExampleProviderSwitching() {
	// Create a payment service with Stripe
	stripeService, err := NewPaymentServiceWithProvider(ProviderStripe)
	if err != nil {
		log.Fatalf("Failed to create Stripe service: %v", err)
	}

	fmt.Printf("Initial provider: %s\n", stripeService.GetProviderName())

	// Check Stripe capabilities
	if stripeService.SupportsSubscriptions() {
		fmt.Println("Stripe supports subscriptions")
	}

	if stripeService.SupportsConnect() {
		fmt.Println("Stripe supports Connect")
	}

	// Switch to Paddle (if available)
	err = stripeService.SwitchProvider(ProviderPaddle)
	if err != nil {
		log.Printf("Failed to switch to Paddle: %v", err)
	} else {
		fmt.Printf("Switched to provider: %s\n", stripeService.GetProviderName())

		// Check Paddle capabilities
		if !stripeService.SupportsConnect() {
			fmt.Println("Paddle does not support Connect")
		}
	}

	// Switch back to Stripe
	err = stripeService.SwitchProvider(ProviderStripe)
	if err != nil {
		log.Printf("Failed to switch back to Stripe: %v", err)
	} else {
		fmt.Printf("Switched back to provider: %s\n", stripeService.GetProviderName())
	}
}

// ExampleCapabilityChecking demonstrates how to check provider capabilities
func ExampleCapabilityChecking() {
	// Create a payment service
	paymentService, err := NewPaymentService()
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	// Check specific capabilities
	if paymentService.SupportsSubscriptions() {
		fmt.Println("Current provider supports subscriptions")
	} else {
		fmt.Println("Current provider does not support subscriptions")
	}

	if paymentService.SupportsConnect() {
		fmt.Println("Current provider supports Connect")
	} else {
		fmt.Println("Current provider does not support Connect")
	}

	if paymentService.SupportsTax() {
		fmt.Println("Current provider supports Tax")
	} else {
		fmt.Println("Current provider does not support Tax")
	}

	// Get supported currencies and countries
	currencies := paymentService.GetSupportedCurrencies()
	fmt.Printf("Supported currencies: %v\n", currencies)

	countries := paymentService.GetSupportedCountries()
	fmt.Printf("Supported countries: %v\n", countries)

	maxAmount := paymentService.GetMaxPaymentAmount()
	fmt.Printf("Maximum payment amount: $%.2f\n", float64(maxAmount)/100)
}

// ExampleErrorHandling demonstrates how to handle provider-specific errors
func ExampleErrorHandling() {
	// Create a payment service
	paymentService, err := NewPaymentService()
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
	}

	// Try to create a charge with an unsupported currency
	chargeReq := &ChargeRequest{
		Amount:     1000,
		Currency:   "unsupported",
		CustomerID: "cus_test_123",
	}

	// This should fail validation
	err = paymentService.ValidatePaymentRequest(chargeReq)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	// Try to create a charge with an amount that's too high
	highAmountReq := &ChargeRequest{
		Amount:     999999999999,
		Currency:   "usd",
		CustomerID: "cus_test_123",
	}

	// This should also fail validation
	err = paymentService.ValidatePaymentRequest(highAmountReq)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}
}
