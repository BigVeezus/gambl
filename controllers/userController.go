package controllers

import (
	"context"
	"log"
	"strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gambl/database"

	config "gambl/config"
	helper "gambl/helpers"
	"gambl/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var validateUser = validator.New()

// HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or passowrd is incorrect"
		check = false
	}

	return check, msg
}

// CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		var user models.SignUpUser

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, _ := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		user.Status = "INACTIVE"
		token, _, err := helper.GenerateAllTokens(*user.Email, "UNBOARDED", user.User_id)

		if err != nil {
			msg := "couldnt generate token"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		user.OTP = config.GenerateOTP(4)

		config.SendOTPMail(*user.Email, user.OTP)

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":      resultInsertionNumber,
			"jwt_token": string(token)})

	}
}

func ValidateOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var otp models.Otp
		var user models.User

		if err := c.BindJSON(&otp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(otp)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		userId, _ := c.Get("uid")
		id := userId.(string)

		err := userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user doesnt exist"})
			return
		}

		if user.OTP != *otp.OTP {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OTP!"})
			return
		}

		filter := bson.D{{Key: "user_id", Value: id}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "otpVerified", Value: true},
		}}}

		_, insertErr := userCollection.UpdateOne(ctx, filter, update)
		if insertErr != nil {
			msg := "otp was not updated"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OTP validated",
		})
	}
}

func ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var newPasswordPayload models.ChangeUserPassword
		var user models.User

		if err := c.BindJSON(&newPasswordPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(newPasswordPayload)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if *newPasswordPayload.New_password != *newPasswordPayload.Confirm_password {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new password doesnt match confirm password!"})
			return
		}

		userId, _ := c.Get("uid")
		id := userId.(string)

		err := userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user doesnt exist"})
			return
		}

		passwordIsValid, _ := VerifyPassword(*newPasswordPayload.Old_password, *user.Password)
		if !passwordIsValid {
			c.JSON(http.StatusForbidden, gin.H{"error": "old password incorrect"})
			return
		}

		password := HashPassword(*newPasswordPayload.New_password)

		filter := bson.D{{Key: "user_id", Value: id}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "password", Value: password},
		}}}

		_, insertErr := userCollection.UpdateOne(ctx, filter, update)
		if insertErr != nil {
			msg := "password was not updated"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "password changed",
		})
	}
}

func ResendOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var resendOtp models.ResendOtp
		var user models.User

		if err := c.BindJSON(&resendOtp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateUser.Struct(resendOtp)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": resendOtp.Email}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user doesnt exist"})
			return
		}

		u_type := user.User_type
		otp := config.GenerateOTP(4)

		if *u_type != "UNBOARDED" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user has been onboarded"})
			return
		}

		token, _, err := helper.GenerateAllTokens(*user.Email, "UNBOARDED", user.User_id)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not generate token"})
			return
		}

		filter := bson.D{{Key: "email", Value: resendOtp.Email}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "otp", Value: otp},
		}}}

		_, insertErr := userCollection.UpdateOne(ctx, filter, update)
		if insertErr != nil {
			msg := "otp was not sent"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"msg":   "OTP sent",
			"token": token,
		})
	}
}

func TestOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var _, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config.SendOTPMail("elvis.osujic@gmail.com", "4000")

		c.JSON(http.StatusOK, gin.H{
			"msg": "OTP sent",
		})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "login or passowrd is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusForbidden, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		token, _, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.User_type, foundUser.User_id)

		c.JSON(http.StatusOK, gin.H{
			"jwt_token": string(token),
			"user":      foundUser,
		})

	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.RoleTypeCheck(c, "ADMIN", "65f2e8b6ab423f0349550c4f", "staff_indicators_charts:readk"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

		var editUser models.EditUser
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

		if editUser.First_name == nil {
			editUser.First_name = user.First_name
		}
		if editUser.Address == nil {
			editUser.Address = user.Address
		}
		if editUser.Last_name == nil {
			editUser.Last_name = user.Last_name
		}
		if editUser.Phone == nil {
			editUser.Phone = &user.Phone
		}
		if editUser.Role == nil {
			editUser.Role = &user.Role
		}

		editUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		filter := bson.D{{Key: "user_id", Value: userId}}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "first_name", Value: editUser.First_name},
			{Key: "last_name", Value: editUser.Last_name},
			{Key: "address", Value: editUser.Address},
			{Key: "phone", Value: editUser.Phone},
			{Key: "role", Value: editUser.Role},
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
