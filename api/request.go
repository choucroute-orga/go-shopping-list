package api

import "shopping-list/db"

type Quantity struct {
	Amount float64 `json:"amount" validate:"required,min=0.1"`
	Unit   string  `json:"unit" validate:"oneof=i is cs tbsp tsp g kg"`
}

type Ingredient struct {
	ID         string     `json:"id" validate:"required"`
	Quantities []Quantity `json:"quantities" validate:"required,dive,required"`
}

type ShoppingList struct {
	Recipes     []AddRecipeRequest `json:"recipes" validate:"required,dive,required"`
	Ingredients []Ingredient       `json:"ingredients" validate:"required,dive,required"`
}

type AddIngredientRequest struct {
	ID string `param:"id" json:"id" validate:"required"`
	Quantity
}

type AddRecipeRequest struct {
	ID          string                 `json:"id" validate:"required"`
	Ingredients []AddIngredientRequest `json:"ingredients" validate:"required,dive,required"`
}

func NewRecipe(addRecipeRequest *AddRecipeRequest) (*db.Recipe, *[]db.Ingredient) {
	recipe := &db.Recipe{
		IngredientsID: make([]string, len(addRecipeRequest.Ingredients)),
	}

	for i, ingredient := range addRecipeRequest.Ingredients {
		recipe.IngredientsID[i] = ingredient.ID
	}

	ingredients := make([]db.Ingredient, len(addRecipeRequest.Ingredients))

	for i, ingredient := range addRecipeRequest.Ingredients {

		ingredients[i] = db.Ingredient{
			Quantities: []db.Quantity{
				{
					Amount:   ingredient.Amount,
					Unit:     ingredient.Unit,
					RecipeID: addRecipeRequest.ID,
				},
			},
		}

	}

	return recipe, &ingredients
}

func NewIngredient(addIngredientRequest *AddIngredientRequest, recipeID string) *db.Ingredient {
	return &db.Ingredient{
		Quantities: []db.Quantity{
			{
				Amount:   addIngredientRequest.Amount,
				Unit:     addIngredientRequest.Unit,
				RecipeID: recipeID,
			},
		},
	}
}
