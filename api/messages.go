package api

import (
	"context"
	"encoding/json"
	"shopping-list/db"
	"shopping-list/messages"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (api *ApiHandler) ConsumeAddIngredientMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down AddIngredientMessage consumer")
			return
		default:
			api.consumeAddIngredientMessage(ctx)
			time.Sleep(time.Microsecond)
		}
	}
}

func (api *ApiHandler) consumeAddIngredientMessage(ctx context.Context) {
	ctx, span := api.tracer.Start(ctx, "consumeAddIngredientMessage")
	defer span.End()
	l := logger.WithContext(ctx).WithField("method", "consumeAddIngredientMessage")
	q, ch, err := messages.GetIngredientShoppingListQueue(api.amqp)
	if err != nil {
		l.WithError(err).Error("Failed to get the shopping list queue")
		return
	}

	msgs, err := ch.Consume(
		q.Name,          // queue
		"shopping-list", // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	l = l.WithField("queue", q.Name)
	if err != nil {
		l.WithError(err).Error("Failed to register a consumer")
		return
	}

	l.Info("Consuming messages")

	for {
		select {
		case <-ctx.Done():
			l.Info("Shutting down AddIngredientMessage consumer")
			return
		case msg, ok := <-msgs:
			if !ok {
				l.Warn("Channel closed")
				return
			}
			messageCtx, messageSpan := api.tracer.Start(ctx, "processShoppingListAddRecipeMessage")
			startTime := time.Now()

			retryCount := 0
			maxRetries := 3
			var processErr error
			for retryCount < maxRetries {
				processErr = api.processAddIngredientMessage(messageCtx, l, msg)
				if processErr == nil {
					break
				}
				retryCount++
				l.WithError(processErr).WithField("retry", retryCount).Warn("Retrying message processing")
				time.Sleep(time.Second * time.Duration(retryCount)) // Exponential backoff
			}

			duration := time.Since(startTime)
			processStatus := "success"
			if processErr != nil {
				processStatus = "failure"
			}
			messageSpan.SetAttributes(
				attribute.Int("retries", retryCount),
				attribute.String("status", processStatus),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			)
			messageSpan.End()

			if processErr != nil {
				l.WithError(processErr).Error("Failed to process message after max retries")
				// Send to dead-letter queue
				err := ch.Publish(
					"",                           // exchange
					messages.DeadLetterQueueName, // routing key
					false,                        // mandatory
					false,                        // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        msg.Body,
						Headers: amqp.Table{
							"x-original-queue": messages.AddIngredientShoppingList,
							"x-error":          processErr.Error(),
						},
					},
				)
				if err != nil {
					l.WithError(err).Error("Failed to send message to dead-letter queue")
				}
			}

			msg.Ack(false)
		}
	}

}

func (api *ApiHandler) processAddIngredientMessage(ctx context.Context, l *logrus.Entry, msg amqp.Delivery) error {
	ctx, span := api.tracer.Start(ctx, "processAddIngredientMessage")
	defer span.End()
	
	l = l.WithContext(ctx).WithField("function", "processAddIngredientMessage")
	l.Info("Processing message")
	ingredient := new(messages.AddIngredientMessage)
	err := json.Unmarshal(msg.Body, ingredient)
	
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal the message")
		l.WithError(err).Error("Failed to unmarshal the message")
		return err
	}

	if err := api.validation.Validate.Struct(ingredient); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to validate the message")
		l.WithError(err).Error("Failed to validate the message")
		return err
	}

	quantities := make([]db.Quantity, 0)
	quantities = append(quantities, db.Quantity{
		Unit:   ingredient.Unit,
		Amount: ingredient.Amount,
	})
	ingredientDb := db.Ingredient{
		ID:         ingredient.ID,
		Quantities: quantities,
	}

	addCtx, addSpan := api.tracer.Start(ctx, "AddIngredientDB")
	ingInserted, err := db.AddIngredient(api.rdb, ingredient.UserID, ingredient.ID, ingredientDb)
	l = l.WithContext(addCtx).WithField("ingredientId", ingredient.ID)
	defer addSpan.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to insert the ingredient")
		l.WithError(err).Error("Failed to insert the ingredient")
		return err
	}

	l.WithFields(logrus.Fields{
		"ingredientId": ingInserted.ID,
		"userId":       ingredient.UserID,
		"quantites":    quantities,
	}).Info("Ingredient added to the shopping list")

	return nil
}

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
