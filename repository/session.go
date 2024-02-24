package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func GetUserForSession(ctx context.Context, database *mongo.Database, sessionToken dm.UserSessionToken, sessionType dm.UserSessionType) (dm.User, error) {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	var user dm.User
	sessionMatches := bson.M{"token": sessionToken, "type": sessionType}
	notExpired := bson.M{"timeoutAt": bson.M{"$gt": time.Now()}}

	err := database.Collection(dm.UserCollectionName).FindOne(queryCtx, bson.M{"sessions": bson.M{"$elemMatch": bson.M{"$and": bson.A{sessionMatches, notExpired}}}}).
		Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return dm.User{}, nil
		}
		return dm.User{}, errs.Wrap("error loading user for session", err)
	}
	return user, nil
}

func InsertSession(ctx context.Context, database *mongo.Database, userId dm.UserID, session dm.UserSession) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userId), bson.M{"$push": bson.M{"sessions": session}})

	if err != nil {
		return errs.Wrap("error inserting session", err)
	}

	// Prune old sessions
	queryCtx, cancel = db.DefaultQueryContext(ctx)
	defer cancel()
	_, err = database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userId), bson.M{"$pull": bson.M{"sessions": bson.M{"timeoutAt": bson.M{"$lt": time.Now()}}}})

	if err != nil {
		return errs.Wrap("error pruning old sessions", err)
	}

	return nil
}

func UpdateSessionTimeout(ctx context.Context, database *mongo.Database, sessionToken dm.UserSessionToken, timeout time.Time) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateOne(queryCtx,
		bson.M{"sessions": bson.M{"$elemMatch": bson.M{"token": sessionToken}}},
		bson.M{"$set": bson.M{"sessions.$.timeoutAt": timeout}})

	if err != nil {
		return errs.Wrap("error updating session timeout", err)
	}
	return nil
}

func DeleteSession(ctx context.Context, database *mongo.Database, sessionToken dm.UserSessionToken) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateOne(
		queryCtx,
		bson.M{"sessions": bson.M{"$elemMatch": bson.M{"token": sessionToken}}},
		bson.M{"$pull": bson.M{"sessions": bson.M{"token": sessionToken}}})

	if err != nil {
		return errs.Wrap("error deleting session", err)
	}
	return nil
}
