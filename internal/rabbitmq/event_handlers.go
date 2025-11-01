package rabbitmq

import (
	"encoding/json"
	"log"
	"time"

	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	"gorm.io/gorm"

	localModels "github.com/PayeTonKawa-EPSI-2025/Orders/internal/models"
)

// GenericEvent is a structure for parsing event data from other services
type GenericEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// CustomerEvent represents an order event from the Customers service
type CustomerEvent struct {
	Type      string          `json:"type"`
	Customer  models.Customer `json:"customer"`
	Timestamp time.Time       `json:"timestamp"`
}

// ProductEvent represents a product event from the Products service
type ProductEvent struct {
	Type      string         `json:"type"`
	Product   models.Product `json:"product"`
	Timestamp time.Time      `json:"timestamp"`
}

// SetupEventHandlers configures handlers for different event types
func SetupEventHandlers(dbConn *gorm.DB) *EventRouter {
	router := NewEventRouter()

	// Handle customer.created events - create customer in local DB
	router.RegisterHandler("customer.created", func(body []byte) error {
		var event CustomerEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("Error unmarshaling customer event: %v", err)
			return err
		}

		log.Printf("Received customer.created event for customer %d", event.Customer.ID)

		// Create the customer in the local database
		customer := localModels.Customer{
			ID: event.Customer.ID,
		}

		if err := dbConn.Create(&customer).Error; err != nil {
			log.Printf("Error creating customer in DB: %v", err)
			return err
		}

		log.Printf("Successfully created customer %d in local database", customer.ID)

		return nil
	})

	// Handle product.created events - create product in local DB
	router.RegisterHandler("product.created", func(body []byte) error {
		var event ProductEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("Error unmarshaling product event: %v", err)
			return err
		}

		log.Printf("Received product.created event for product %d", event.Product.ID)

		// Create the product in the local database
		product := localModels.Product{
			ID: event.Product.ID,
		}

		if err := dbConn.Create(&product).Error; err != nil {
			log.Printf("Error creating product in DB: %v", err)
			return err
		}

		log.Printf("Successfully created product %d in local database", product.ID)

		return nil
	})

	// Catch-all handler for debugging - will receive all events
	// Useful during development, can be removed in production
	router.RegisterHandler("#", func(body []byte) error {
		var generic GenericEvent
		if err := json.Unmarshal(body, &generic); err != nil {
			log.Printf("Error unmarshaling generic event: %v", err)
			// Don't return error here as it might be a different format
			// Just log and continue
		} else {
			log.Printf("Received event of type %s", generic.Type)
		}

		// Log the raw message for debugging
		log.Printf("Raw event: %s", string(body))
		return nil
	})

	return router
}
