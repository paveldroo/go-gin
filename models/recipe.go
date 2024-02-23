package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	//swagger:ignore
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Title     string             `json:"title" bson:"title"`
	Thumbnail string             `json:"thumbnail" bson:"thumbnail"`
	URL       string             `json:"url" bson:"url"`
}
