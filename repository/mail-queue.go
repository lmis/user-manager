package repository

import (
	"database/sql"
	"user-manager/db"
	. "user-manager/db/generated/models/postgres/public/enum"
	. "user-manager/db/generated/models/postgres/public/table"
	dm "user-manager/domain-model"
	"user-manager/util/errors"
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
	priority dm.MailQueuePriority,
) error {
	err := db.ExecSingleMutation(
		MailQueue.INSERT(
			MailQueue.FromAddress,
			MailQueue.ToAddress,
			MailQueue.Content,
			MailQueue.Subject,
			MailQueue.Status,
			MailQueue.Priority).
			VALUES(from, to, content, subject, EmailStatus.Pending, int16(priority)).
			ExecContext,
		r.tx)
	if err != nil {
		return errors.Wrap("issue inserting email in db", err)
	}

	return nil
}
