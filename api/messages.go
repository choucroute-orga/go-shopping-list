package api

import (
	"encoding/json"
	"shopping-list/db"
	"shopping-list/messages"
	"time"
)

func (api *ApiHandler) ConsumeMessages() {
	for {
		api.consumesMessages()
		time.Sleep(time.Microsecond)
	}
}

func (api *ApiHandler) consumesMessages() {

	q := messages.GetShoppingListQueue(api.amqp)
	if q == nil {
		return
	}

	ch, err := api.amqp.Channel()
	if err != nil {
		logger.WithError(err).Error("Failed to open a channel")
	}

	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name,          // queue
		"shopping-list", // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)

	if err != nil {
		logger.WithError(err).Error("Failed to register a consumer")
	}

	for d := range msgs {
		// Transform the message into a string
		// str := string(d.Body)
		logger.WithField("message", string(d.Body)).Info("Received a message")
		recipe := new(AddRecipeRequest)
		err := json.Unmarshal(d.Body, recipe)
		if err != nil {
			logger.WithError(err).Error("Failed to unmarshal the message")
		}
		if err := api.validation.Validate.Struct(recipe); err != nil {
			logger.WithError(err).Error("Failed to validate the message")
			break
		}
		recipeDb, ingredientsDb := NewRecipe(recipe)
		err = db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)
		if err != nil {
			logger.WithError(err).Error("Failed to insert the recipe")
		}
	}
}
