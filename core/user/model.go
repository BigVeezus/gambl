// core/user/model.go
package user

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserType string

const (
	NormalUser UserType = "USER"
	Admin      UserType = "ADMIN"
	SuperAdmin UserType = "SUPER_ADMIN"
	Captain    UserType = "CAPTAIN"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	User_name     string             `json:"user_name" validate:"required"`
	Password      string             `json:"password" validate:"required,min=6"`
	Email         string             `json:"email" validate:"email,required"`
	Country       string             `json:"country"`
	User_type     UserType           `json:"user_type"`
	Score         int                `json:"score"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
}

// Helper function to validate user type
func isValidUserType(userType UserType) bool {
	validUserTypes := []UserType{
		NormalUser,
		Admin,
		SuperAdmin,
		Captain,
	}

	for _, validType := range validUserTypes {
		if userType == validType {
			return true
		}
	}
	return false
}

var (
	ErrScore           = errors.New("score must be greater than 0")
	ErrInvalidUserType = errors.New("invalid user type")
)
