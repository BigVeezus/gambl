package controllers

import (
	"context"
	"log"
	"strconv"
	"strings"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gambl/database"

	helper "gambl/helpers"
	"gambl/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var validateUser = validator.New()

// CreateUser
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		userName := cleanUsername(*user.User_name)
		user.User_name = &userName

		filter := bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"user_name": user.User_name},
			},
		}

		count, err := userCollection.CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or username already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.User_type = "USER"

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at = user.Created_at
		user.ID = primitive.NewObjectID()
		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, user.User_type, user.ID.Hex())
		user.Refresh_token = &refreshToken

		if err != nil {
			msg := "couldnt generate token"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":          resultInsertionNumber,
			"jwt_token":     string(token),
			"refresh_token": refreshToken,
		})

	}
}

// Create Admin
func SignUpAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		userName := strings.ToLower(*user.User_name)
		user.User_name = &userName

		filter := bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"user_name": user.User_name},
			},
		}

		count, err := userCollection.CountDocuments(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or username already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.User_type = "ADMIN"

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at = user.Created_at
		user.ID = primitive.NewObjectID()
		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, user.User_type, user.ID.Hex())
		user.Refresh_token = &refreshToken

		if err != nil {
			msg := "couldnt generate token"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":          resultInsertionNumber,
			"jwt_token":     string(token),
			"refresh_token": refreshToken,
		})

	}
}

// func ChangePassword() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		var newPasswordPayload models.ChangeUserPassword
// 		var user models.User

// 		if err := c.BindJSON(&newPasswordPayload); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		validationErr := validateUser.Struct(newPasswordPayload)

// 		if validationErr != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}

// 		if *newPasswordPayload.New_password != *newPasswordPayload.Confirm_password {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "new password doesnt match confirm password!"})
// 			return
// 		}

// 		userId, _ := c.Get("uid")
// 		id := userId.(string)

// 		err := userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "user doesnt exist"})
// 			return
// 		}

// 		passwordIsValid, _ := VerifyPassword(*newPasswordPayload.Old_password, *user.Password)
// 		if !passwordIsValid {
// 			c.JSON(http.StatusForbidden, gin.H{"error": "old password incorrect"})
// 			return
// 		}

// 		password := HashPassword(*newPasswordPayload.New_password)

// 		filter := bson.D{{Key: "user_id", Value: id}}

// 		update := bson.D{{Key: "$set", Value: bson.D{
// 			{Key: "password", Value: password},
// 		}}}

// 		_, insertErr := userCollection.UpdateOne(ctx, filter, update)
// 		if insertErr != nil {
// 			msg := "password was not updated"
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"success": true,
// 			"msg":     "password changed",
// 		})
// 	}
// }

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel() // Ensures the context is canceled after execution

		var user models.User
		var foundUser models.User

		// Parse incoming JSON
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate input
		if user.Email == nil || user.Password == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
			return
		}

		// Find user by email
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "login or password is incorrect"})
			return
		}

		// Verify password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusForbidden, gin.H{"error": msg})
			return
		}

		// Generate tokens
		token, refreshToken, err := helper.GenerateAllTokens(*foundUser.Email, foundUser.User_type, foundUser.ID.Hex())
		if err != nil {
			log.Printf("Error generating tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
			return
		}

		// Update refresh token in database
		update := bson.M{"$set": bson.M{"refresh_token": refreshToken}}
		_, err = userCollection.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, update)
		if err != nil {
			log.Printf("Error updating refresh token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update refresh token"})
			return
		}

		foundUser.Password = nil
		foundUser.Refresh_token = nil

		// Send response
		c.JSON(http.StatusOK, gin.H{
			"jwt_token":     token,
			"refresh_token": refreshToken,
			"user":          foundUser,
		})
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// recordPerPage := 10
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		// startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])

	}
}

// GetUser is the api used to tget a single user
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		idLength := len(userId)
		if idLength == 0 {
			err := "no userId found in param!"
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}

func EditUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		var editUser models.User
		var user models.User

		idLength := len(userId)
		if idLength == 0 {
			err := "no userId found in param!"
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.BindJSON(&editUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(editUser)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if editUser.User_name == nil {
			editUser.User_name = user.User_name
		}

		editUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		filter := bson.D{{Key: "user_id", Value: userId}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "user_name", Value: editUser.User_name},
			{Key: "updated_at", Value: editUser.Updated_at},
		}}}

		resultInsertionNumber, insertErr := userCollection.UpdateOne(ctx, filter, update)
		if insertErr != nil {
			msg := "user was not updated"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func UpdateUserType() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		// List of valid user types
		validUserTypes := []string{"ADMIN", "USER", "CAPTAIN", "SUPER_ADMIN"}

		// var editUser models.User
		type User_type struct {
			User_type string `json:"user_type"`
		}
		var user models.User

		// Ensure userId is present in the URL parameter
		idLength := len(userId)
		if idLength == 0 {
			err := "no userId found in param!"
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var userType User_type

		// Parse the incoming JSON body to extract the new user type
		if err := c.BindJSON(&userType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate the user type
		if !contains(validUserTypes, userType.User_type) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user type. Valid types are ADMIN, USER, CAPTAIN, SUPER_ADMIN"})
			return
		}

		userObjId, _ := primitive.ObjectIDFromHex(userId)

		// Find the user from the database
		err := userCollection.FindOne(ctx, bson.M{"_id": userObjId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		// Update the user type and update timestamp
		user.Updated_at = time.Now()

		filter := bson.D{{Key: "_id", Value: userObjId}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "user_type", Value: userType.User_type}, // Update user_type
			{Key: "updated_at", Value: user.Updated_at},   // Update updated_at
		}}}

		// Perform the update in the database
		_, insertErr := userCollection.UpdateOne(ctx, filter, update)
		if insertErr != nil {
			msg := "user type was not updated"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}
