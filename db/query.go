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

func getIngredient(rdb *redis.Client, userId string, ingredientId string) (*Ingredient, error) {

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
	return &Ingredient{
		Quantities: quantities,
	}, nil
}

func getRecipe(rdb *redis.Client, userId string, recipeId string) (*Recipe, error) {

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

func AddRecipe(rdb *redis.Client, userId string, recipeID string, recipe *Recipe, ingredients *[]Ingredient) (*Ingredient, error) {

	ctx := context.Background()

	// Check if the recipe already exists
	recipeSaved, _ := getRecipe(rdb, userId, recipeID)
	if recipeSaved != nil {
		// Save the recipe if it does not exist
		ingredientsID, err := json.Marshal(recipe.IngredientsID)
		if err != nil {
			logger.WithField("recipe", recipe).WithError(err).Error("Failed to marshal recipe")
			return nil, err
		}
		err = rdb.Set(ctx, userId+":recipe:"+recipeID, ingredientsID, 0).Err()
		if err != nil {
			logger.WithError(err).Error("Failed to set recipe: " + recipeID)
			return nil, err
		}
	}

	var ingredientSaved *Ingredient
	var err error
	// Save the ingredients
	for i, ingredient := range *ingredients {
		ingredientSaved, err = AddIngredient(rdb, userId, recipe.IngredientsID[i], ingredient)
	}

	return ingredientSaved, err
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
		ingredient, err := getIngredient(rdb, userId, ingredientID)
		if err != nil {
			return nil, err
		}
		ingredient.ID = ingredientID
		ingredients = append(ingredients, *ingredient)
	}

	return &ingredients, nil
}

// TODO Refactor the function
func AddIngredient(rdb *redis.Client, userId string, ingredientID string, ingredient Ingredient) (*Ingredient, error) {

	ctx := context.Background()
	ingredientSaved, _ := getIngredient(rdb, userId, ingredientID)

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

// func PostRecipe(rdb *redis.Client, userId string, recipe Recipe) error {

// }

// RecipeID is a unique identifier for the recipe
// IngredientsID is a list of unique identifiers for the ingredients
// Ingredients are stored in 2 ways:  and userId:ingredientID

// func GetShoppingList(rdb *redis.Client, userId string) (*[]Recipe, error) {

// 	ctx := context.Background()
// 	res, err := rdb.Keys(ctx, userId+":*").Result()
// 	if err != nil {
// 		logger.WithError(err).Error("Failed to get recipes")
// 		return nil, err
// 	}

// 	// check if there

// 	var recipes []Recipe
// 	for _, key := range res {

// 		// Check if the key match the regex for a recipe with 2 : in the key
// 		// If it does not match, it is an ingredient
// 		if strings.Count(key, ":") == 2 {

// 			// Extract the part in the middle of the key for the recipe ID
// 			recipeId := strings.Split(key, ":")[1]
// 			ingredients, err := GetIngredientsForRecipe(rdb, userId, recipeId)
// 			if err != nil {
// 				return nil, err
// 			}
// 			recipes = append(recipes, Recipe{
// 				Ingredients: *ingredients,
// 			})
// 		} else {
// 			ingredientID := strings.Split(key, ":")[1]
// 			ingredient, err := GetIngredient(rdb, userId, key)
// 			if err != nil {
// 				return nil, err
// 			}
// 			recipes = append(recipes, *recipe)

// 		}
// 	}
// 	return &recipes, nil
// }

// func GetIngredientsForRecipe(rdb *redis.Client, userId string, recipeId string) (*[]Ingredient, error) {

// 	ctx := context.Background()
// 	res, err := rdb.Keys(ctx, userId+":"+recipeId+":*").Result()

// 	// Unmarshall the List of Quantities stored in the recipe
// 	if err != nil {
// 		logger.WithError(err).Error("Failed to get recipe: " + recipeId)
// 		return nil, err
// 	}

// 	// Get the ingredients for the recipe
// 	var ingredients []Ingredient
// 	for _, key := range res {
// 		ingredientID := strings.Split(key, ":")[2]
// 		ingredient, err := GetIngredient(rdb, userId, ingredientID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		ingredients = append(ingredients, *ingredient)
// 	}
// 	return &ingredients, nil
// }
