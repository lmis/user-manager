package auth

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	dm "user-manager/domain-model"
	"user-manager/util/db"
	"user-manager/util/errs"
)

func UpdateSecondFactorThrottling(ctx context.Context, database *mongo.Database, userID dm.UserID, failedAttemptsSinceLastSuccess int32, maybeTimeoutUntil *time.Time) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"failedAttemptsSinceLastSuccess": failedAttemptsSinceLastSuccess, "timeoutUntil": maybeTimeoutUntil, "updatedAt": time.Now()}})
	if err != nil {
		return errs.Wrap("cannot update second factor throttling", err)
	}
	return nil
}
