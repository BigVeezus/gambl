// core/user/repository.go
package user

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{Collection: collection}
}

// Create a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
	user.ID = primitive.NewObjectID()
	user.Created_at = time.Now()
	user.Updated_at = time.Now()

	_, err := r.Collection.InsertOne(ctx, user)
	return err
}

// Edit an existing user's details
func (r *UserRepository) EditUser(ctx context.Context, userID primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)
	return err
}

// Change the user's status (User_type)
func (r *UserRepository) ChangeUserStatus(ctx context.Context, userID primitive.ObjectID, userType UserType) error {
	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"user_type": userType, "updated_at": time.Now()}},
	)
	return err
}
