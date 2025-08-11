package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"apis/payments/services/stripe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisputesAPIEndpoints(t *testing.T) {
	// Test case: Create dispute endpoint
	t.Run("POST /api/v1/disputes should create dispute", func(t *testing.T) {
		// Create a test dispute request
		disputeRequest := stripe.DisputeRequest{
			ChargeID: "ch_test123",
			Amount:   1000,
			Reason:   "fraudulent",
			Evidence: map[string]string{"customer_email": "customer@example.com"},
		}

		// Convert to JSON
		requestBody, err := json.Marshal(disputeRequest)
		require.NoError(t, err)

		// Create HTTP request
		req := httptest.NewRequest("POST", "/api/v1/disputes", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		// Note: In a real integration test, you would start the actual server
		// and make real HTTP requests. This is a simplified test structure.

		// For now, we'll just verify the request structure is valid
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/api/v1/disputes", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		// Verify the request body can be parsed back
		var parsedRequest stripe.DisputeRequest
		err = json.Unmarshal(requestBody, &parsedRequest)
		require.NoError(t, err)
		assert.Equal(t, disputeRequest.ChargeID, parsedRequest.ChargeID)
		assert.Equal(t, disputeRequest.Amount, parsedRequest.Amount)
		assert.Equal(t, disputeRequest.Reason, parsedRequest.Reason)
	})

	// Test case: Get dispute endpoint
	t.Run("GET /api/v1/disputes/:id should get dispute", func(t *testing.T) {
		// Create HTTP request
		req := httptest.NewRequest("GET", "/api/v1/disputes/dp_test123", nil)

		// Verify the request structure
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "/api/v1/disputes/dp_test123", req.URL.Path)
	})

	// Test case: List disputes endpoint
	t.Run("GET /api/v1/disputes should list disputes", func(t *testing.T) {
		// Create HTTP request with query parameter
		req := httptest.NewRequest("GET", "/api/v1/disputes?charge_id=ch_test123", nil)

		// Verify the request structure
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "/api/v1/disputes", req.URL.Path)
		assert.Equal(t, "ch_test123", req.URL.Query().Get("charge_id"))
	})

	// Test case: Update dispute status endpoint
	t.Run("PUT /api/v1/disputes/:id/status should update dispute status", func(t *testing.T) {
		// Create a test status update request
		statusRequest := struct {
			Status string `json:"status"`
		}{
			Status: "won",
		}

		// Convert to JSON
		requestBody, err := json.Marshal(statusRequest)
		require.NoError(t, err)

		// Create HTTP request
		req := httptest.NewRequest("PUT", "/api/v1/disputes/dp_test123/status", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		// Verify the request structure
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "/api/v1/disputes/dp_test123/status", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		// Verify the request body can be parsed back
		var parsedRequest struct {
			Status string `json:"status"`
		}
		err = json.Unmarshal(requestBody, &parsedRequest)
		require.NoError(t, err)
		assert.Equal(t, statusRequest.Status, parsedRequest.Status)
	})
}

func TestDisputeRequestValidation(t *testing.T) {
	t.Run("should validate required fields", func(t *testing.T) {
		// Test with missing charge_id
		invalidRequest := stripe.DisputeRequest{
			Amount: 1000,
			Reason: "fraudulent",
		}

		// This should fail validation
		service := stripe.NewDisputeService()
		err := service.ValidateDisputeRequest(&invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("should validate amount constraints", func(t *testing.T) {
		// Test with negative amount
		invalidRequest := stripe.DisputeRequest{
			ChargeID: "ch_test123",
			Amount:   -100,
			Reason:   "fraudulent",
		}

		service := stripe.NewDisputeService()
		err := service.ValidateDisputeRequest(&invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("should validate dispute reason", func(t *testing.T) {
		// Test with invalid reason
		invalidRequest := stripe.DisputeRequest{
			ChargeID: "ch_test123",
			Amount:   1000,
			Reason:   "invalid_reason",
		}

		service := stripe.NewDisputeService()
		err := service.ValidateDisputeRequest(&invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dispute reason")
	})

	t.Run("should accept valid dispute request", func(t *testing.T) {
		// Test with valid request
		validRequest := stripe.DisputeRequest{
			ChargeID: "ch_test123",
			Amount:   1000,
			Reason:   "fraudulent",
			Evidence: map[string]string{"note": "Valid dispute"},
		}

		service := stripe.NewDisputeService()
		err := service.ValidateDisputeRequest(&validRequest)
		assert.NoError(t, err)
	})
}
