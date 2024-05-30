package api

import (
	"shopping-list/configuration"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type ApiHandler struct {
	conf *configuration.Configuration
	rdb  *redis.Client
}

func NewApiHandler(conf *configuration.Configuration, rdb *redis.Client) *ApiHandler {
	handler := ApiHandler{
		conf: conf,
		rdb:  rdb,
	}
	return &handler
}

func (api *ApiHandler) Register(v1 *echo.Group, conf *configuration.Configuration) {

	health := v1.Group("/health")
	health.GET("/alive", api.getAliveStatus)
	health.GET("/live", api.getAliveStatus)
	health.GET("/ready", api.getReadyStatus)
	ingredients := v1.Group("/ingredient")
	ingredients.POST("/:id", api.addIngredient)
	recipe := v1.Group("/recipe")
	recipe.POST("/:recipe_id/ingredient/:id", api.addIngredient)
	// shoppingList := v1.Group("/shopping-list")
}
