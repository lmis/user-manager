package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"user-manager/db"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func InsertPendingMail(
	ctx context.Context,
	database *mongo.Database,
	mail dm.MailInsert,
) error {
	queryCtx, cancel := db.DefaultQueryContext(ctx)
	defer cancel()

	_, err := database.Collection(dm.MailQueueCollectionName).InsertOne(queryCtx, dm.Mail{
		From:     mail.From,
		To:       mail.To,
		Subject:  mail.Subject,
		Content:  mail.Content,
		Priority: mail.Priority,
		Status:   dm.MailStatusPending,
	})

	if err != nil {
		return errs.Wrap("issue inserting email in db", err)
	}

	return nil
}
