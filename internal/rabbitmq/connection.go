package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect() (*amqp.Connection, *amqp.Channel) {
	dsn := os.Getenv("RABBIT_DSN")

	conn, err := amqp.Dial(dsn)
	if err != nil {
		log.Fatalf("Erreur de connexion à RabbitMQ : %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Erreur ouverture channel : %v", err)
	}

	err = ch.ExchangeDeclare(
		"events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Erreur déclaration exchange : %v", err)
	}

	return conn, ch
}
