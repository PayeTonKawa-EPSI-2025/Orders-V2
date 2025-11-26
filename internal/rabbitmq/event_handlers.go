package rabbitmq

import (
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/rabbitmq/event_handlers"
	"gorm.io/gorm"
)

// SetupEventHandlers configures handlers for different event types
func SetupEventHandlers(dbConn *gorm.DB) *EventRouter {
	router := NewEventRouter()

	// Initialize event handlers
	customerHandlers := event_handlers.NewCustomerEventHandlers(dbConn)
	productHandlers := event_handlers.NewProductEventHandlers(dbConn)
	debugHandlers := event_handlers.NewDebugEventHandlers()

	// Register customer event handlers
	router.RegisterHandler("customer.created", customerHandlers.HandleCustomerCreated)
	router.RegisterHandler("customer.updated", customerHandlers.HandleCustomerUpdated)
	router.RegisterHandler("customer.deleted", customerHandlers.HandleCustomerDeleted)

	// Register product event handlers
	router.RegisterHandler("product.created", productHandlers.HandleProductCreated)
	router.RegisterHandler("product.updated", productHandlers.HandleProductUpdated)
	router.RegisterHandler("product.deleted", productHandlers.HandleProductDeleted)

	// Register debug catch-all handler
	// Useful during development, can be removed in production
	router.RegisterHandler("#", debugHandlers.HandleAllEvents)

	return router
}
