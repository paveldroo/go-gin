package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/paveldroo/go-gin/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
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
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()
	if errors.Is(err, redis.Nil) || val == "[]" {
		log.Printf("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		defer cur.Close(handler.ctx)

		recipes := make([]models.Recipe, 0)
		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		handler.redisClient.Set(handler.ctx, "recipes", string(data), 0)

		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
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
func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var recipe models.Recipe
	cur := handler.collection.FindOne(handler.ctx, bson.M{"_id": objectId})
	err := cur.Decode(&recipe)
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
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del(handler.ctx, "recipes").Result()

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
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, err := handler.collection.UpdateOne(
		handler.ctx,
		bson.M{"_id": objectId},
		bson.D{{"$set", bson.D{
			{"title", recipe.Title},
			{"thumbnail", recipe.Thumbnail},
			{"url", recipe.URL},
		}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	fmt.Println("Remove data from Redis")
	handler.redisClient.Del(handler.ctx, "recipes")

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
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	dResult, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": objectId})
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
func (handler *RecipesHandler) SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")

	cur, err := handler.collection.Find(handler.ctx, bson.M{"tags": tag})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	var recipes []models.Recipe

	for cur.Next(handler.ctx) {
		var recipe models.Recipe
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
