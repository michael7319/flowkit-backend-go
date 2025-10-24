package utils

import (
	"context"
	"fmt"

	"github.com/flowkit/backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if password matches hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateStaffID generates a unique staff ID
func GenerateStaffID(ctx context.Context) (string, error) {
	count, err := config.UsersCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return "", err
	}

	staffID := fmt.Sprintf("%04d", count+1)
	return staffID, nil
}

// GetNextSequence gets next sequence value for counters
func GetNextSequence(ctx context.Context, name string) (int64, error) {
	counters := config.DB.Collection("counters")

	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int64 `bson:"seq"`
	}

	err := counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.Seq, nil
}
