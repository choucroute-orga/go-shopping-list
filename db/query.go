package db

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "db/query",
})

func GetIngredient(rdb *redis.Client, userId string, ingredientId string, recipeIds ...string) (*Ingredient, error) {

	ctx := context.Background()
	res, err := rdb.Get(ctx, userId+":ingredient:"+ingredientId).Result()
	if err != nil {
		logger.WithError(err).Error("Failed to get ingredient: " + ingredientId)
		return nil, err
	}

	var quantities []Quantity
	err = json.Unmarshal([]byte(res), &quantities)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal ingredient: " + ingredientId)
		return nil, err
	}

	// Filter the quantities by recipeId if it is provided
	recipeId := ""

	if len(recipeIds) > 0 {
		recipeId = recipeIds[0]
	}
	if recipeId != "" {
		// Filter the quantities by recipeId
		var filteredQuantities []Quantity
		for _, quantity := range quantities {
			if quantity.RecipeID == recipeId {
				filteredQuantities = append(filteredQuantities, quantity)
			}
		}
		quantities = filteredQuantities
	}

	// Otherwise, return all the quantities

	return &Ingredient{
		ID:         ingredientId,
		Quantities: quantities,
	}, nil
}

func GetIngredientRecipe(rdb *redis.Client, userId string, ingredientId string, recipeId string) (*Ingredient, error) {
	return GetIngredient(rdb, userId, ingredientId, recipeId)
}

func GetRecipe(rdb *redis.Client, userId string, recipeId string) (*Recipe, error) {

	ctx := context.Background()
	res, err := rdb.Get(ctx, userId+":recipe:"+recipeId).Result()

	if err != nil {
		logger.WithError(err).Error("Failed to get recipe: " + recipeId)
		return nil, err
	}

	var ingredientsID []string
	err = json.Unmarshal([]byte(res), &ingredientsID)
	if err != nil {
		logger.WithError(err).Error("Failed to unmarshal recipe: " + recipeId)
		return nil, err
	}
	return &Recipe{
		IngredientsID: ingredientsID,
	}, nil

}

// TODO: Add a counter of time to check how many times the recipe is used
func AddRecipe(rdb *redis.Client, userId string, recipeID string, recipe *Recipe, ingredients *[]Ingredient) error {

	ctx := context.Background()

	recipeSaved, _ := GetRecipe(rdb, userId, recipeID)

	// Save the recipe if it does not exist
	if recipeSaved == nil {
		ingredientsID, err := json.Marshal(recipe.IngredientsID)
		if err != nil {
			logger.WithField("recipe", recipe).WithError(err).Error("Failed to marshal recipe")
			return err
		}
		err = rdb.Set(ctx, userId+":recipe:"+recipeID, ingredientsID, 0).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to set recipe: " + recipeID)
			return err
		}
	}

	var err error
	// Save the ingredients
	for i, ingredient := range *ingredients {
		_, err = AddIngredient(rdb, userId, recipe.IngredientsID[i], ingredient)
	}

	return err
}

func GetShoppingList(rdb *redis.Client, userId string) (*[]Ingredient, error) {
	ctx := context.Background()
	res, err := rdb.Keys(ctx, userId+":ingredient:*").Result()
	if err != nil {
		logger.WithError(err).Error("Failed to get ingredients")
		return nil, err
	}
	var ingredients []Ingredient
	for _, key := range res {
		ingredientID := key[len(userId)+12:]
		ingredient, err := GetIngredient(rdb, userId, ingredientID)
		if err != nil {
			return nil, err
		}
		ingredient.ID = ingredientID
		ingredients = append(ingredients, *ingredient)
	}

	return &ingredients, nil
}

func RemoveRecipe(rdb *redis.Client, userId string, recipeId string) error {
	r, err := GetRecipe(rdb, userId, recipeId)
	if err != nil {
		return err
	}
	for _, ingredientID := range r.IngredientsID {
		err := RemoveIngredient(rdb, userId, ingredientID, recipeId, false)
		if err != nil {
			return err
		}
	}

	// Remove the recipe
	ctx := context.Background()
	err = rdb.Del(ctx, userId+":recipe:"+recipeId).Err()
	if err != nil {
		logger.WithError(err).Error("Failed to delete recipe: " + recipeId)
		return err
	}
	return nil

}

func RemoveIngredientFromRecipe(rdb *redis.Client, userId string, ingredientID string, recipeId string) error {

	r, err := GetRecipe(rdb, userId, recipeId)
	if err != nil {
		return err
	}
	// Check the ingedientID is in the recipe
	newIngredientsID := make([]string, 0)
	for _, id := range r.IngredientsID {
		if id != ingredientID {
			newIngredientsID = append(newIngredientsID, id)
		}
	}

	if len(newIngredientsID) == 0 {
		err := RemoveRecipe(rdb, userId, recipeId)
		return err
	} else {
		// Save the updated ingredients
		ingredientsID, err := json.Marshal(newIngredientsID)
		if err != nil {
			logger.WithField("recipe", r).WithError(err).Error("Failed to marshal recipe")
			return err
		}
		ctx := context.Background()
		err = rdb.Set(ctx, userId+":recipe:"+recipeId, ingredientsID, 0).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to set recipe: " + recipeId)
			return err
		}
	}

	return nil
}

func RemoveIngredient(rdb *redis.Client, userId string, ingredientID string, recipeId string, removeAll bool) error {
	ctx := context.Background()
	ingredient, err := GetIngredient(rdb, userId, ingredientID)

	if removeAll {
		err = rdb.Del(ctx, userId+":ingredient:"+ingredientID).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to delete ingredient: " + ingredientID)
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}
	for i, quantity := range ingredient.Quantities {
		if quantity.RecipeID == recipeId {
			ingredient.Quantities = append(ingredient.Quantities[:i], ingredient.Quantities[i+1:]...)
			break
		}
	}

	// If quantities is empty, we remove the ingredient
	if len(ingredient.Quantities) == 0 {
		err = rdb.Del(ctx, userId+":ingredient:"+ingredientID).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to delete ingredient: " + ingredientID)
			return err
		}
		return nil
	} else {

		// Save the updated quantities
		quantities, err := json.Marshal(ingredient.Quantities)
		if err != nil {
			logger.WithField("ingredient", ingredient).WithError(err).Error("Failed to marshal ingredient")
			return err
		}

		err = rdb.Set(ctx, userId+":ingredient:"+ingredientID, quantities, 0).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to set ingredient: " + ingredientID)
			return err
		}
	}
	return nil

}

// TODO Refactor the function
func AddIngredient(rdb *redis.Client, userId string, ingredientID string, ingredient Ingredient) (*Ingredient, error) {

	ctx := context.Background()
	ingredientSaved, _ := GetIngredient(rdb, userId, ingredientID)

	for _, quantity := range ingredient.Quantities {

		if ingredientSaved != nil {

			// If we find the same ingredient, we add the quantity to the existing one if the unit and the recipeID are the same
			added := false
			for i, savedQuantity := range ingredientSaved.Quantities {
				if quantity.Unit == savedQuantity.Unit && quantity.RecipeID == savedQuantity.RecipeID {
					ingredientSaved.Quantities[i].Amount += quantity.Amount
					added = true
					break
				}
			}

			// If we do not find the same ingredient, we add the quantity to the existing one
			if !added {
				ingredientSaved.Quantities = append(ingredientSaved.Quantities, quantity)
			}

		} else {
			ingredientSaved = &Ingredient{
				Quantities: ingredient.Quantities,
			}

		}

	}

	// Add the ingredient to the database

	quantities, err := json.Marshal(ingredientSaved.Quantities)
	if err != nil {
		logger.WithField("ingredient", ingredientSaved).WithError(err).Error("Failed to marshal ingredient")
		return nil, err
	}

	err = rdb.Set(ctx, userId+":ingredient:"+ingredientID, quantities, 0).Err()
	if err != nil {
		logger.WithError(err).Error("Failed to set ingredient: " + ingredientID)
		return nil, err
	}

	return ingredientSaved, nil
}
