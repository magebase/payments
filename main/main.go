package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apis/payments/services/stripe"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// App represents the main application
type App struct {
	fiberApp        *fiber.App
	customerService *stripe.CustomerService
	chargeService   *stripe.ChargeService
	refundService   *stripe.RefundService
}

// NewApp creates a new application instance
func NewApp() *App {
	// Initialize services
	customerService := stripe.NewCustomerService()
	chargeService := stripe.NewChargeService()
	refundService := stripe.NewRefundService()

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "Payments API",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// Add middleware
	fiberApp.Use(recover.New())
	fiberApp.Use(logger.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Register routes
	app := &App{
		fiberApp:        fiberApp,
		customerService: customerService,
		chargeService:   chargeService,
		refundService:   refundService,
	}

	app.registerRoutes()

	return app
}

// registerRoutes registers all API routes
func (a *App) registerRoutes() {
	// Health check
	a.fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "payments",
			"time":    time.Now().UTC(),
		})
	})

	// API routes
	api := a.fiberApp.Group("/api/v1")

	// Customer routes
	customers := api.Group("/customers")
	customers.Post("/", a.createCustomer)
	customers.Get("/:id", a.getCustomer)
	customers.Put("/:id", a.updateCustomer)
	customers.Delete("/:id", a.deleteCustomer)

	// Payment method routes
	paymentMethods := api.Group("/customers/:customerId/payment-methods")
	paymentMethods.Post("/", a.addPaymentMethod)
	paymentMethods.Get("/", a.listPaymentMethods)
	paymentMethods.Get("/:id", a.getPaymentMethod)
	paymentMethods.Delete("/:id", a.detachPaymentMethod)

	// Charge routes
	charges := api.Group("/charges")
	charges.Post("/", a.createCharge)
	charges.Get("/:id", a.getCharge)
	charges.Get("/", a.listCharges)

	// Refund routes
	refunds := api.Group("/refunds")
	refunds.Post("/", a.createRefund)
	refunds.Get("/:id", a.getRefund)
	refunds.Get("/", a.listRefunds)
}

// createCustomer handles customer creation
func (a *App) createCustomer(c *fiber.Ctx) error {
	var request stripe.CustomerRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	customer, err := a.customerService.CreateCustomer(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

// getCustomer handles customer retrieval
func (a *App) getCustomer(c *fiber.Ctx) error {
	customerID := c.Params("id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}

	customer, err := a.customerService.GetCustomer(c.Context(), customerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(customer)
}

// updateCustomer handles customer updates
func (a *App) updateCustomer(c *fiber.Ctx) error {
	customerID := c.Params("id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}

	var request stripe.CustomerRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	customer, err := a.customerService.UpdateCustomer(c.Context(), customerID, &request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(customer)
}

// deleteCustomer handles customer deletion
func (a *App) deleteCustomer(c *fiber.Ctx) error {
	customerID := c.Params("id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}

	err := a.customerService.DeleteCustomer(c.Context(), customerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// addPaymentMethod handles adding payment methods to customers
func (a *App) addPaymentMethod(c *fiber.Ctx) error {
	customerID := c.Params("customerId")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}

	var request stripe.PaymentMethodRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set the customer ID from the URL parameter
	request.Customer = customerID

	paymentMethod, err := a.customerService.AddPaymentMethod(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(paymentMethod)
}

// listPaymentMethods handles listing payment methods for a customer
func (a *App) listPaymentMethods(c *fiber.Ctx) error {
	customerID := c.Params("customerId")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer ID is required",
		})
	}

	paymentMethods, err := a.customerService.ListPaymentMethods(c.Context(), customerID, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(paymentMethods)
}

// getPaymentMethod handles payment method retrieval
func (a *App) getPaymentMethod(c *fiber.Ctx) error {
	paymentMethodID := c.Params("id")
	if paymentMethodID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Payment method ID is required",
		})
	}

	paymentMethod, err := a.customerService.GetPaymentMethod(c.Context(), paymentMethodID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(paymentMethod)
}

// detachPaymentMethod handles payment method detachment
func (a *App) detachPaymentMethod(c *fiber.Ctx) error {
	paymentMethodID := c.Params("id")
	if paymentMethodID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Payment method ID is required",
		})
	}

	err := a.customerService.DetachPaymentMethod(c.Context(), paymentMethodID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// createCharge handles charge creation
func (a *App) createCharge(c *fiber.Ctx) error {
	var request stripe.ChargeRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	charge, err := a.chargeService.CreateCharge(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(charge)
}

// getCharge handles charge retrieval
func (a *App) getCharge(c *fiber.Ctx) error {
	chargeID := c.Params("id")
	if chargeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Charge ID is required",
		})
	}

	charge, err := a.chargeService.GetCharge(c.Context(), chargeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(charge)
}

// listCharges handles listing charges
func (a *App) listCharges(c *fiber.Ctx) error {
	customerID := c.Query("customer_id")
	
	charges, err := a.chargeService.ListCharges(c.Context(), customerID, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(charges)
}

// createRefund handles refund creation
func (a *App) createRefund(c *fiber.Ctx) error {
	var request stripe.RefundRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	refund, err := a.refundService.CreateRefund(c.Context(), &request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(refund)
}

// getRefund handles refund retrieval
func (a *App) getRefund(c *fiber.Ctx) error {
	refundID := c.Params("id")
	if refundID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refund ID is required",
		})
	}

	refund, err := a.refundService.GetRefund(c.Context(), refundID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(refund)
}

// listRefunds handles listing refunds
func (a *App) listRefunds(c *fiber.Ctx) error {
	chargeID := c.Query("charge_id")
	if chargeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Charge ID is required",
		})
	}
	
	refunds, err := a.refundService.ListRefunds(c.Context(), chargeID, 100)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(refunds)
}

// Run starts the application
func (a *App) Run(port string) error {
	// Start the server
	go func() {
		if err := a.fiberApp.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.fiberApp.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
	return nil
}

// initTracing initializes OpenTelemetry tracing
func initTracing() error {
	ctx := context.Background()

	// Create OTLP exporter
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4317"))
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %v", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("payments"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %v", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return nil
}

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize tracing
	if err := initTracing(); err != nil {
		log.Printf("Warning: Failed to initialize tracing: %v", err)
	}

	// Create and run the application
	app := NewApp()
	
	log.Printf("Starting Payments API server on port %s", port)
	if err := app.Run(port); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
