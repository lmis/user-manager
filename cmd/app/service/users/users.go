package users

import (
	"context"
	"errors"
	bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	dm "user-manager/domain-model"
	"user-manager/util/db"
	"user-manager/util/errs"
)

func GetUserForEmail(ctx context.Context, database *mongo.Database, email string) (dm.User, error) {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	var user dm.User

	err := database.Collection(dm.UserCollectionName).FindOne(queryCtx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, nil
		}
		return user, errs.Wrap("error loading user for email", err)
	}

	return user, nil
}

func UpdateUserEmailVerificationToken(ctx context.Context, database *mongo.Database, userID dm.UserID, token string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"emailVerificationToken": token}})
	if err != nil {
		return errs.Wrap("cannot update email verification token", err)
	}

	return nil
}

func SetEmailToVerified(ctx context.Context, database *mongo.Database, userID dm.UserID) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"emailVerified": true}})
	if err != nil {
		return errs.Wrap("cannot set email to verified", err)
	}

	return nil
}

func SetNextEmail(ctx context.Context, database *mongo.Database, userID dm.UserID, nextEmail string, verificationToken string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"nextEmail": nextEmail, "emailVerificationToken": verificationToken}})
	if err != nil {
		return errs.Wrap("cannot set next email", err)
	}

	return nil
}

func SetEmailAndClearNextEmail(ctx context.Context, database *mongo.Database, userID dm.UserID, email string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"email": email, "nextEmail": nil}})
	if err != nil {
		return errs.Wrap("cannot set email and clear next email", err)
	}

	return nil

}

func SetPasswordResetToken(ctx context.Context, database *mongo.Database, userID dm.UserID, token string, validUntil time.Time) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userID), bson.M{"$set": bson.M{"passwordResetToken": token, "passwordResetTokenValidUntil": validUntil}})
	if err != nil {
		return errs.Wrap("cannot set password reset token", err)
	}

	return nil
}

func SetCredentials(ctx context.Context, database *mongo.Database, userId dm.UserID, credentials dm.UserCredentials) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, primitive.ObjectID(userId), bson.M{"$set": bson.M{"credentials": credentials}})
	if err != nil {
		return errs.Wrap("cannot set credentials", err)
	}
	return nil
}

func InsertUser(ctx context.Context, database *mongo.Database, user dm.UserInsert) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).InsertOne(queryCtx, dm.User{
		Name:                   user.UserName,
		Credentials:            user.Credentials,
		Email:                  user.Email,
		EmailVerified:          user.EmailVerified,
		EmailVerificationToken: user.EmailVerificationToken,
		UserRoles:              user.UserRoles,
		Sessions:               []dm.UserSession{},
	})
	if err != nil {
		return errs.Wrap("cannot insert user", err)
	}
	return nil
}
