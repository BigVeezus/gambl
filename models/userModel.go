package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	User_name     string             `json:"user_name" validate:"required"`
	Password      string             `json:"password" validate:"required,min=6"`
	Email         string             `json:"email" validate:"email,required"`
	Country       string             `json:"country"`
	User_type     string             `json:"user_type"`
	Score         int                `json:"score"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
}
