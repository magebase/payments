package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"apis/payments/services"
)

// TestSubscriptionPlanCreation tests that subscription plans can be created and managed
func TestSubscriptionPlanCreation(t *testing.T) {
	t.Run("Can_create_subscription_plan", func(t *testing.T) {
		// Arrange
		service := services.NewPaymentServiceWithMockGateway()

		planRequest := &services.SubscriptionPlanRequest{
			Name:        "Pro Plan",
			Description: "Professional subscription plan",
			Amount:      2999, // $29.99 in cents
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
			Metadata: map[string]string{
				"features": "unlimited_access,priority_support",
			},
		}

		// Act
		plan, err := service.CreateSubscriptionPlan(context.Background(), planRequest)

		// Assert
		require.NoError(t, err, "Should be able to create subscription plan")
		assert.NotNil(t, plan, "Plan should not be nil")
		assert.Equal(t, "Pro Plan", plan.Name, "Plan name should match")
		assert.Equal(t, "Professional subscription plan", plan.Description, "Plan description should match")
		assert.Equal(t, int64(2999), plan.Amount, "Plan amount should match")
		assert.Equal(t, "usd", plan.Currency, "Plan currency should match")
		assert.Equal(t, "month", plan.Interval, "Plan interval should match")
		assert.Equal(t, 1, plan.IntervalCount, "Plan interval count should match")
		assert.Equal(t, "unlimited_access,priority_support", plan.Metadata["features"], "Plan metadata should match")
		assert.NotEmpty(t, plan.ID, "Plan should have an ID")
		assert.True(t, plan.Created > 0, "Plan should have creation timestamp")
		assert.True(t, plan.Updated > 0, "Plan should have update timestamp")
	})

	t.Run("Can_get_subscription_plan_by_id", func(t *testing.T) {
		// Arrange
		service := services.NewPaymentServiceWithMockGateway()

		planRequest := &services.SubscriptionPlanRequest{
			Name:        "Basic Plan",
			Description: "Basic subscription plan",
			Amount:      999, // $9.99 in cents
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
		}

		createdPlan, err := service.CreateSubscriptionPlan(context.Background(), planRequest)
		require.NoError(t, err, "Should be able to create plan")

		// Act
		retrievedPlan, err := service.GetSubscriptionPlan(context.Background(), createdPlan.ID)

		// Assert
		require.NoError(t, err, "Should be able to retrieve subscription plan")
		assert.NotNil(t, retrievedPlan, "Retrieved plan should not be nil")
		assert.Equal(t, createdPlan.ID, retrievedPlan.ID, "Plan ID should match")
		assert.Equal(t, "Basic Plan", retrievedPlan.Name, "Plan name should match")
		assert.Equal(t, int64(999), retrievedPlan.Amount, "Plan amount should match")
	})

	t.Run("Can_list_subscription_plans", func(t *testing.T) {
		// Arrange
		service := services.NewPaymentServiceWithMockGateway()

		// Create multiple plans
		plan1 := &services.SubscriptionPlanRequest{
			Name:        "Plan 1",
			Description: "First plan",
			Amount:      1000,
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
		}

		plan2 := &services.SubscriptionPlanRequest{
			Name:        "Plan 2",
			Description: "Second plan",
			Amount:      2000,
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
		}

		// Create both plans before listing
		_, err := service.CreateSubscriptionPlan(context.Background(), plan1)
		require.NoError(t, err, "Should be able to create first plan")

		_, err = service.CreateSubscriptionPlan(context.Background(), plan2)
		require.NoError(t, err, "Should be able to create second plan")

		// Act
		plans, err := service.ListSubscriptionPlans(context.Background(), &services.SubscriptionPlanListParams{})

		// Assert
		require.NoError(t, err, "Should be able to list subscription plans")
		assert.NotNil(t, plans, "Plans list should not be nil")
		assert.Len(t, plans, 2, "Should have two plans")
		
		// Verify the plans are correct
		planNames := make([]string, len(plans))
		for i, plan := range plans {
			planNames[i] = plan.Name
		}
		assert.Contains(t, planNames, "Plan 1", "Should contain Plan 1")
		assert.Contains(t, planNames, "Plan 2", "Should contain Plan 2")
	})

	t.Run("Can_update_subscription_plan", func(t *testing.T) {
		// Arrange
		service := services.NewPaymentServiceWithMockGateway()

		planRequest := &services.SubscriptionPlanRequest{
			Name:        "Original Plan",
			Description: "Original description",
			Amount:      1000,
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
		}

		createdPlan, err := service.CreateSubscriptionPlan(context.Background(), planRequest)
		require.NoError(t, err, "Should be able to create plan")

		// Store the original timestamp for comparison
		originalUpdated := createdPlan.Updated

		updateRequest := &services.SubscriptionPlanUpdateRequest{
			Name:        &[]string{"Updated Plan"}[0],
			Description: &[]string{"Updated description"}[0],
			Amount:      &[]int64{1500}[0],
		}

		// Act
		updatedPlan, err := service.UpdateSubscriptionPlan(context.Background(), createdPlan.ID, updateRequest)

		// Assert
		require.NoError(t, err, "Should be able to update subscription plan")
		assert.NotNil(t, updatedPlan, "Updated plan should not be nil")
		assert.Equal(t, "Updated Plan", updatedPlan.Name, "Plan name should be updated")
		assert.Equal(t, "Updated description", updatedPlan.Description, "Plan description should be updated")
		assert.Equal(t, int64(1500), updatedPlan.Amount, "Plan amount should be updated")
		assert.True(t, updatedPlan.Updated > originalUpdated, "Update timestamp should be newer")
	})

	t.Run("Can_delete_subscription_plan", func(t *testing.T) {
		// Arrange
		service := services.NewPaymentServiceWithMockGateway()

		planRequest := &services.SubscriptionPlanRequest{
			Name:        "Plan to Delete",
			Description: "This plan will be deleted",
			Amount:      1000,
			Currency:    "usd",
			Interval:    "month",
			IntervalCount: 1,
		}

		createdPlan, err := service.CreateSubscriptionPlan(context.Background(), planRequest)
		require.NoError(t, err, "Should be able to create plan")

		// Act
		err = service.DeleteSubscriptionPlan(context.Background(), createdPlan.ID)

		// Assert
		require.NoError(t, err, "Should be able to delete subscription plan")

		// Verify plan is deleted
		_, err = service.GetSubscriptionPlan(context.Background(), createdPlan.ID)
		assert.Error(t, err, "Should not be able to retrieve deleted plan")
	})
}
