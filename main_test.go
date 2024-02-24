package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/paveldroo/go-gin/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestListRecipesHandler(t *testing.T) {
	r := SetupRouter()
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	req, _ := http.NewRequest("GET", "/recipes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var recipes []models.Recipe
	json.Unmarshal([]byte(w.Body.String()), &recipes)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 392, len(recipes))
}

func TestNewRecipeHandler(t *testing.T) {
	r := SetupRouter()
	r.POST("/recipes", recipesHandler.NewRecipeHandler)
	r.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	recipe := models.Recipe{
		Title: "New York Pizza",
	}
	jsonValue, _ := json.Marshal(recipe)
	req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(jsonValue))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal([]byte(w.Body.String()), &recipe)
	req, _ = http.NewRequest("DELETE", "/recipes/"+recipe.ID.Hex(), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOneRecipeHandler(t *testing.T) {
	r := SetupRouter()
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.GET("/recipes/:id", recipesHandler.GetOneRecipeHandler)

	req, _ := http.NewRequest("GET", "/recipes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var recipes []models.Recipe
	json.Unmarshal([]byte(w.Body.String()), &recipes)
	recipeId := recipes[0].ID.Hex()
	req, _ = http.NewRequest("GET", "/recipes/"+recipeId, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var recipe models.Recipe
	json.Unmarshal([]byte(w.Body.String()), &recipe)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, recipe, recipes[0])
}

func TestUpdateRecipeHandler(t *testing.T) {
	r := SetupRouter()
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)

	req, _ := http.NewRequest("GET", "/recipes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var recipes []models.Recipe
	json.Unmarshal([]byte(w.Body.String()), &recipes)

	recipe := models.Recipe{
		ID:    recipes[0].ID,
		Title: recipes[0].Title,
	}
	jsonValue, _ := json.Marshal(recipe)
	req, _ = http.NewRequest("PUT", "/recipes/"+recipe.ID.Hex(), bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
