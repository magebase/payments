package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"apis/billing/sdk"
	"github.com/stretchr/testify/require"
)

func TestBillingSDKCreatesBillableEvent(t *testing.T) {
	// Mock billing service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/bill", r.URL.Path)
		w.WriteHeader(201)
	}))
	defer server.Close()

	// Point SDK to mock server
	sdk.SetBillingServiceURL(server.URL)

	// Call SDK
	err := sdk.CreateBillableEvent("user-123", "api-call", 1.23)
	require.NoError(t, err)
}
