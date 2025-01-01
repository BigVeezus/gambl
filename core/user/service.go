// core/user/service.go
package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	Collection *mongo.Collection
}

func NewUserService(collection *mongo.Collection) *UserService {
	return &UserService{Collection: collection}
}

// Create a new user
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	var existingUser User

	filter := bson.M{
		"$or": []bson.M{
			{"email": user.Email},
			{"user_name": user.User_name},
		},
	}

	projection := bson.M{
		"_id":       1,
		"email":     1,
		"user_name": 1,
	}

	// Perform the query with the filter and projection, directly handle error
	err := s.Collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&existingUser)
	if err == nil && existingUser.ID != primitive.NilObjectID {
		// Early return if user exists, minimizing further logic execution
		return errors.New("email or username already exists")
	} else if err != nil && err != mongo.ErrNoDocuments {
		// Handle other errors
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	// _ = s.Collection.FindOne(ctx, filter).Decode(&existingUser)

	// if existingUser.ID != primitive.NilObjectID {
	// 	return errors.New("email or username already exists")
	// }

	user.Score = 1
	user.User_type = NormalUser

	// Hash the password before saving it
	user.Password = HashPassword(user.Password)

	user.User_name = CleanUsername(user.User_name)

	user.ID = primitive.NewObjectID()
	user.Created_at = time.Now()
	user.Updated_at = user.Created_at

	_, err = s.Collection.InsertOne(ctx, user)

	return err
}

// Edit a user's details
func (s *UserService) EditUser(ctx context.Context, userID primitive.ObjectID, update bson.M) error {
	// Ensure no invalid fields are being updated (optional validation)
	if _, ok := update["user_type"]; ok {
		if !isValidUserType(UserType(update["user_type"].(string))) {
			return ErrInvalidUserType
		}
	}
	update["updated_at"] = time.Now()

	_, err := s.Collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)

	return err
}

// Change the user's status (User_type)
func (s *UserService) ChangeUserStatus(ctx context.Context, userID primitive.ObjectID, userType UserType) error {
	// Validate the new user type
	if !isValidUserType(userType) {
		return ErrInvalidUserType
	}

	_, err := s.Collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"user_type": userType, "updated_at": time.Now()}},
	)
	return err
}

// GetUserById fetches a user by their ID
func (s *UserService) GetUserById(ctx context.Context, userID string) (*User, error) {
	// Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Query the database for the user
	var user User
	err = s.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserById fetches a user by their ID
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// Query the database for the user
	var user User
	err := s.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates the user details
func (s *UserService) UpdateUser(ctx context.Context, userID string, updates *User) error {
	// Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Create the update document
	updateDoc := bson.M{"$set": updates}

	// Perform the update
	result, err := s.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetAllUsers fetches all users from the database
func (s *UserService) GetAllUsers(ctx context.Context) ([]User, error) {
	// Find all users
	cursor, err := s.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the users
	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserType updates the user type for a specific user
func (s *UserService) UpdateUserType(ctx context.Context, userID string, userType string) error {
	// Convert string ID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Create the update document
	updateDoc := bson.M{"$set": bson.M{"user_type": userType}}

	// Perform the update
	result, err := s.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

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

func CleanUsername(username string) string {
	// Remove everything that is not an alphabet or a number
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	cleanedUsername := re.ReplaceAllString(username, "")

	// Optionally, convert to lowercase
	return strings.ToLower(cleanedUsername)
}
