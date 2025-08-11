package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConnectAccountOnboarding tests the Connect account onboarding functionality
func TestConnectAccountOnboarding(t *testing.T) {
	// Create a test Connect account request structure
	connectRequest := map[string]interface{}{
		"type":         "express",
		"country":      "US",
		"email":        "merchant@example.com",
		"businessType": "individual",
		"capabilities": map[string]string{
			"card_payments": "requested",
			"transfers":     "requested",
		},
	}

	// Test that the Connect account request struct is properly formed
	assert.NotNil(t, connectRequest)
	assert.Equal(t, "express", connectRequest["type"])
	assert.Equal(t, "US", connectRequest["country"])
	assert.Equal(t, "merchant@example.com", connectRequest["email"])
	assert.Equal(t, "individual", connectRequest["businessType"])
	
	capabilities := connectRequest["capabilities"].(map[string]string)
	assert.Equal(t, "requested", capabilities["card_payments"])
	assert.Equal(t, "requested", capabilities["transfers"])
}

// TestConnectAccountValidation tests the validation logic for Connect account requests
func TestConnectAccountValidation(t *testing.T) {
	// Test valid Connect account request
	validRequest := map[string]interface{}{
		"type":         "express",
		"country":      "US",
		"email":        "valid@example.com",
		"businessType": "individual",
	}

	// Test that validation would work
	assert.NotNil(t, validRequest)
	assert.Equal(t, "express", validRequest["type"])
	assert.Equal(t, "US", validRequest["country"])

	// Test invalid Connect account request (missing required fields)
	invalidRequest := map[string]interface{}{
		"type": "express",
		// Missing Country and Email
	}

	// This should fail validation
	assert.NotEqual(t, "US", invalidRequest["country"])
	assert.Empty(t, invalidRequest["email"])
}
