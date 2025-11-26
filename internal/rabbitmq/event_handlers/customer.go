package event_handlers

import (
	"encoding/json"
	"log"

	"github.com/PayeTonKawa-EPSI-2025/Common-V2/events"
	localModels "github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/models"
	"gorm.io/gorm"
)

// CustomerEventHandlers provides handlers for customer-related events
type CustomerEventHandlers struct {
	db *gorm.DB
}

// NewCustomerEventHandlers creates a new customer event handlers instance
func NewCustomerEventHandlers(db *gorm.DB) *CustomerEventHandlers {
	return &CustomerEventHandlers{db: db}
}

// HandleCustomerCreated handles the customer.created event
func (h *CustomerEventHandlers) HandleCustomerCreated(body []byte) error {
	var event events.CustomerEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling customer.created event: %v", err)
		return err
	}

	log.Printf("Received customer.created event for customer %d", event.Customer.ID)

	// Create the customer in the local database
	customer := localModels.Customer{}
	customer.ID = event.Customer.ID

	if err := h.db.Create(&customer).Error; err != nil {
		log.Printf("Error creating customer in DB: %v", err)
		return err
	}

	log.Printf("Successfully created customer %d in local database", customer.ID)
	return nil
}

// HandleCustomerUpdated handles the customer.updated event
func (h *CustomerEventHandlers) HandleCustomerUpdated(body []byte) error {
	var event events.CustomerEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling customer.updated event: %v", err)
		return err
	}

	log.Printf("Received customer.updated event for customer %d", event.Customer.ID)

	// Update the customer in the local database
	customer := localModels.Customer{}
	customer.ID = event.Customer.ID

	if err := h.db.Save(&customer).Error; err != nil {
		log.Printf("Error updating customer in DB: %v", err)
		return err
	}

	log.Printf("Successfully updated customer %d in local database", customer.ID)
	return nil
}

// HandleCustomerDeleted handles the customer.deleted event
func (h *CustomerEventHandlers) HandleCustomerDeleted(body []byte) error {
	var event events.CustomerEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error unmarshaling customer.deleted event: %v", err)
		return err
	}

	log.Printf("Received customer.deleted event for customer %d", event.Customer.ID)

	// Delete the customer from the local database
	if err := h.db.Delete(&localModels.Customer{}, event.Customer.ID).Error; err != nil {
		log.Printf("Error deleting customer from DB: %v", err)
		return err
	}

	log.Printf("Successfully deleted customer %d from local database", event.Customer.ID)
	return nil
}
