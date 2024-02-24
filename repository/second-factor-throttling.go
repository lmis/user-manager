package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
)

func UpdateSecondFactorThrottling(ctx context.Context, database *mongo.Database, userID dm.UserID, failedAttemptsSinceLastSuccess int32, maybeTimeoutUntil *time.Time) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"failedAttemptsSinceLastSuccess": failedAttemptsSinceLastSuccess, "timeoutUntil": maybeTimeoutUntil, "updatedAt": time.Now()}})
	if err != nil {
		return errors.Wrap("cannot update second factor throttling", err)
	}
	return nil
}
