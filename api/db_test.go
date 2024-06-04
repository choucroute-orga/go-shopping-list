package api

import (
	"context"
	"log"
	"shopping-list/db"
	"shopping-list/tests"
	"testing"

	"github.com/sirupsen/logrus"
)

// You can use testing.T, if you want to test the code without benchmarking
func setupSuite(tb testing.TB) func(tb testing.TB) {
	log.Println("setup suite")

	// Return a function to teardown the test
	return func(tb testing.TB) {
		log.Println("teardown suite")
	}
}

// Almost the same as the above, but this one is for single test instead of collection of tests
func setupTest(tb testing.TB) (*ApiHandler, func(tb testing.TB)) {
	// log.Println("setup test")

	// return func(tb testing.TB) {
	// 	log.Println("teardown test")
	// }

	// Get a random port for the test, between 1024 and 65535
	exposedPort := "6379" //fmt.Sprint(rand.Intn(65525-1024) + 1024)
	redis, pool, resource := tests.InitTestDocker(exposedPort)
	conf := tests.GetDefaultConf()
	api := NewApiHandler(conf, redis, nil)
	return api, func(tb testing.TB) {
		tests.CloseTestDocker(redis, pool, resource)
	}
}

func TestRedisManipulation(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	t.Run("Add ingredient to the DB", func(t *testing.T) {
		api, teardownTest := setupTest(t)
		defer teardownTest(t)
		// Test the ping to the Redis server
		errr := api.rdb.Ping(context.Background()).Err()
		if errr != nil {
			t.Errorf("Failed to ping the Redis server: %v", errr)
		}

		i1 := db.Ingredient{
			ID: "000000000000000000000002",
			Quantities: []db.Quantity{
				{
					Amount: 100,
					Unit:   "g",
				},
			},
		}

		i2 := db.Ingredient{
			ID: "000000000000000000000002",
			Quantities: []db.Quantity{
				{
					Amount: 5,
					Unit:   "g",
				},
			},
		}

		_, err := db.AddIngredient(api.rdb, "1", i1.ID, i1)
		if err != nil {
			t.Errorf("Failed to add first ingredient: %v", err)
		}
		_, err = db.AddIngredient(api.rdb, "1", i2.ID, i2)
		if err != nil {
			t.Errorf("Failed to add 2nd ingredient: %v", err)
		}
		i, err := db.GetIngredient(api.rdb, "1", i1.ID)

		if len(i.Quantities) != 1 {
			t.Errorf("Failed to get the ingredient: %v", i)
		}
		if i.Quantities[0].Amount != 105 {
			t.Errorf("Failed to add the correct quantity to ingredient: %v", i)
		}
	})

	t.Run("Insert one Ingredient in the DB", func(t *testing.T) {

		// Test the ping to the Redis server

		logrus.SetLevel(logrus.DebugLevel)
		// l := logrus.WithField("test", "Insert one Recipe in the DB")
		api, teardownTest := setupTest(t)
		errr := api.rdb.Ping(context.Background()).Err()
		if errr != nil {
			t.Errorf("Failed to ping the Redis server: %v", errr)
		}
		i := db.Ingredient{
			ID: "000000000000000000000001",
			Quantities: []db.Quantity{
				{
					Amount: 1.0,
					Unit:   "g",
				},
			},
		}
		ii, err := db.AddIngredient(api.rdb, "1", i.ID, i)
		if err != nil {
			t.Errorf("Failed to add ingredient: %v", err)
		}
		// Print the original and the inserted ingredient
		logrus.WithFields(logrus.Fields{
			"original": i,
			"inserted": ii,
		}).Debug("Inserted ingredient")

		if ii.Quantities[0].Amount != i.Quantities[0].Amount || ii.Quantities[0].Unit != i.Quantities[0].Unit {
			t.Errorf("Failed to insert the ingredient: %v", ii)
		}
		ig, err := db.GetIngredient(api.rdb, "1", i.ID)
		if err != nil {
			t.Errorf("Failed to get ingredient: %v", err)
		}
		if ig.ID != i.ID || ig.Quantities[0].Amount != i.Quantities[0].Amount || ig.Quantities[0].Unit != i.Quantities[0].Unit {
			t.Errorf("Failed to get the ingredient: %v", ig)
		}
		defer teardownTest(t)
	})

	t.Run("Insert one Recipe in the DB", func(t *testing.T) {
		// Test the ping to the Redis server
		api, teardownTest := setupTest(t)
		errr := api.rdb.Ping(context.Background()).Err()
		if errr != nil {
			t.Errorf("Failed to ping the Redis server: %v", errr)
		}
		r := db.Recipe{
			IngredientsID: []string{"000000000000000000000001", "000000000000000000000002"},
		}
		ings := []db.Ingredient{
			{
				ID: "000000000000000000000001",
				Quantities: []db.Quantity{
					{
						Amount: 1.0,
						Unit:   "g",
					},
				},
			},
			{
				ID: "000000000000000000000002",
				Quantities: []db.Quantity{
					{
						Amount: 90.0,
						Unit:   "g",
					},
				},
			},
		}

		err := db.AddRecipe(api.rdb, "1", "000000000000000000000001", &r, &ings)

		if err != nil {
			t.Errorf("Failed to add recipe: %v", err)
		}

		recipe, err := db.GetRecipe(api.rdb, "1", "000000000000000000000001")
		if err != nil {
			t.Errorf("Failed to get recipe: %v", err)
		}

		if recipe.IngredientsID[0] != r.IngredientsID[0] || recipe.IngredientsID[1] != r.IngredientsID[1] {
			t.Errorf("Failed to get the recipe: %v", recipe)
		}

		// Check if the ingredients have the correct quantities and are associated with the recipe
		for i, id := range recipe.IngredientsID {
			ingredient, err := db.GetIngredient(api.rdb, "1", id)
			if err != nil {
				t.Errorf("Failed to get ingredient: %v", err)
			}
			if ingredient.ID != ings[i].ID || ingredient.Quantities[0].Amount != ings[i].Quantities[0].Amount || ingredient.Quantities[0].Unit != ings[i].Quantities[0].Unit {
				t.Errorf("Failed to get the ingredient: %v", ingredient)
			}
		}

		defer teardownTest(t)
	})

	t.Run("Check multiple insertion of the same ingredient", func(t *testing.T) {

		api, teardownTest := setupTest(t)

		i1 := db.Ingredient{
			ID: "000000000000000000000001",
			Quantities: []db.Quantity{
				{
					Amount: 100,
					Unit:   "g",
				},
			},
		}

		i2 := db.Ingredient{
			ID: "000000000000000000000002",
			Quantities: []db.Quantity{
				{
					Amount: 5,
					Unit:   "i",
				},
			},
		}

		r := db.Recipe{
			IngredientsID: []string{"000000000000000000000001", "000000000000000000000002"},
		}
		ings := []db.Ingredient{}
		db.AddIngredient(api.rdb, "1", i1.ID, i1)
		db.AddIngredient(api.rdb, "1", i2.ID, i2)

		i1.Quantities[0].Amount = 1
		i1.Quantities[0].Unit = "kg"
		i1.Quantities[0].RecipeID = "000000000000000000000001"

		ings = append(ings, i1)
		ings = append(ings, i2)

		// recipe := AddRecipeRequest {
		// 	ID: "000000000000000000000001",
		// 	Ingredients: []AddIngredientRequest{
		// 		{
		// 			ID: "000000000000000000000001",

		// }
		// recipeDb, ingredientsDb := NewRecipe(recipe)
		db.AddRecipe(api.rdb, "1", "000000000000000000000001", &r, &ings)

		// Check if the ingredients have the correct quantities and are associated with the recipe
		i, _ := db.GetIngredient(api.rdb, "1", i1.ID)

		if i.Quantities[0].Amount != 100 || i.Quantities[0].Unit != "g" {
			t.Errorf("Failed to add the correct quantity to ingredient: %v", i)
		}
		if i.Quantities[1].Amount != 1 || i.Quantities[1].Unit != "kg" || i.Quantities[1].RecipeID != "000000000000000000000001" {
			t.Errorf("Failed to add the correct quantity to ingredient: %v", i)
		}
		defer teardownTest(t)
	})

	t.Run("The recipe converts correctly to the DB", func(t *testing.T) {
		logrus.SetLevel(logrus.DebugLevel)
		api, teardownTest := setupTest(t)
		recipe := AddRecipeRequest{
			ID: "000000000000000000000001",
			Ingredients: []AddIngredientRequest{
				{
					ID: "000000000000000000000001",
					Quantity: Quantity{
						Amount: 1.0,
						Unit:   "g",
					},
				},
				{
					ID: "000000000000000000000002",
					Quantity: Quantity{
						Amount: 10.0,
						Unit:   "i",
					},
				},
			},
		}
		recipeDb, ingredientsDb := NewRecipe(&recipe)
		db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)

		r, _ := db.GetRecipe(api.rdb, "1", recipe.ID)

		if r.IngredientsID[0] != "000000000000000000000001" || r.IngredientsID[1] != "000000000000000000000002" {
			t.Errorf("Failed to convert the recipe to the DB: %v", r)
		}

		i, _ := db.GetIngredient(api.rdb, "1", "000000000000000000000001")

		if i.Quantities[0].Amount != 1 || i.Quantities[0].Unit != "g" {
			t.Errorf("Failed to convert the recipe to the DB: %v", i)
		}

		defer teardownTest(t)
	})

	t.Run("Delete the recipe, also delete ingredients", func(t *testing.T) {
		api, teardownTest := setupTest(t)

		recipe := AddRecipeRequest{
			ID: "000000000000000000000001",
			Ingredients: []AddIngredientRequest{
				{
					ID: "000000000000000000000001",
					Quantity: Quantity{
						Amount: 1.0,
						Unit:   "g",
					},
				},
				{
					ID: "000000000000000000000002",
					Quantity: Quantity{
						Amount: 10.0,
						Unit:   "i",
					},
				},
			},
		}
		recipeDb, ingredientsDb := NewRecipe(&recipe)
		db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)

		db.RemoveRecipe(api.rdb, "1", recipe.ID)

		r, err := db.GetRecipe(api.rdb, "1", recipe.ID)
		if err == nil || r != nil {
			t.Errorf("Failed to delete the recipe: %v", err)
		}

		i1, err := db.GetIngredient(api.rdb, "1", "000000000000000000000001")

		if err == nil || i1 != nil {
			t.Errorf("Failed to delete the ingredient: %v", err)
		}

		i2, err := db.GetIngredient(api.rdb, "1", "000000000000000000000002")

		if err == nil || i2 != nil {
			t.Errorf("Failed to delete the 2nd ingredient: %v", err)
		}

		defer teardownTest(t)
	})

	t.Run("Remove ingredient from the recipe that already exists", func(t *testing.T) {
		api, teardownTest := setupTest(t)

		recipe := AddRecipeRequest{
			ID: "000000000000000000000001",
			Ingredients: []AddIngredientRequest{
				{
					ID: "000000000000000000000001",
					Quantity: Quantity{
						Amount: 1.0,
						Unit:   "g",
					},
				},
				{
					ID: "000000000000000000000002",
					Quantity: Quantity{
						Amount: 10.0,
						Unit:   "i",
					},
				},
			},
		}
		recipeDb, ingredientsDb := NewRecipe(&recipe)
		db.AddRecipe(api.rdb, "1", recipe.ID, recipeDb, ingredientsDb)

		i1 := db.Ingredient{
			ID: "000000000000000000000001",
			Quantities: []db.Quantity{
				{
					Amount: 20.0,
					Unit:   "kg",
				},
			},
		}
		db.AddIngredient(api.rdb, "1", i1.ID, i1)

		db.RemoveIngredient(api.rdb, "1", i1.ID, "000000000000000000000001", false)

		err := db.RemoveIngredientFromRecipe(api.rdb, "1", "000000000000000000000001", recipe.ID)
		if err != nil {
			t.Errorf("Failed to remove the ingredient from the recipe: %v", err)
		}
		i, _ := db.GetIngredient(api.rdb, "1", i1.ID)

		if i.Quantities[0].Amount != 20 || i.Quantities[0].Unit != "kg" || len(i.Quantities) != 1 {
			t.Errorf("Failed to remove the ingredient from the ingredients list: %v", i)
		}

		i2, err := db.GetIngredient(api.rdb, "1", "000000000000000000000002")

		if i2.Quantities[0].Amount != 10 {
			t.Errorf("Failed to retrieve the correct quantity for the ingredient that should stay: %v", i2)
		}

		r, _ := db.GetRecipe(api.rdb, "1", recipe.ID)

		if r.IngredientsID[0] != "000000000000000000000002" {
			t.Errorf("Failed to remove the ingredient from the recipe: %v", r)
		}
		defer teardownTest(t)
	})

}
