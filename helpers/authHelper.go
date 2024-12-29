package helper

import (
	"errors"
	"gambl/database"

	// "net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var rolesCollection *mongo.Collection = database.OpenCollection(database.Client, "roles")

// CheckUserType renews the user tokens when they login
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	// println(userType)
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	return err
}

// MatchUserTypeToUid only allows the user to access their data and no other data. Only the admin can access all user data
func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	println(userType)

	if (userType == "USER" || userType != "ADMIN") && uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)

	return err
}
