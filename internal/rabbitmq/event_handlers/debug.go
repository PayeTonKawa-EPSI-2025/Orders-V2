package event_handlers

import (
	"encoding/json"
	"log"

	"github.com/PayeTonKawa-EPSI-2025/Common-V2/events"
)

// DebugEventHandlers provides handlers for debugging purposes
type DebugEventHandlers struct{}

// NewDebugEventHandlers creates a new debug event handlers instance
func NewDebugEventHandlers() *DebugEventHandlers {
	return &DebugEventHandlers{}
}

// HandleAllEvents is a catch-all handler for debugging purposes
// Useful during development, can be removed in production
func (h *DebugEventHandlers) HandleAllEvents(body []byte) error {
	var generic events.GenericEvent
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
}
