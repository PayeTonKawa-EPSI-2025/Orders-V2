package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EventHandler is a function type that processes RabbitMQ events
type EventHandler func(body []byte) error

// EventRouter routes events to specific handlers based on routing keys
type EventRouter struct {
	handlers map[string]EventHandler
}

// NewEventRouter creates a new event router
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[string]EventHandler),
	}
}

// RegisterHandler registers a handler for a specific routing key (event type)
func (r *EventRouter) RegisterHandler(routingKey string, handler EventHandler) {
	r.handlers[routingKey] = handler
}

// handleMessage routes the message to the appropriate handler
func (r *EventRouter) handleMessage(d amqp.Delivery) {
	log.Printf("Received message with routing key: %s", d.RoutingKey)

	// Find the appropriate handler for this routing key
	handler, exists := r.handlers[d.RoutingKey]
	if !exists {
		// Check if there's a wildcard handler that matches
		for pattern, wildcardHandler := range r.handlers {
			if matchesWildcard(pattern, d.RoutingKey) {
				handler = wildcardHandler
				exists = true
				break
			}
		}
	}

	if !exists {
		log.Printf("No handler registered for routing key: %s", d.RoutingKey)
		// Acknowledge the message to remove it from the queue
		d.Ack(false)
		return
	}

	// Process the message with the handler
	err := handler(d.Body)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		// You might want to implement retries or dead letter queue here
		// For now, we'll just acknowledge the message to remove it from the queue
		d.Ack(false)
		return
	}

	// Successfully processed the message, acknowledge it
	d.Ack(false)
}

// matchesWildcard checks if a routing key matches a pattern with wildcards
func matchesWildcard(pattern, routingKey string) bool {
	// Simple implementation: only supports * at the end
	if pattern == "#" {
		return true // # matches anything
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(routingKey) >= len(prefix) && routingKey[:len(prefix)] == prefix
	}

	return pattern == routingKey
}

// StartListening sets up a consumer to listen for RabbitMQ events
func StartListening(ch *amqp.Channel, router *EventRouter) (string, error) {
	// Declare a queue with a random name
	q, err := ch.QueueDeclare(
		"",    // name (empty for auto-generated name)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return "", err
	}

	// Bind the queue to the exchange with routing keys
	for routingKey := range router.handlers {
		if routingKey == "#" {
			// Special case: bind to all messages
			err = ch.QueueBind(
				q.Name,   // queue name
				"#",      // routing key
				"events", // exchange
				false,    // no-wait
				nil,      // arguments
			)
		} else if len(routingKey) > 0 && routingKey[len(routingKey)-1] == '*' {
			// Wildcard binding
			bindingKey := routingKey[:len(routingKey)-1] + "#"
			err = ch.QueueBind(
				q.Name,     // queue name
				bindingKey, // routing key with AMQP wildcard
				"events",   // exchange
				false,      // no-wait
				nil,        // arguments
			)
		} else {
			// Standard binding
			err = ch.QueueBind(
				q.Name,     // queue name
				routingKey, // routing key
				"events",   // exchange
				false,      // no-wait
				nil,        // arguments
			)
		}

		if err != nil {
			return "", err
		}
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (false means manual acknowledgment)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return "", err
	}

	// Start a goroutine to process messages
	go func() {
		for d := range msgs {
			router.handleMessage(d)
		}
		log.Println("RabbitMQ consumer channel closed")
	}()

	log.Printf("Started listening for events on queue %s", q.Name)
	return q.Name, nil
}
