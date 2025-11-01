package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventType represents the type of event being published
type EventType string

const (
	OrderCreated EventType = "order.created"
	OrderUpdated EventType = "order.updated"
	OrderDeleted EventType = "order.deleted"
)

// OrderEvent represents the structure of a order event
type OrderEvent struct {
	Type      EventType    `json:"type"`
	Order     models.Order `json:"order"`
	Timestamp time.Time    `json:"timestamp"`
}

// PublishOrderEvent publishes a order event to RabbitMQ
func PublishOrderEvent(ch *amqp.Channel, eventType EventType, order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := OrderEvent{
		Type:      eventType,
		Order:     order,
		Timestamp: time.Now(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return err
	}

	// Use a routing key based on the event type
	routingKey := string(eventType)

	err = ch.PublishWithContext(
		ctx,
		"events", // exchange
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("Published %s event for order %d", eventType, order.ID)
	return nil
}
