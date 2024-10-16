package messages

type AddIngredient struct {
	ID     string  `param:"id" json:"id" validate:"required"`
	Amount float64 `json:"amount" validate:"required,min=0.1"`
	Unit   string  `json:"unit" validate:"oneof=i is cup tbsp tsp g kg"`
}

type AddRecipe struct {
	ID          string          `json:"id" validate:"required"`
	Ingredients []AddIngredient `json:"ingredients" validate:"required,dive,required"`
}

type AddIngredientMessage struct {
	ID     string  `json:"id" validate:"required"`
	UserID string  `json:"userId" validate:"required"`
	Amount float64 `json:"amount" validate:"required,min=0.1"`
	Unit   string  `json:"unit" validate:"oneof=i is cup tbsp tsp g kg"`
}
