package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)

}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := -1
	for i, r := range recipes {
		if r.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	recipe.ID = id
	recipes[index] = recipe

	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1

	for i, r := range recipes {
		if r.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")

	var matchedRecipes []Recipe

	for _, r := range recipes {
		for _, t := range r.Tags {
			if strings.EqualFold(t, tag) {
				matchedRecipes = append(matchedRecipes, r)
			}
		}
	}

	if len(matchedRecipes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Recipes not found",
		})
		return
	}

	c.JSON(http.StatusOK, matchedRecipes)
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	log.Fatal(router.Run())
}
