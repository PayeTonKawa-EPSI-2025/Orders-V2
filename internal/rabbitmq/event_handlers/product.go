package event_handlers

import (
	"encoding/json"
	"log"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	localModels "github.com/PayeTonKawa-EPSI-2025/Orders/internal/models"
	"gorm.io/gorm"
)

// ProductEventHandlers provides handlers for product-related events
type ProductEventHandlers struct {
	db *gorm.DB
}

// NewProductEventHandlers creates a new product event handlers instance
func NewProductEventHandlers(db *gorm.DB) *ProductEventHandlers {
	return &ProductEventHandlers{db: db}
}

// HandleProductCreated handles the product.created event
func (h *ProductEventHandlers) HandleProductCreated(body []byte) error {
	var event events.ProductEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling product.created event: %v", err)
		return err
	}

	log.Printf("Received product.created event for product %d", event.Product.ID)

	// Create the product in the local database
	product := localModels.Product{}
	product.ID = event.Product.ID

	if err := h.db.Create(&product).Error; err != nil {
		log.Printf("Error creating product in DB: %v", err)
		return err
	}

	log.Printf("Successfully created product %d in local database", product.ID)
	return nil
}

// HandleProductUpdated handles the product.updated event
func (h *ProductEventHandlers) HandleProductUpdated(body []byte) error {
	var event events.ProductEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling product.updated event: %v", err)
		return err
	}

	log.Printf("Received product.updated event for product %d", event.Product.ID)

	// Update the product in the local database
	product := localModels.Product{}
	product.ID = event.Product.ID

	if err := h.db.Save(&product).Error; err != nil {
		log.Printf("Error updating product in DB: %v", err)
		return err
	}

	log.Printf("Successfully updated product %d in local database", product.ID)
	return nil
}

// HandleProductDeleted handles the product.deleted event
func (h *ProductEventHandlers) HandleProductDeleted(body []byte) error {
	var event events.ProductEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling product.deleted event: %v", err)
		return err
	}

	log.Printf("Received product.deleted event for product %d", event.Product.ID)

	// Delete the product from the local database
	if err := h.db.Delete(&localModels.Product{}, event.Product.ID).Error; err != nil {
		log.Printf("Error deleting product from DB: %v", err)
		return err
	}

	log.Printf("Successfully deleted product %d from local database", event.Product.ID)
	return nil
}
