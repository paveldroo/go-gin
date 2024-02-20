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
	"github.com/go-redis/redis/v8"
	"github.com/paveldroo/go-gin/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
)

var authHandler *handlers.AuthHandler
var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()

	// redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(ctx)
	fmt.Println(status)

	// mongodb
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	authHandler = &handlers.AuthHandler{}
}

func main() {
	router := gin.Default()
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.POST("/signin", authHandler.SignInHandler)

	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	authorized.GET("/recipes/:id", recipesHandler.GetOneRecipeHandler)
	authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
	authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	authorized.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	log.Fatal(router.Run())
}
