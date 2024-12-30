package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type Reputation struct {
	ID            primitive.ObjectID `bson:"_id"`
	Tier          string             `json:"tier" validate:"required"`
	ScoreRangeMin int                `json:"score_range_min" validate:"required"`
	ScoreRangeMax int                `json:"score_range_max" validate:"required"`
	Priviledges   []string           `json:"priviledges" validate:"required"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
}
