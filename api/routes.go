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
	err := db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)
	if err != nil {
		FailOnError(l, err, "Failed to add recipe")
		return NewInternalServerError(err)
	}
	// TODO Change the response to return the recipe
	return c.JSON(http.StatusNoContent, recipe)
}

func (api *ApiHandler) getIngredient(c echo.Context) error {
	l := logger.WithField("request", "getIngredient")
	l.Debug("Getting Ingredient")

	ingredientId := c.Param("id")
	recipeId := c.Param("recipe_id")

	ingredient, err := db.GetIngredient(api.rdb, "1", ingredientId, recipeId)

	if err != nil {
		FailOnError(l, err, "Failed to get ingredient")
		return NewNotFoundError(err)
	}

	return c.JSON(http.StatusOK, ingredient)
}

func (api *ApiHandler) removeIngredient(c echo.Context) error {
	l := logger.WithField("request", "removeIngredient")

	l.Debug("Removing Ingredient")

	ingredientId := c.Param("id")
	recipeId := c.Param("recipe_id")
	allQuantities := c.QueryParam("all")
	removeAll := false
	// If all is true, remove all quantities of the ingredient
	if allQuantities == "true" {
		removeAll = true
	}

	err := db.RemoveIngredient(api.rdb, "1", ingredientId, recipeId, removeAll)
	if err != nil {
		FailOnError(l, err, "Failed to remove ingredient")
		return NewNotFoundError(err)
	}

	l.WithFields(logrus.Fields{
		ingredientId: ingredientId,
		recipeId:     recipeId,
	}).Info("Removed Ingredient")

	// Then remove the ingredients from the recipe if recipeId is not empty
	if recipeId != "" {
		err = db.RemoveIngredientFromRecipe(api.rdb, "1", ingredientId, recipeId)
		if err != nil {
			FailOnError(l, err, "Failed to remove ingredient from recipe")
			return NewInternalServerError(err)
		}
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (api *ApiHandler) getRecipe(c echo.Context) error {
	l := logger.WithField("request", "getRecipe")

	l.Debug("Getting Recipe")

	recipeId := c.Param("id")

	recipe, err := db.GetRecipe(api.rdb, "1", recipeId)
	if err != nil {
		FailOnError(l, err, "Failed to get recipe")
		return NewNotFoundError(err)
	}
	return c.JSON(http.StatusOK, recipe)

}

func (api *ApiHandler) removeRecipe(c echo.Context) error {
	l := logger.WithField("request", "removeRecipe")

	l.Debug("Removing Recipe")

	recipeId := c.Param("id")

	err := db.RemoveRecipe(api.rdb, "1", recipeId)
	if err != nil {
		FailOnError(l, err, "Failed to remove recipe")
		return NewNotFoundError(err)
	}
	return c.JSON(http.StatusNoContent, nil)
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
