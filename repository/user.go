package repository

import (
	"context"
	bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
)

func GetUserForEmail(ctx context.Context, database *mongo.Database, email string) (dm.User, error) {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	var user dm.User

	err := database.Collection(dm.UserCollectionName).FindOne(queryCtx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return user, errors.Wrap("error loading user for email", err)
	}

	return user, nil
}

func UpdateUserEmailVerificationToken(ctx context.Context, database *mongo.Database, userID dm.UserID, token string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"emailVerificationToken": token}})
	if err != nil {
		return errors.Wrap("cannot update email verification token", err)
	}

	return nil
}

func SetEmailToVerified(ctx context.Context, database *mongo.Database, userID dm.UserID) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"emailVerified": true}})
	if err != nil {
		return errors.Wrap("cannot set email to verified", err)
	}

	return nil
}

func SetNextEmail(ctx context.Context, database *mongo.Database, userID dm.UserID, nextEmail string, verificationToken string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"nextEmail": nextEmail, "emailVerificationToken": verificationToken}})
	if err != nil {
		return errors.Wrap("cannot set next email", err)
	}

	return nil
}

func SetEmailAndClearNextEmail(ctx context.Context, database *mongo.Database, userID dm.UserID, email string) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"email": email, "nextEmail": nil}})
	if err != nil {
		return errors.Wrap("cannot set email and clear next email", err)
	}

	return nil

}

func SetPasswordResetToken(ctx context.Context, database *mongo.Database, userID dm.UserID, token string, validUntil time.Time) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"passwordResetToken": token, "passwordResetTokenValidUntil": validUntil}})
	if err != nil {
		return errors.Wrap("cannot set password reset token", err)
	}

	return nil
}

func SetCredentials(ctx context.Context, database *mongo.Database, userId dm.UserID, credentials dm.UserCredentials) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userId, bson.M{"$set": bson.M{"credentials": credentials}})
	if err != nil {
		return errors.Wrap("cannot set credentials", err)
	}
	return nil
}

func SetLanguage(ctx context.Context, database *mongo.Database, userID dm.UserID, language dm.UserLanguage) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).UpdateByID(queryCtx, userID, bson.M{"$set": bson.M{"language": language}})
	if err != nil {
		return errors.Wrap("cannot set language", err)
	}

	return nil
}

func InsertUser(ctx context.Context, database *mongo.Database, user dm.UserInsert) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.UserCollectionName).InsertOne(queryCtx, dm.User{
		Language:                     user.Language,
		UserName:                     user.UserName,
		Credentials:                  user.Credentials,
		Email:                        user.Email,
		EmailVerified:                user.EmailVerified,
		EmailVerificationToken:       user.EmailVerificationToken,
		NextEmail:                    user.NextEmail,
		PasswordResetToken:           user.PasswordResetToken,
		PasswordResetTokenValidUntil: user.PasswordResetTokenValidUntil,
		SecondFactorToken:            user.SecondFactorToken,
		TemporarySecondFactorToken:   user.TemporarySecondFactorToken,
		UserRoles:                    user.UserRoles,
		Sessions:                     user.Sessions,
		SecondFactorThrottling:       user.SecondFactorThrottling,
	})
	if err != nil {
		return errors.Wrap("cannot insert user", err)
	}
	return nil
}
