package db

type Quantity struct {
	Amount   float64 `json:"amount"`
	Unit     string  `json:"unit"`
	RecipeID string  `json:"recipe_id" validate:"omitempty"`
}

type Ingredient struct {
	Quantities []Quantity `json:"quantities"`
}

type Recipe struct {
	IngredientsID []string `json:"ingredients"`
}
