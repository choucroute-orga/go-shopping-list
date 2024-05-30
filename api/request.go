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

type Recipe struct {
	ID          string       `json:"id" validate:"required"`
	Ingredients []Ingredient `json:"ingredients" validate:"required,dive,required"`
}

type ShoppingList struct {
	Recipes     []Recipe     `json:"recipes" validate:"required,dive,required"`
	Ingredients []Ingredient `json:"ingredients" validate:"required,dive,required"`
}

type AddIngredientRequest struct {
	ID string `param:"id" json:"id" validate:"required"`
	Quantity
}

func NewIngredient(AddIngredientRequest *AddIngredientRequest, recipeID string) *db.Ingredient {
	return &db.Ingredient{
		Quantities: []db.Quantity{
			{
				Amount:   AddIngredientRequest.Amount,
				Unit:     AddIngredientRequest.Unit,
				RecipeID: recipeID,
			},
		},
	}
}
