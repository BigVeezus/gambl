package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Calendar Model
type RolesDTO struct {
	ID            primitive.ObjectID `bson:"_id"`
	Branch_id		  string			  `json:"branch_id"`
	Role_id		  string			  `json:"role_id"`
	Role_name	  string			  `json:"role_name"`
	Permissions   []string			  `json:"permissions"`
	Description	  string			  `json:"description"`
	Role_type	  string			  `json:"role_type"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
}

type CreateRolesDTO struct {
	ID            primitive.ObjectID `bson:"_id"`
	Branch_id	  *string		  `json:"branch_id" validate:"required"`
	Role_id		  string			  `json:"role_id"`
	Role_name	  *string			  `json:"name" validate:"required"`
	Permissions   *[]string			  `json:"permissions" validate:"required"`
	Description	  *string			  `json:"description" validate:"required"`
	Role_type	  *string			  `json:"role_type" validate:"required"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
}