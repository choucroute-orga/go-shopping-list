package api

import (
	"net/http"
	"shopping-list/db"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("context", "api/routes")

func (api *ApiHandler) getAliveStatus(c echo.Context) error {
	l := logger.WithField("request", "getAliveStatus")
	status := NewHealthResponse(LiveStatus)
	if err := c.Bind(status); err != nil {
		FailOnError(l, err, "Response binding failed")
		return NewInternalServerError(err)
	}
	l.WithFields(logrus.Fields{
		"action": "getStatus",
		"status": status,
	}).Debug("Health Status ping")

	return c.JSON(http.StatusOK, &status)
}

func (api *ApiHandler) getReadyStatus(c echo.Context) error {
	l := logger.WithField("request", "getReadyStatus")
	err := api.rdb.Ping(c.Request().Context()).Err()
	if err != nil {
		FailOnError(l, err, "Redis ping failed")
		return c.JSON(http.StatusServiceUnavailable, NewHealthResponse(NotReadyStatus))
	}

	return c.JSON(http.StatusOK, NewHealthResponse(ReadyStatus))
}

func (api *ApiHandler) getShoppingList(c echo.Context) error {
	l := logger.WithField("request", "getShoppingList")

	l.Debug("Getting Shopping List")

	ingredients, err := db.GetShoppingList(api.rdb, "1")

	if err != nil {
		FailOnError(l, err, "Failed to get shopping list")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, ingredients)
}

func (api *ApiHandler) addRecipe(c echo.Context) error {
	l := logger.WithField("request", "addRecipe")

	l.Debug("Adding Recipe")
	recipe := new(AddRecipeRequest)

	if err := c.Bind(recipe); err != nil {
		FailOnError(l, err, "Binding recipe failed")
		return NewBadRequestError(err)
	}
	if err := c.Validate(recipe); err != nil {
		FailOnError(l, err, "Validation failed")
		return NewBadRequestError(err)
	}
	l.Info("Validating Recipe " + recipe.ID)
	recipeDb, ingredientsDb := NewRecipe(recipe)
	ingredientRes, err := db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)
	if err != nil {
		FailOnError(l, err, "Failed to add recipe")
		return NewInternalServerError(err)
	}
	// TODO Change the response to return the recipe
	return c.JSON(http.StatusOK, ingredientRes)
}

func (api *ApiHandler) addIngredient(c echo.Context) error {
	l := logger.WithField("request", "addIngredient")

	l.Debug("Adding Ingredient")

	ingredient := new(AddIngredientRequest)

	if err := c.Bind(ingredient); err != nil {
		FailOnError(l, err, "Binding ingredient failed")
		return NewBadRequestError(err)
	}
	if err := c.Validate(ingredient); err != nil {
		FailOnError(l, err, "Validation failed")
		return NewBadRequestError(err)
	}

	recipeId := ""
	recipeId = c.Param("recipe_id")

	ingredientDb := NewIngredient(ingredient, recipeId)
	ingredientRes, err := db.AddIngredient(api.rdb, "1", ingredient.ID, *ingredientDb)
	if err != nil {
		FailOnError(l, err, "Failed to add ingredient")
		return NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, ingredientRes)
}
