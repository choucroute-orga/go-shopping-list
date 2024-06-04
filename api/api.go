package api

import (
	"shopping-list/configuration"
	"shopping-list/validation"

	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type ApiHandler struct {
	conf       *configuration.Configuration
	rdb        *redis.Client
	amqp       *amqp.Connection
	validation *validation.Validation
}

func NewApiHandler(conf *configuration.Configuration, rdb *redis.Client, amqp *amqp.Connection) *ApiHandler {
	handler := ApiHandler{
		conf:       conf,
		rdb:        rdb,
		amqp:       amqp,
		validation: validation.New(conf),
	}
	return &handler
}

func (api *ApiHandler) Register(v1 *echo.Group, conf *configuration.Configuration) {

	health := v1.Group("/health")
	health.GET("/alive", api.getAliveStatus)
	health.GET("/live", api.getAliveStatus)
	health.GET("/ready", api.getReadyStatus)
	ingredients := v1.Group("/ingredient")
	ingredients.GET("/:id", api.getIngredient)
	ingredients.POST("/:id", api.addIngredient)
	ingredients.DELETE("/:id", api.removeIngredient)
	recipe := v1.Group("/recipe")
	recipe.GET("/:id", api.getRecipe)
	recipe.GET("/:recipe_id/ingredient/:id", api.getIngredient)
	recipe.POST("", api.addRecipe)
	recipe.POST("/:recipe_id/ingredient/:id", api.addIngredient)
	recipe.DELETE("/:id", api.removeRecipe)
	recipe.DELETE("/:recipe_id/ingredient/:id", api.removeIngredient)
	shoppingList := v1.Group("/shopping-list")
	shoppingList.GET("", api.getShoppingList)
}
