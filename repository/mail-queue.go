package repository

import (
	"database/sql"
	"user-manager/db"
	"user-manager/db/generated/models"
	domain_model "user-manager/domain-model"
	"user-manager/util"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type MailQueueRepository struct {
	tx *sql.Tx
}

func ProvideMailQueueRepository(tx *sql.Tx) *MailQueueRepository {
	return &MailQueueRepository{tx}
}

func (r *MailQueueRepository) InsertPending(
	from string,
	to string,
	content string,
	subject string,
	priority domain_model.MailQueuePriority,
) error {
	mail := models.MailQueue{
		FromAddress: from,
		ToAddress:   to,
		Content:     content,
		Subject:     subject,
		Status:      models.EmailStatusPENDING,
		Priority:    int16(priority),
	}
	ctx, cancel := db.DefaultQueryContext()
	defer cancel()
	if err := mail.Insert(ctx, r.tx, boil.Infer()); err != nil {
		return util.Wrap("issue inserting email in db", err)
	}

	return nil
}
