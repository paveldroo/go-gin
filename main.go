// Recipes API
//
// # This is a sample recipes API
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Pavel Droo <123@213.com>
//
// Consumes:
// - application.json
//
// Produces:
// - application.json
// swagger:meta
package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	//swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	defer cur.Close(ctx)

	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

// swagger:operation GET /recipes/{:id} recipes getOneRecipe
// Search recipes by tag
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of a recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'404':
//		description: Invalid recipe ID
func GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var recipe Recipe
	cur := collection.FindOne(ctx, bson.M{"_id": objectId})
	err = cur.Decode(&recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, recipe)
}

// swagger:operation POST /recipes recipes newRecipe
// Creates new recipe
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'400':
//		description: Invalid input
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err = collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'400':
//		description: Invalid input
//	'404':
//		description: Invalid recipe ID
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectId},
		bson.D{{"$set", bson.D{
			{"name", recipe.Name},
			{"instructions", recipe.Instructions},
			{"ingredients", recipe.Ingredients},
			{"tags", recipe.Tags},
		}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been updated",
	})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'404':
//		description: Invalid recipe ID
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	dResult, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	if dResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

// swagger:operation GET /recipes/search recipes searchRecipes
// Search recipes by tag
// ---
// parameters:
//   - name: id
//     in: query
//     description: recipe tag
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//		description: Successful operation
//	'404':
//		description: Invalid recipe ID
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")

	cur, err := collection.Find(ctx, bson.M{"tags": tag})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	var recipes []Recipe

	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	if len(recipes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Recipes not found",
		})
		return
	}

	c.JSON(http.StatusOK, recipes)
}

var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection

func init() {
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB")
}

func main() {
	router := gin.Default()
	router.GET("/recipes", ListRecipesHandler)
	router.GET("/recipes/:id", GetOneRecipeHandler)
	router.POST("/recipes", NewRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	log.Fatal(router.Run())
}
