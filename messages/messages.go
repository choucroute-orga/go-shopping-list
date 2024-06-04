package messages

import (
	"shopping-list/configuration"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "messages",
})

const (
	AddRecipesShoppingList = "add-recipes-shopping-list"
)

func New(conf *configuration.Configuration) *amqp.Connection {
	logger.Info("Connecting to RabbitMQ... " + conf.RabbitURI)
	conn, err := amqp.Dial(conf.RabbitURI)
	if err != nil {
		panic(err)
	}
	logger.Info("Connected to RabbitMQ!")
	return conn
}

func GetShoppingListQueue(conn *amqp.Connection) *amqp.Queue {
	ch, err := conn.Channel()
	if err != nil {
		logger.WithError(err).Error("Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		AddRecipesShoppingList, // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		logger.WithError(err).Error("Failed to declare a queue")
	}

	return &q
}
