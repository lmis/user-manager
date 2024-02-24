package domain_model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MailQueueID primitive.ObjectID
type MailQueuePriority int16
type MailStatus string

const (
	MailQueuePrioLow  MailQueuePriority = 10
	MailQueuePrioMid  MailQueuePriority = 100
	MailQueuePrioHigh MailQueuePriority = 1000
	MailStatusPending MailStatus        = "pending"
	MailStatusSent    MailStatus        = "sent"
	MailStatusFailed  MailStatus        = "failed"

	MailQueueCollectionName = "mailQueue"
)

type Mail struct {
	ObjectID       primitive.ObjectID `bson:"_id,omitempty"`
	From           string             `bson:"from,omitempty"`
	To             string             `bson:"to,omitempty"`
	Content        string             `bson:"content,omitempty"`
	Subject        string             `bson:"subject,omitempty"`
	Status         MailStatus         `bson:"status,omitempty"`
	Priority       MailQueuePriority  `bson:"priority,omitempty"`
	FailedAttempts int8               `bson:"failedAttempts,omitempty"`
	UpdatedAt      time.Time          `bson:"updatedAt,omitempty"`
}

func (m Mail) ID() MailQueueID {
	return MailQueueID(m.ObjectID)
}
func (m Mail) IsPresent() bool {
	return m.ObjectID != primitive.NilObjectID
}

type MailInsert struct {
	From     string
	To       string
	Content  string
	Subject  string
	Priority MailQueuePriority
}
