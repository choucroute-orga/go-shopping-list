package api

import (
	"encoding/json"
	"shopping-list/db"
	"shopping-list/messages"
	"time"

	"github.com/sirupsen/logrus"
)

func (api *ApiHandler) ConsumeMessages() {
	for {
		api.consumesMessages()
		time.Sleep(time.Microsecond)
	}
}

func (api *ApiHandler) consumesMessages() {
	l := logger.WithField("method", "consumesMessages")
	q := messages.GetShoppingListQueue(api.amqp)
	if q == nil {
		return
	}

	ch, err := api.amqp.Channel()
	if err != nil {
		l.WithError(err).Error("Failed to open a channel")
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
		l.WithError(err).Error("Failed to register a consumer")
	}

	for d := range msgs {
		// Transform the message into a string
		recipe := new(AddRecipeRequest)
		err := json.Unmarshal(d.Body, recipe)
		if err != nil {
			l.WithField("message", string(d.Body)).WithError(err).Error("Failed to unmarshal the message")
		}
		if err := api.validation.Validate.Struct(recipe); err != nil {
			l.WithField("message", string(d.Body)).WithError(err).Error("Failed to validate the message")
			break
		}
		recipeDb, ingredientsDb := NewRecipe(recipe)
		l.WithFields(logrus.Fields{
			"recipeId":         recipe.ID,
			"recipeUserId":     recipe.UserID,
			"ingredientsCount": len(recipe.Ingredients),
		}).Info("Received a message")

		l.WithField("ingredients", ingredientsDb).Debug("Creating shopping list with list of ingredients")
		err = db.AddRecipe(api.rdb, recipe.UserID, recipe.ID, recipeDb, ingredientsDb)
		if err != nil {
			l.WithError(err).Error("Failed to insert the recipe")
		}
	}
}
